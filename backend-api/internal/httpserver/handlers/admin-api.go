package handlers

import (
	app2 "demo-server/internal/app"
	models2 "demo-server/internal/httpserver/models"
	"demo-server/internal/storage"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func HandleAdminLogin(postgres *storage.Postgres, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models2.LoginAdminRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models2.Response{IsOk: false, Message: "invalid request"})
			return
		}

		tokens, err := app2.LoginAdminSession(postgres, jwtSecret, request.Login, request.Password)
		if err != nil {
			switch {
			case errors.Is(err, app2.ErrInvalidValues):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid values"})
			case errors.Is(err, app2.ErrSession):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database session error"})
			case errors.Is(err, app2.ErrJWT):
				writeJSON(w, http.StatusInternalServerError, models2.Response{IsOk: false, Message: "jwt error"})
			default:
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database login error"})
			}
			return
		}

		writeJSON(w, http.StatusOK, models2.Response{IsOk: true, Message: "login", Payload: models2.ResponseTokens{
			Refresh: tokens.Refresh,
			Access:  tokens.Access,
		}})
	}
}

func HandleAdminRefresh(postgres *storage.Postgres, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models2.RefreshAdminRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models2.Response{IsOk: false, Message: "invalid request"})
			return
		}

		tokens, err := app2.RefreshAdminSession(postgres, jwtSecret, request.Refresh)
		if err != nil {
			switch {
			case errors.Is(err, app2.ErrInvalidValues):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid values"})
			case errors.Is(err, app2.ErrTokenExpired):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "token expired"})
			case errors.Is(err, app2.ErrJWT):
				writeJSON(w, http.StatusInternalServerError, models2.Response{IsOk: false, Message: "jwt error"})
			default:
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database session error"})
			}
			return
		}

		writeJSON(w, http.StatusOK, models2.Response{IsOk: true, Message: "refresh", Payload: models2.ResponseTokens{
			Refresh: tokens.Refresh,
			Access:  tokens.Access,
		}})
	}
}

func HandleAdminLogout(postgres *storage.Postgres) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models2.LogoutAdminRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid request"})
			return
		}

		if err := app2.LogoutAdmin(postgres, request.Refresh); err != nil {
			writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database session error"})
			return
		}

		writeJSON(w, http.StatusOK, models2.Response{IsOk: true, Message: "logout"})
	}
}
