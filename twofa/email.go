package twofa

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/cosmotek/mailgo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/cosmotek/api-commons/database"
)

// magicLinkURLTemplate is used for creating magic links that
// will be emailed to users and will open the admin dashboard.
// (DO NOT EDIT WITHOUT CORRESPONDING UI CHANGES)
const magicLinkURLTemplate = "%s/#/handoffs/magiclinks/%s/%s/avp/%s"

var (
	// ErrMagicLinkExpired is an error returned when someone attempts
	// to login (verify) with a expired magic link.
	ErrMagicLinkExpired = errors.New("magic link has expired")

	// ErrMagicLinkConsumed is an error returned when someone attempts
	// to login (verify) with a consumed magic link.
	ErrMagicLinkConsumed = errors.New("magic link has been consumed")
)

// SendHTMLEmail generates and sends an email
// using the provided HTML string, returning the sender
// email (with formatted name included).
func SendHTMLEmail(messenger mailgo.Messenger, emailHTML, subject, recipientEmail string) (string, error) {
	sender := messenger.GenerateSender("_____", "noreply")
	return string(sender), messenger.SendHTML(
		subject,
		recipientEmail,
		emailHTML,
		sender,
	)
}

// magicLink is used for creating magic links as well as retrieving them
// from the database (used as a destination).
type magicLink struct {
	RequestID        string    `json:"requestId" db:"id"`
	Email            string    `json:"email" db:"email"`
	VerificationCode string    `json:"verificationCode" db:"verification_code"`
	ExpiresAt        time.Time `json:"expiresAt" db:"expires_at"`
	Consumed         bool      `json:"consumed" db:"consumed"`
}

// generateMagicLinkURL creates a magic link url using the template,
// host url, and the magic link input (which includes metadata required).
func generateMagicLinkURL(hostURL string, input magicLink) string {
	return fmt.Sprintf(
		magicLinkURLTemplate,
		hostURL,
		base64.RawURLEncoding.EncodeToString([]byte(input.Email)),
		input.RequestID,
		input.VerificationCode,
	)
}

// VerifyMagicLink searches the database for a magic link with the
// matching requestId, email and verification code. This functional
// also checks if the link has expired and returns ErrMagicLinkExpired
// if that is the case.
func VerifyMagicLink(db *database.DB, requestID, email, verificationCode string) error {
	queryStr := `
	SELECT m.expires_at FROM email_magiclinks m 
	WHERE
		m.id = $1 AND
		m.email = $2 AND
		m.verification_code = $3
	LIMIT 1;
	`

	mlink := magicLink{}
	err := db.View(context.Background(), func(tx *sqlx.Tx) error {
		return tx.Get(&mlink, queryStr, requestID, email, verificationCode)
	})
	if err != nil {
		return err
	}

	if mlink.Consumed {
		return ErrMagicLinkConsumed
	}

	if mlink.ExpiresAt.Before(time.Now().UTC()) {
		return ErrMagicLinkExpired
	}

	return nil
}

// ConsumeMagicLink marks a link as "consumed" in the database. This function should
// be called when a magic link has been verified and a token has been issued.
func ConsumeMagicLink(db *database.DB, requestID, email, verificationCode string) error {
	queryStr := `
	UPDATE
		email_magiclinks m
	SET
		consumed = TRUE
	WHERE
		m.id = $1 AND
		m.email = $2 AND
		m.verification_code = $3;
	`

	err := db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.Exec(queryStr, requestID, email, verificationCode)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateMagicLink creates the database records later used for magic link verification
// and returns a generated url which may be emailed to a user.
func CreateMagicLink(db *database.DB, frontendURL, email string) (string, string, error) {
	queryStr := `
	INSERT INTO email_magiclinks (
		id,
		email,
		verification_code,
		expires_at
	) VALUES (
		:id,
		:email,
		:verification_code,
		:expires_at
	);
	`

	id := uuid.New().String()
	mlink := magicLink{
		RequestID:        id,
		Email:            email,
		VerificationCode: GenerateAlphaNumericCode(36),
		ExpiresAt:        time.Now().UTC().Add(time.Hour * 48),
	}

	err := db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(queryStr, mlink)
		return err
	})

	return id, generateMagicLinkURL(frontendURL, mlink), err
}
