package models

import "time"

type (
	Response struct {
		IsOk    bool   `json:"is_ok"`
		Message string `json:"message"`
		Payload any    `json:"payload"`
	}

	ResponseLeadCreate struct {
		ID string `json:"id"`
	}
	ResponseTokens struct {
		Refresh string `json:"refresh"`
		Access  string `json:"access"`
	}
	ResponseLead struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Contact     string    `json:"contact"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	ResponseLeadsList struct {
		Items  []ResponseLead `json:"items"`
		Count  int            `json:"count"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
)
