package database

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	ID       string    `db:"id"`
	File     string    `db:"file"`
	Hash     string    `db:"hash"`
	Version  uint64    `db:"version"`
	Complete bool      `db:"complete"`
	LastRun  time.Time `db:"last_run"`
}

type MigrationStatus struct {
	Applied uint64
	Failed  uint64
	Skipped uint64
	Latest  uint64
}

func (d *Database) SyncMigrations() (MigrationStatus, error) {
	currentMig, err := d.GetCurrentMigration()
	if err != nil {
		return MigrationStatus{}, err
	}

	diffMigs, err := d.DiffMigrations(false)
	if err != nil {
		return MigrationStatus{}, err
	}

	return d.RunMigrations(currentMig, diffMigs...)
}

func (d *Database) GetCurrentMigration() (Migration, error) {
	migration := Migration{}
	err := d.View(context.Background(), func(tx *sqlx.Tx) error {
		err := tx.Get(&migration, "SELECT * FROM db_version WHERE id = '1' LIMIT 1")
		if err != nil {
			if err == sql.ErrNoRows {
				return err
			}

			return fmt.Errorf("failed to fetch current migration status: %s", err.Error())
		}

		return nil
	})

	return migration, err
}

func (d *Database) DiffMigrations(compareHashes bool) ([]Migration, error) {
	currentMigration, err := d.GetCurrentMigration()
	if err != nil {
		return nil, err
	}

	if !currentMigration.Complete {
		return nil, fmt.Errorf(
			"migration %d in file %s appears to have failed, please rectify manually",
			currentMigration.Version, currentMigration.File,
		)
	}

	migrations := make([]Migration, 0)
	err = filepath.Walk(d.migrationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".sql" {
			return nil
		}

		version, err := strconv.ParseInt(strings.Replace(info.Name(), ".sql", "", -1), 10, 64)
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		migration := Migration{
			File:     path,
			Hash:     fmt.Sprintf("%x", md5.Sum(bytes)),
			Version:  uint64(version),
			Complete: uint64(version) <= currentMigration.Version,
		}

		if compareHashes && migration.Version == currentMigration.Version && migration.Hash != currentMigration.Hash {
			return errors.New("migrations are up to date but appear to have been modified (latest hash mismatch)")
		}

		migrations = append(migrations, migration)
		return nil
	})

	return migrations, err
}

func (d *Database) RunMigrations(currentMigration Migration, migrations ...Migration) (MigrationStatus, error) {
	sort.Sort(MigrationSet(migrations))
	migrationStatus := MigrationStatus{
		Latest: currentMigration.Version,
	}

	for _, migration := range migrations {
		if migration.Complete {
			// if a migration is already complete just skip it
			migrationStatus.Skipped += 1
		} else {
			err := d.Update(context.Background(), func(tx *sqlx.Tx) error {
				_, err := tx.Exec(
					"UPDATE db_version SET version = $1, hash = $2, file = $3, last_run = $4, complete = $5 WHERE id = '1'",
					migration.Version, migration.Hash, migration.File, time.Now(), false,
				)
				if err != nil {
					return fmt.Errorf("failed to open migration step: %s", err.Error())
				}

				return nil
			})
			if err != nil {
				migrationStatus.Failed += 1

				return migrationStatus, err
			}

			err = d.ExecFile(migration.File)
			if err != nil {
				migrationStatus.Failed += 1

				return migrationStatus, err
			}

			err = d.Update(context.Background(), func(tx *sqlx.Tx) error {
				_, err := tx.Exec("UPDATE db_version SET complete = $1 WHERE id = '1' AND version = $2", true, migration.Version)
				return err
			})
			if err != nil {
				migrationStatus.Failed += 1

				return migrationStatus, err
			}

			migrationStatus.Applied += 1
			migrationStatus.Latest = migration.Version

		}
	}

	return migrationStatus, nil
}
