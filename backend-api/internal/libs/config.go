package libs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr         string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerAdminLogin   string
	ServerAdminPass    string
	JWTSecret          string

	ServerLogPath         string
	ServerHttpLogPath     string
	ServerPostgresLogPath string

	PostgresAddr     string
	PostgresUser     string
	PostgresPassword string
	PostgresDb       string
}

var requiredEnvKeys = []string{
	"SERVER_ADDR",
	"SERVER_ADMIN_LOGIN",
	"SERVER_ADMIN_PASSWORD",
	"JWT_SECRET",
	"SERVER_LOG_PATH",
	"SERVER_HTTP_LOG_PATH",
	"SERVER_POSTGRES_LOG_PATH",
	"POSTGRES_ADDR",
	"POSTGRES_USER",
	"POSTGRES_PASSWORD",
	"POSTGRES_DB",
}

func CreateConfig() (*Config, error) {
	readTimeout, err := seconds("SERVER_READ_TIMEOUT_SECOND")
	if err != nil {
		return nil, err
	}

	writeTimeout, err := seconds("SERVER_WRITE_TIMEOUT_SECOND")
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(requiredEnvKeys))
	for _, key := range requiredEnvKeys {
		value, err := required(key)
		if err != nil {
			return nil, err
		}

		values[key] = value
	}

	return &Config{
		ServerAddr:            values["SERVER_ADDR"],
		ServerReadTimeout:     readTimeout,
		ServerWriteTimeout:    writeTimeout,
		ServerAdminLogin:      strings.ToLower(values["SERVER_ADMIN_LOGIN"]),
		ServerAdminPass:       values["SERVER_ADMIN_PASSWORD"],
		JWTSecret:             values["JWT_SECRET"],
		PostgresAddr:          values["POSTGRES_ADDR"],
		PostgresUser:          values["POSTGRES_USER"],
		PostgresPassword:      values["POSTGRES_PASSWORD"],
		PostgresDb:            values["POSTGRES_DB"],
		ServerLogPath:         values["SERVER_LOG_PATH"],
		ServerHttpLogPath:     values["SERVER_HTTP_LOG_PATH"],
		ServerPostgresLogPath: values["SERVER_POSTGRES_LOG_PATH"],
	}, nil
}

func seconds(key string) (time.Duration, error) {
	value, err := required(key)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}

	return time.Duration(seconds) * time.Second, nil
}

func required(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return value, nil
}
