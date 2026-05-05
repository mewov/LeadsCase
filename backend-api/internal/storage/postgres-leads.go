package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type (
	Lead struct {
		Id                string
		ClientName        string
		ClientTitle       string
		ClientDescription string
		ClientContact     string
		Status            string
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}

	LeadSearchParams struct {
		Status   string
		Query    string
		DateFrom *time.Time
		DateTo   *time.Time
		Limit    int
		Offset   int
		Sort     string
		Order    string
	}
)

// CreateLead - returns id, error
func (p *Postgres) CreateLead(clientName string, clientTitle string, clientDescription string, clientContact string) (string, error) {
	var id string
	err := p.db.QueryRow("INSERT INTO leads(client_name, client_title, client_description, client_contact) VALUES ($1, $2, $3, $4) RETURNING id",
		clientName, clientTitle, clientDescription, clientContact).Scan(&id)

	if err != nil {
		p.logg.Error("create lead", "error", err)
		return "", err
	}

	p.logg.Info("create lead", "name", clientName)
	return id, nil
}

func (p *Postgres) ChangeLeadStatus(id string, status string) error {
	result, err := p.db.Exec("UPDATE leads SET status = $1, updated_at = now() WHERE id = $2", status, id)
	if err != nil {
		p.logg.Error("update lead", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		p.logg.Error("update lead rows affected", "error", err, "id", id)
		return err
	}
	if rowsAffected == 0 {
		p.logg.Warn("update lead not found", "id", id)
		return sql.ErrNoRows
	}

	p.logg.Info("update lead", "id", id)
	return nil
}

func (p *Postgres) SearchLeadById(id string) (*Lead, error) {
	lead := &Lead{}
	err := p.db.QueryRow("SELECT id, client_name, client_title, client_description, client_contact, status, created_at, updated_at FROM leads WHERE id = $1", id).Scan(
		&lead.Id,
		&lead.ClientName, &lead.ClientTitle,
		&lead.ClientDescription, &lead.ClientContact,
		&lead.Status, &lead.CreatedAt, &lead.UpdatedAt,
	)
	if err != nil {
		p.logg.Error("query lead", "error", err, "id", id)
		return nil, err
	}

	p.logg.Info("query lead", "id", id)
	return lead, nil
}

func (p *Postgres) SearchLeads(params *LeadSearchParams) ([]*Lead, error) {
	if params == nil {
		params = &LeadSearchParams{}
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	query := `
		SELECT id, client_name, client_title, client_description, client_contact, status, created_at, updated_at
		FROM leads
	`
	args := make([]any, 0)
	where := make([]string, 0)

	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	status := strings.TrimSpace(params.Status)
	if status != "" {
		where = append(where, "status = "+addArg(status))
	}

	searchQuery := strings.TrimSpace(params.Query)
	if searchQuery != "" {
		placeholder := addArg("%" + searchQuery + "%")
		where = append(where, fmt.Sprintf(
			"(client_name ILIKE %s OR client_title ILIKE %s OR client_description ILIKE %s OR client_contact ILIKE %s)",
			placeholder, placeholder, placeholder, placeholder,
		))
	}

	if params.DateFrom != nil {
		where = append(where, "created_at >= "+addArg(*params.DateFrom))
	}

	if params.DateTo != nil {
		where = append(where, "created_at <= "+addArg(*params.DateTo))
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	sortFields := map[string]string{
		"client_name": "client_name",
		"clientname":  "client_name",
		"name":        "client_name",
		"status":      "status",
		"created_at":  "created_at",
		"createdat":   "created_at",
		"created":     "created_at",
		"updated_at":  "updated_at",
		"updatedat":   "updated_at",
		"updated":     "updated_at",
	}

	sort := sortFields[strings.ToLower(strings.TrimSpace(params.Sort))]
	if sort == "" {
		sort = "created_at"
	}

	order := strings.ToUpper(strings.TrimSpace(params.Order))
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s, id DESC", sort, order)
	query += " LIMIT " + addArg(params.Limit)

	if params.Offset > 0 {
		query += " OFFSET " + addArg(params.Offset)
	}

	rows, err := p.db.Query(query, args...)
	if err != nil {
		p.logg.Error("query leads failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	leads := make([]*Lead, 0)
	for rows.Next() {
		lead := &Lead{}
		err = rows.Scan(
			&lead.Id,
			&lead.ClientName,
			&lead.ClientTitle,
			&lead.ClientDescription,
			&lead.ClientContact,
			&lead.Status,
			&lead.CreatedAt,
			&lead.UpdatedAt,
		)
		if err != nil {
			p.logg.Error("scan lead failed", "error", err)
			return nil, err
		}

		leads = append(leads, lead)
	}

	if err = rows.Err(); err != nil {
		p.logg.Error("query leads rows failed", "error", err)
		return nil, err
	}

	p.logg.Info(
		"query leads succeeded",
		"count", len(leads),
		"limit", params.Limit,
		"offset", params.Offset,
	)

	return leads, nil
}
