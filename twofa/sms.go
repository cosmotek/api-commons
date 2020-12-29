package twofa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cosmotek/nexgo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/cosmotek/api-commons/database"
)

var (
	// ErrSMSPincodeExpires is an error returned by the verify function
	// if a pincode has expired.
	ErrSMSPincodeExpired = errors.New("sms pincode has expired")
)

// smsPincode is used internally to retreive pincode data from the database.
type smsPincode struct {
	RequestID   string    `json:"requestId" db:"id"`
	PhoneNumber string    `json:"phoneNumber" db:"phone_number"`
	Pincode     string    `json:"pincode" db:"pincode"`
	ExpiresAt   time.Time `json:"expiresAt" db:"expires_at"`
}

const smsPincodeMessageTemplate = `Your verification code is %s`

// SendSMSPincode sends a pincode verification text using the template string
// included internally in this package and the provided pincode string.
func SendSMSPincode(messenger nexgo.Messenger, phoneNumber, pincode string) error {
	return messenger.Send("Verification Code", phoneNumber, fmt.Sprintf(smsPincodeMessageTemplate, pincode))
}

// CreateSMSPincode creates a pincode and pincode record, writing the record
// to the database, and returning the pincode, as well as recordId to the
// caller (ordered id, pincode, error).
func CreateSMSPincode(db *database.DB, phoneNumber string) (string, string, error) {
	queryStr := `
	INSERT INTO sms_pincodes (
		id,
		phone_number,
		pincode,
		expires_at
	) VALUES (
		:id,
		:phone_number,
		:pincode,
		:expires_at
	);
	`

	id := uuid.New().String()
	code := GenerateNumericPincode(4)

	pin := smsPincode{
		RequestID:   id,
		PhoneNumber: phoneNumber,
		Pincode:     code,
		ExpiresAt:   time.Now().UTC().Add(time.Minute * 30),
	}

	err := db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(queryStr, pin)
		return err
	})

	return id, code, err
}

// VerifySMSPincode searches for records in the database with the matching
// pincode and phoneNumber, checks if the record has expired and returns the
// requestId or ErrSMSPincodeExpired.
func VerifySMSPincode(db *database.DB, phoneNumber, pincode string) (string, error) {
	queryStr := `
	SELECT p.expires_at, p.id FROM sms_pincodes p 
	WHERE
		p.phone_number = $1 AND
		p.pincode = $2
	LIMIT 1;
	`

	pin := smsPincode{}
	err := db.View(context.Background(), func(tx *sqlx.Tx) error {
		return tx.Get(&pin, queryStr, phoneNumber, pincode)
	})
	if err != nil {
		return "", err
	}

	if pin.ExpiresAt.Before(time.Now().UTC()) {
		return "", ErrSMSPincodeExpired
	}

	return pin.RequestID, nil
}

// DeleteSMSPincode removes a pincode record from the database.
// This function should be used when a pincode has been successfully
// verified in order to prevent double-booking pincodes.
func DeleteSMSPincode(db *database.DB, id string) error {
	queryStr := `DELETE FROM sms_pincodes s WHERE s.id = $1`
	return db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.Exec(queryStr, id)
		return err
	})
}
