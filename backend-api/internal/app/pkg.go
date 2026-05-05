package app

import (
	"crypto/rand"
	"demo-server/internal/storage"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const refresh_len = 64
const refresh_symbols = "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm0123456789"
const jwtTTL = time.Minute * 30

func GenerateRefresh() string {
	bytes := make([]byte, refresh_len)
	for i := range bytes {
		number, _ := rand.Int(rand.Reader, big.NewInt(int64(len(refresh_symbols))))
		bytes[i] = refresh_symbols[number.Int64()]
	}
	return string(bytes)
}

type jwtClaims struct {
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidValues = errors.New("invalid values")
	ErrInvalidQuery  = errors.New("invalid query")
	ErrNotFound      = errors.New("not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrLogin         = errors.New("login error")
	ErrSession       = errors.New("session error")
	ErrJWT           = errors.New("jwt error")
	ErrDatabase      = errors.New("database error")

	errEmptyJwtSecret = errors.New("jwt secret is empty")
	errInvalidJwt     = errors.New("invalid jwt")
)

func CreateJwt(admin *storage.Admin, jwtSecretValue string) (string, error) {
	if admin == nil {
		return "", errors.New("admin is nil")
	}

	secret, err := jwtSecret(jwtSecretValue)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwtClaims{
		Login:     admin.Login,
		CreatedAt: admin.CreatedAt,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   admin.Id,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtTTL)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func ValidateJwt(token string, jwtSecretValue string) (*storage.Admin, error) {
	secret, err := jwtSecret(jwtSecretValue)
	if err != nil {
		return nil, err
	}

	claims := &jwtClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, errInvalidJwt
	}
	if !parsedToken.Valid {
		return nil, errInvalidJwt
	}
	if strings.TrimSpace(claims.Subject) == "" || strings.TrimSpace(claims.Login) == "" {
		return nil, errInvalidJwt
	}

	return &storage.Admin{
		Id:        claims.Subject,
		Login:     claims.Login,
		CreatedAt: claims.CreatedAt,
	}, nil
}

func jwtSecret(value string) ([]byte, error) {
	secret := strings.TrimSpace(value)
	if secret == "" {
		return nil, errEmptyJwtSecret
	}

	return []byte(secret), nil
}
