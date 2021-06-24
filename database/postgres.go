package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/lib/pq"
)

type Config struct {
	User, Password, Host, Port, DatabaseName, MigrationDir string
	SSLDisabled                                            bool
}

type Database struct {
	client            *sqlx.DB
	migrationDir      string
	sqlBuilderDialect goqu.DialectWrapper
	sqlSmartExec      *goqu.Database
}

// DB is an alias to Database (less to type out).
type DB = Database

// Dial connects to a postgres database using the provided configuration,
// and creates/updates the migration table `db_version` with the current version.
func Dial(conf Config) (*Database, error) {
	sslMode := "require"
	if conf.SSLDisabled {
		sslMode = "disable"
	}

	url := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DatabaseName,
		sslMode,
	)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS db_version (
			id VARCHAR(1),
			version bigint,
			hash VARCHAR(256),
			file VARCHAR(256),
			last_run TIMESTAMPTZ,
			complete BOOLEAN
		);
	`)
	if err != nil {
		return nil, err
	}

	d := &Database{
		client:            sqlx.NewDb(db, "postgres"),
		sqlBuilderDialect: goqu.Dialect("postgres"),
		sqlSmartExec:      goqu.New("postgres", db),
		migrationDir:      conf.MigrationDir,
	}
	_, err = d.GetCurrentMigration()
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := db.Exec(`
				INSERT INTO db_version
				(id, version, hash, file, last_run, complete) VALUES
				('1', 0, '', '', NOW(), true);
			`)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return d, nil
}

func (d *Database) Build() goqu.DialectWrapper {
	return d.sqlBuilderDialect
}

func (d *Database) Exec() *goqu.Database {
	return d.sqlSmartExec
}

// Ping sends a ping message to the database to check for signs of life.
func (d *Database) Ping() error {
	return d.client.Ping()
}

// Close gracefully closes the connection to the database.
func (d *Database) Close() error {
	return d.client.Close()
}

// exec provides the underlying functionality for the db.View and db.Update transaction handling methods.
func (d *Database) exec(ctx context.Context, callback func(*sqlx.Tx) error, readOnly bool) error {
	tx, err := d.client.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: readOnly})
	if err != nil {
		return err
	}

	err = callback(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// View creates a read-only database transaction around the provided
// callback to manage handling the transaction auto-magically.
func (d *Database) View(ctx context.Context, callback func(*sqlx.Tx) error) error {
	return d.exec(ctx, callback, true)
}

// Update creates a read-write database transaction around the provided
// callback to manage handling the transaction auto-magically.
func (d *Database) Update(ctx context.Context, callback func(*sqlx.Tx) error) error {
	return d.exec(ctx, callback, false)
}

// ExecFile parses the SQL blocks within a file and executes them independently
// from first to last.
func (d *Database) ExecFile(filepath string) error {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	// split on the semicolon delimiter
	blocks := strings.Split(string(bytes), ";")

	return d.Update(context.Background(), func(tx *sqlx.Tx) error {
		for i, block := range blocks {
			_, err := tx.Exec(block)
			if err != nil {
				return fmt.Errorf("failed to execute block %d of sql file: %s", i, err.Error())
			}
		}

		return nil
	})
}
