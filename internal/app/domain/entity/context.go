package entity

import (
	"time"
)

type ExternalContext struct {
	Subject   string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf,omitempty"`
	Issuer    string    `json:"iss,omitempty"`
	Audience  []string  `json:"aud,omitempty"`
	ID        string    `json:"jti,omitempty"`
}

type UserContext struct {
	User
}
