package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id                  string         `db:"id" json:"id"`
	MTime               int            `db:"mtime" json:"mtime"`
	FirstName           string         `db:"first_name" json:"first_name"`
	LastName            string         `db:"last_name" json:"last_name"`
	UserName            string         `db:"username" json:"username"`
	Password            string         `db:"password" json:"password"`
	Email               string         `db:"email" json:"email"`
	LastPresence        string         `db:"last_presence" json:"last_presence"`
	LoginFailedAttempts int            `db:"login_failed_attempts" json:"login_failed_attempts"`
	LastLoginFailed     sql.NullTime   `db:"last_login_failed" json:"last_login_failed"` // Use sql.NullTime
	TotpEnabled         bool           `db:"totp_enabled" json:"totp_enabled"`
	TotpSecret          sql.NullString `db:"totp_secret" json:"totp_secret"` // Handle nullable string
	EmailOtpEnabled     bool           `db:"email_otp_enabled" json:"email_otp_enabled"`
	Photo               []byte         `db:"photo" json:"photo"`
	EmailOtpSecret      sql.NullString `db:"email_otp_secret" json:"email_otp_secret"` // Handle nullable string
	HasAvatar           bool           `db:"has_avatar" json:"has_avatar"`
	AddedAt             time.Time      `db:"added_at" json:"added_at"`
	Active              bool           `db:"active" json:"active"`
	IsDeleted           bool           `db:"is_deleted" json:"is_deleted"`
}
type Studio struct {
	Id        string `db:"id" json:"id"`
	MTime     int    `db:"mtime" json:"mtime"`
	Name      string `db:"name" json:"name"`
	Key       string `db:"key" json:"key"`
	URL       string `db:"url" json:"url"`
	AltURL    string `db:"alt_url" json:"alt_url"`
	Active    string `db:"active" json:"active"`
	IsDeleted string `db:"is_deleted" json:"is_deleted"`
}

type MinimalStudio struct {
	Id     string `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	URL    string `db:"url" json:"url"`
	AltURL string `db:"alt_url" json:"alt_url"`
	Active string `db:"active" json:"active"`
}
type StudioUser struct {
	Id       string `db:"id" json:"id"`
	MTime    int    `db:"mtime" json:"mtime"`
	StudioId string `db:"studio_id" json:"studio_id"`
	UserId   string `db:"user_id" json:"user_id"`
	RoleId   string `db:"role_id" json:"role_id"`
}

type StudioUserInfo struct {
	Id         string `db:"id" json:"id"`
	FirstName  string `db:"first_name" json:"first_name"`
	LastName   string `db:"last_name" json:"last_name"`
	UserName   string `db:"username" json:"username"`
	Email      string `db:"email" json:"email"`
	Active     bool   `db:"active" json:"active"`
	RoleName   string `db:"role_name" json:"role_name"`
	StudioName string `db:"studio_name" json:"studio_name"`
	StudioId   string `db:"studio_id" json:"studio_id"`
	RoleId     string `db:"role_id" json:"role_id"`
	Photo      []byte `db:"photo" json:"photo"`
}

type Role struct {
	Id    string `db:"id" json:"id"`
	MTime int    `db:"mtime" json:"mtime"`
	Name  string `db:"name" json:"name"`
}

type UserResponse struct {
	Id        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Active    bool   `db:"active" json:"active"`
}
