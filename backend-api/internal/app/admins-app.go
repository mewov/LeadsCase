package app

import (
	"database/sql"
	"demo-server/internal/libs"
	storage2 "demo-server/internal/storage"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Tokens struct {
	Refresh string
	Access  string
}

func SeedDefaultAdmin(postgres *storage2.Postgres, config *libs.Config) error {
	_, err := postgres.SearchAdmin(config.ServerAdminLogin)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(config.ServerAdminPass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		err = postgres.CreateAdmin(strings.ToLower(strings.TrimSpace(config.ServerAdminLogin)), string(passwordHash))
		return err
	}

	return nil
}

func LoginAdmin(postgres *storage2.Postgres, login string, password string) (*storage2.Admin, error) {
	return authenticateAdmin(postgres, login, password)
}

func LoginAdminSession(postgres *storage2.Postgres, jwtSecret string, login string, password string) (*Tokens, error) {
	admin, err := LoginAdmin(postgres, login, password)
	if err != nil {
		if errors.Is(err, ErrInvalidValues) {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %w", ErrLogin, err)
	}

	return createAdminTokens(postgres, admin, jwtSecret)
}

func RefreshAdminSession(postgres *storage2.Postgres, jwtSecret string, refresh string) (*Tokens, error) {
	refresh = strings.TrimSpace(refresh)
	if refresh == "" {
		return nil, ErrInvalidValues
	}

	session, err := postgres.SearchSession(refresh)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSession, err)
	}

	if err := postgres.DropSession(refresh); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSession, err)
	}

	if session.ExpiresAt.Unix() < time.Now().Unix() {
		return nil, ErrTokenExpired
	}

	admin, err := postgres.SearchAdminById(session.AdminId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSession, err)
	}

	return createAdminTokens(postgres, admin, jwtSecret)
}

func LogoutAdmin(postgres *storage2.Postgres, refresh string) error {
	refresh = strings.TrimSpace(refresh)
	if refresh == "" {
		return nil
	}

	if err := postgres.DropSession(refresh); err != nil {
		return fmt.Errorf("%w: %w", ErrSession, err)
	}

	return nil
}

func authenticateAdmin(postgres *storage2.Postgres, login string, password string) (*storage2.Admin, error) {
	login = strings.ToLower(strings.TrimSpace(login))
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return nil, ErrInvalidValues
	}

	admin, err := postgres.SearchAdmin(login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func CreateAdminSession(postgres *storage2.Postgres, admin *storage2.Admin) (string, error) {
	if admin == nil || strings.TrimSpace(admin.Id) == "" {
		return "", ErrInvalidValues
	}

	token := GenerateRefresh()
	expiresAt := time.Now().Add(time.Hour * 24 * 7)
	_, err := postgres.CreateSession(admin.Id, token, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func createAdminTokens(postgres *storage2.Postgres, admin *storage2.Admin, jwtSecret string) (*Tokens, error) {
	refresh, err := CreateAdminSession(postgres, admin)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSession, err)
	}

	access, err := CreateJwt(admin, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrJWT, err)
	}

	return &Tokens{
		Refresh: refresh,
		Access:  access,
	}, nil
}
