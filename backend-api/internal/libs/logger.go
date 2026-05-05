package libs

import (
	"log/slog"
	"os"
	"path/filepath"
)

func CreateLogger(path string) (*slog.Logger, error) {
	os.MkdirAll(filepath.Dir(path), 0755)
	source, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	return slog.New(slog.NewJSONHandler(
		source, &slog.HandlerOptions{},
	)), nil
}

func MustCreateLogger(path string) *slog.Logger {
	logg, err := CreateLogger(path)
	if err != nil {
		panic(err)
	}
	return logg
}
