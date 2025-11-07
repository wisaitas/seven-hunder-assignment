package jwtx

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/gofiber/fiber/v2"
)

type Claims interface {
	GetID() string
}

type StandardClaims struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

func (s StandardClaims) GetID() string {
	return s.ID
}

type Jwt interface {
	Generate(claims interface{}, secret string) (string, error)
	Parse(tokenString string, secret string, result interface{}) error
	ExtractTokenFromHeader(c *fiber.Ctx) (string, error)
	ValidateToken(tokenString string, secret string) error
	CreateStandardClaims(id string, expireTime time.Duration) StandardClaims
}

type jwtx struct{}

func NewJwt() Jwt {
	return &jwtx{}
}

func (j *jwtx) Generate(claims interface{}, secret string) (string, error) {
	// สร้าง encrypter ด้วย AES-GCM
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.DIRECT,
			Key:       []byte(secret),
		},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"),
	)
	if err != nil {
		return "", fmt.Errorf("[jwtx] failed to create encrypter: %w", err)
	}

	// แปลง claims เป็น JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("[jwtx] failed to marshal claims: %w", err)
	}

	// Encrypt payload
	jwe, err := encrypter.Encrypt(claimsJSON)
	if err != nil {
		return "", fmt.Errorf("[jwtx] failed to encrypt: %w", err)
	}

	// Serialize เป็น compact format
	tokenString, err := jwe.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("[jwtx] failed to serialize: %w", err)
	}

	return tokenString, nil
}

func (j *jwtx) Parse(tokenString string, secret string, result interface{}) error {
	// Parse JWE token
	jwe, err := jose.ParseEncrypted(tokenString, []jose.KeyAlgorithm{jose.DIRECT}, []jose.ContentEncryption{jose.A256GCM})
	if err != nil {
		return fmt.Errorf("[jwtx] failed to parse token: %w", err)
	}

	// Decrypt payload
	decrypted, err := jwe.Decrypt([]byte(secret))
	if err != nil {
		return fmt.Errorf("[jwtx] failed to decrypt: %w", err)
	}

	// Unmarshal claims
	if err := json.Unmarshal(decrypted, result); err != nil {
		return fmt.Errorf("[jwtx] failed to unmarshal claims: %w", err)
	}

	return nil
}

func (j *jwtx) ExtractTokenFromHeader(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("[jwtx] invalid token type")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func (j *jwtx) ValidateToken(tokenString string, secret string) error {
	var claims StandardClaims
	err := j.Parse(tokenString, secret, &claims)
	if err != nil {
		return fmt.Errorf("[jwtx] %w", err)
	}

	// ตรวจสอบว่า token หมดอายุหรือไม่
	if time.Now().After(claims.ExpiresAt) {
		return fmt.Errorf("[jwtx] token expired")
	}

	return nil
}

func (j *jwtx) CreateStandardClaims(id string, expireTime time.Duration) StandardClaims {
	return StandardClaims{
		ID:        id,
		ExpiresAt: time.Now().Add(expireTime),
		IssuedAt:  time.Now(),
	}
}
