package models

type (
	CreateLeadRequest struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Contact     string `json:"contact"`
	}

	LoginAdminRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	RefreshAdminRequest struct {
		Refresh string `json:"refresh"`
	}

	LogoutAdminRequest struct {
		Refresh string `json:"refresh"`
	}

	ChangeLeadStatusRequest struct {
		Status string `json:"status"`
	}
)
