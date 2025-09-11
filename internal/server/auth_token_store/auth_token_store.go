package auth_token_store

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

func Add(db *sqlx.DB, email, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO tokens (email, token, expires_at) 
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE 
		SET token = $2, expires_at = $3
	`
	_, err := db.Exec(query, email, token, expiresAt)
	if err != nil {
		return err
	}
	return nil
}

func Get(db *sqlx.DB, email string) (string, error) {
	token := ""
	var expiresAt time.Time
	query := "SELECT token, expires_at FROM tokens WHERE email = ?"
	err := db.QueryRow(query, email).Scan(&token, &expiresAt)
	if err != nil {
		return token, err
	}
	if time.Now().After(expiresAt) {
		return token, errors.New("refresh token expired")
	}
	return token, nil
}

func Delete(db *sqlx.DB, email string) error {
	query := "DELETE FROM tokens WHERE email = ?"
	_, err := db.Exec(query, email)
	if err != nil {
		return err
	}
	return nil
}
