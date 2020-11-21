package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	jwt "gopkg.in/dgrijalva/jwt-go.v2"

	"gitlab.com/fivepm/api-commons/database"
)

var (
	// ErrTokenExpired is returned by the Verify function
	// and is used when attempted to refresh tokens.
	// If ErrTokenExpired is returned then everything else
	// on the token should be valid, and the token should
	// be allowed to refresh/renew.
	ErrTokenExpired  = errors.New("user error: failed to authenticate, token has expired")
	ErrTokenOutdated = errors.New("user error: failed to authenticate, token is outdated")
)

// Role is a system identifier used to determine permissions that
// an actor has within the application.
type Role string

// SubjectType is type (and set of constants) used to identify
// the type of actor/user identifier included in a AuthToken.
type SubjectType string

const (
	// AdminUser is a user logged-in via Email, who has access to everything
	// in the system.
	AdminUser Role = "admin_user"

	// OrgUser is a user logged-in via Email, and only has access to public
	// data as well as data owned by the user's org.
	OrgUser Role = "org_user"

	// User is a user logged-in via SMS, and only has access to public
	// and personally owned data.
	Vuer Role = "vuer"

	// ServiceUser is a API/Machine consumer, and has the same rights
	// as a OrgUser.
	ServiceUser Role = "service_user"

	PhoneNumber SubjectType = "phone_number"
	APIKey      SubjectType = "api_key"
	Email       SubjectType = "email"
)

// AuthToken is a object used to create/carry JWT claims
// in tokens, request context etc.
type AuthToken struct {
	ID           string      `json:"jti" mapstructure:"jti" db:"jti"`
	TokenVersion int         `json:"tver" mapstructure:"tver" db:"token_version"`
	Role         Role        `json:"role" mapstructure:"role" db:"role"`
	Audience     string      `json:"aud" mapstructure:"aud" db:"audience"`
	Subject      string      `json:"sub,omitempty" mapstructure:"sub" db:"subject"`
	SubjectType  SubjectType `json:"subtyp,omitempty" mapstructure:"subtyp" db:"subject_type"`

	IssuedAt  time.Time `json:"iat" mapstructure:"iat" db:"issued_at"`
	NotBefore time.Time `json:"nbf" mapstructure:"nbf" db:"not_before"`
	ExpiresAt time.Time `json:"exp" mapstructure:"exp" db:"expires_at"`

	RefreshCount uint64 `json:"-" mapstructure:"-" db:"refresh_count"`
	Deactivated  bool   `json:"-" mapstructure:"-" db:"deactivated"`
}

// Verify takes in a JWT token string and rsaKey []byte and validates it.
// If the token is invalid in any way other than expiry, a custom error will
// be returned. If the token is expired, ErrTokenExpired will be returned.
// If the token is valid, an AuthToken with the JWT token claims will be
// returned.
func Verify(inputStr string, rsaKey *rsa.PublicKey) (AuthToken, error) {
	token, err := jwt.Parse(inputStr, func(token *jwt.Token) (interface{}, error) {
		return rsaKey, nil
	})

	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok {
			return AuthToken{}, fmt.Errorf("failed to decode, token invalid: %v", err)
		}

		if validationErr.Errors&jwt.ValidationErrorMalformed != 0 {
			return AuthToken{}, errors.New("failed to decode, token malformed")
		} else if validationErr.Errors&jwt.ValidationErrorNotValidYet != 0 {
			return AuthToken{}, errors.New("failed to decode, token not active")
		} else if validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
			return AuthToken{}, errors.New("failed to decode, token signature invalid")
		}
	}

	var t AuthToken
	err = decode(token.Claims, &t)
	if err != nil {
		return AuthToken{}, err
	}

	if t.ExpiresAt.Before(time.Now().UTC()) {
		return AuthToken{}, ErrTokenExpired
	}

	if t.TokenVersion < 1 {
		return AuthToken{}, ErrTokenOutdated
	}

	return t, nil
}

// VerifyIgnoringExpiry parses and returns the AuthToken (claims) for an expired token,
// but only if the token is valid in all other ways.
func VerifyIgnoringExpiry(inputStr string, rsaKey *rsa.PublicKey) (AuthToken, error) {
	token, err := jwt.Parse(inputStr, func(token *jwt.Token) (interface{}, error) {
		return rsaKey, nil
	})

	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok {
			return AuthToken{}, fmt.Errorf("failed to decode, token invalid: %v", err)
		}

		if validationErr.Errors&jwt.ValidationErrorMalformed != 0 {
			return AuthToken{}, errors.New("failed to decode, token malformed")
		} else if validationErr.Errors&jwt.ValidationErrorNotValidYet != 0 {
			return AuthToken{}, errors.New("failed to decode, token not active")
		} else if validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
			return AuthToken{}, errors.New("failed to decode, token signature invalid")
		}
	}

	var t AuthToken
	err = decode(token.Claims, &t)
	if err != nil {
		return AuthToken{}, err
	}

	if t.TokenVersion < 1 {
		return AuthToken{}, ErrTokenOutdated
	}

	return t, nil
}

// Sign creates and signs a JWT token using the provided AuthToken and rsaKey []byte slice,
// returning the JWT token string or an error.
func Sign(authToken AuthToken, rsaKey *rsa.PrivateKey) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = map[string]interface{}{
		"jti":    authToken.ID,
		"role":   authToken.Role,
		"aud":    authToken.Audience,
		"sub":    authToken.Subject,
		"subtyp": authToken.SubjectType,
		"iat":    authToken.IssuedAt,
		"nbf":    authToken.NotBefore,
		"exp":    authToken.ExpiresAt,
		"tver":   authToken.TokenVersion,
	}

	return token.SignedString(rsaKey)
}

func RecordTokenCreate(db *database.DB, authToken AuthToken) error {
	query := `
INSERT INTO jwt_tokens (
	jti,
	token_version,
	role,
	audience,
	subject,
	subject_type,
	issued_at,
	not_before,
	expires_at,
	refresh_count,
	deactivated
) VALUES (
	:jti,
	:token_version,
	:role,
	:audience,
	:subject,
	:subject_type,
	:issued_at,
	:not_before,
	:expires_at,
	:refresh_count,
	:deactivated
);
	`

	return db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(query, authToken)
		return err
	})
}

func RecordTokenRefresh(db *database.DB, id string, newExp time.Time) error {
	return db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE jwt_tokens SET refresh_count = refresh_count + 1, expires_at = $2 WHERE jti = $1`, id, newExp)
		return err
	})
}

func RecordTokenActivationStatusChange(db *database.DB, subject string, active bool) error {
	return db.Update(context.Background(), func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE jwt_tokens SET deactivated = $2 WHERE subject = $1`, subject, !active)
		return err
	})
}

// toTimeHookFunc provides a time decoder for mapstructure decode calls
func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

// decode wraps mapstructure.Decode to include the toTimeHookFunc for mapping
// jwt token claims to an AuthToken.
func decode(input map[string]interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}
	return err
}
