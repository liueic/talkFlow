package models

import (
	"database/sql"
	"time"
)

type Register struct {
	ID            int64          `json:"id" db:"id"`
	Username      string         `json:"username" db:"username"`
	Password      string         `json:"password" db:"password"`
	Email         string         `json:"email" db:"email"`
	Avatar        string         `json:"avatar" db:"avatar"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	RegisterIP    string         `json:"register_ip" db:"register_ip"`
	IsRegister    bool           `json:"is_register" db:"is_register"`
	LastLoginIP   sql.NullString `json:"last_login_ip,omitempty" db:"last_login_ip"`
	LastLoginTime sql.NullTime   `json:"last_login_time,omitempty" db:"last_login_time"`
}

type Visitor struct {
	ID         int64     `json:"id" db:"id"`
	VisitorID  string    `json:"visitor_id" db:"visitor_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	VisitorIP  string    `json:"visitor_ip" db:"visitor_ip"`
	IsRegister bool      `json:"is_register" db:"is_register"`
}
