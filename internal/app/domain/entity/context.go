package entity

import (
	"github.com/golang-jwt/jwt/v5"
)

type ExternalContext struct {
	jwt.RegisteredClaims
}

type UserContext struct {
	User
}
