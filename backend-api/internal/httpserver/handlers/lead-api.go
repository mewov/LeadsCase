package handlers

import (
	app2 "demo-server/internal/app"
	models2 "demo-server/internal/httpserver/models"
	"demo-server/internal/storage"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
)

func HandleCreateLead(postgres *storage.Postgres) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models2.CreateLeadRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models2.Response{IsOk: false, Message: "invalid request"})
			return
		}

		lastId, err := app2.CreateLead(postgres, app2.CreateLeadInput{
			Name:        request.Name,
			Title:       request.Title,
			Description: request.Description,
			Contact:     request.Contact,
		})
		if err != nil {
			if errors.Is(err, app2.ErrInvalidValues) {
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid values"})
				return
			}

			writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database error"})
			return
		}

		writeJSON(w, http.StatusCreated, models2.Response{IsOk: true, Message: "created", Payload: models2.ResponseLeadCreate{ID: lastId}})
	}
}

func HandleListLeads(postgres *storage.Postgres) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		leads, err := app2.ListLeads(postgres, app2.ListLeadsInput{
			Status:   query.Get("status"),
			Search:   []string{query.Get("q"), query.Get("query"), query.Get("search")},
			DateFrom: query.Get("date_from"),
			DateTo:   query.Get("date_to"),
			Limit:    query.Get("limit"),
			Offset:   query.Get("offset"),
			Sort:     query.Get("sort"),
			Order:    query.Get("order"),
		})
		if err != nil {
			if errors.Is(err, app2.ErrInvalidQuery) {
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid query"})
				return
			}

			writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database error"})
			return
		}

		items := make([]models2.ResponseLead, 0, len(leads.Items))
		for _, lead := range leads.Items {
			items = append(items, leadResponse(lead))
		}

		writeJSON(w, http.StatusOK, models2.Response{
			IsOk:    true,
			Message: "leads",
			Payload: models2.ResponseLeadsList{
				Items:  items,
				Count:  leads.Count,
				Limit:  leads.Limit,
				Offset: leads.Offset,
			},
		})
	}
}

func HandleGetLeadByID(postgres *storage.Postgres) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lead, err := app2.GetLeadByID(postgres, chi.URLParam(r, "id"))
		if err != nil {
			switch {
			case errors.Is(err, app2.ErrInvalidValues):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid values"})
			case errors.Is(err, app2.ErrNotFound):
				writeJSON(w, http.StatusNotFound, models2.Response{IsOk: false, Message: "not found"})
			default:
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database error"})
			}
			return
		}

		writeJSON(w, http.StatusOK, models2.Response{
			IsOk:    true,
			Message: "lead",
			Payload: leadResponse(*lead),
		})
	}
}

func HandleChangeLeadStatus(postgres *storage.Postgres) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models2.ChangeLeadStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid request"})
			return
		}

		if err := app2.ChangeLeadStatus(postgres, chi.URLParam(r, "id"), request.Status); err != nil {
			switch {
			case errors.Is(err, app2.ErrInvalidValues):
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "invalid values"})
			case errors.Is(err, app2.ErrNotFound):
				writeJSON(w, http.StatusNotFound, models2.Response{IsOk: false, Message: "not found"})
			default:
				writeJSON(w, http.StatusBadRequest, models2.Response{IsOk: false, Message: "database error"})
			}
			return
		}

		writeJSON(w, http.StatusOK, models2.Response{IsOk: true, Message: "status changed"})
	}
}

func leadResponse(lead app2.Lead) models2.ResponseLead {
	return models2.ResponseLead{
		ID:          lead.ID,
		Name:        lead.Name,
		Title:       lead.Title,
		Description: lead.Description,
		Contact:     lead.Contact,
		Status:      lead.Status,
		CreatedAt:   lead.CreatedAt,
		UpdatedAt:   lead.UpdatedAt,
	}
}
