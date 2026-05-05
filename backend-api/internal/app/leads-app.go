package app

import (
	"database/sql"
	storage2 "demo-server/internal/storage"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const MaxLeadStatusLength = 32

type (
	CreateLeadInput struct {
		Name        string
		Title       string
		Description string
		Contact     string
	}

	ListLeadsInput struct {
		Status   string
		Search   []string
		DateFrom string
		DateTo   string
		Limit    string
		Offset   string
		Sort     string
		Order    string
	}

	Lead struct {
		ID          string
		Name        string
		Title       string
		Description string
		Contact     string
		Status      string
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	LeadsList struct {
		Items  []Lead
		Count  int
		Limit  int
		Offset int
	}
)

func CreateLead(postgres *storage2.Postgres, input CreateLeadInput) (string, error) {
	name := strings.TrimSpace(input.Name)
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	contact := strings.TrimSpace(input.Contact)
	if name == "" || title == "" || description == "" || contact == "" {
		return "", ErrInvalidValues
	}

	id, err := postgres.CreateLead(name, title, description, contact)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return id, nil
}

func ListLeads(postgres *storage2.Postgres, input ListLeadsInput) (*LeadsList, error) {
	params, err := leadSearchParams(input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidQuery, err)
	}

	leads, err := postgres.SearchLeads(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	items := make([]Lead, 0, len(leads))
	for _, lead := range leads {
		items = append(items, leadFromStorage(lead))
	}

	return &LeadsList{
		Items:  items,
		Count:  len(items),
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func GetLeadByID(postgres *storage2.Postgres, id string) (*Lead, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidValues
	}

	lead, err := postgres.SearchLeadById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	response := leadFromStorage(lead)
	return &response, nil
}

func ChangeLeadStatus(postgres *storage2.Postgres, id string, status string) error {
	id = strings.TrimSpace(id)
	status = strings.TrimSpace(status)
	if id == "" || status == "" || len(status) > MaxLeadStatusLength {
		return ErrInvalidValues
	}

	err := postgres.ChangeLeadStatus(id, status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		return fmt.Errorf("%w: %w", ErrDatabase, err)
	}

	return nil
}

func leadSearchParams(input ListLeadsInput) (*storage2.LeadSearchParams, error) {
	limit, err := optionalInt(input.Limit)
	if err != nil {
		return nil, err
	}

	offset, err := optionalInt(input.Offset)
	if err != nil {
		return nil, err
	}

	dateFrom, err := optionalTime(input.DateFrom, false)
	if err != nil {
		return nil, err
	}

	dateTo, err := optionalTime(input.DateTo, true)
	if err != nil {
		return nil, err
	}

	return &storage2.LeadSearchParams{
		Status:   input.Status,
		Query:    firstSearchValue(input.Search...),
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    limit,
		Offset:   offset,
		Sort:     input.Sort,
		Order:    input.Order,
	}, nil
}

func optionalInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	number, err := strconv.Atoi(value)
	if err != nil || number < 0 {
		return 0, strconv.ErrSyntax
	}

	return number, nil
}

func optionalTime(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}

	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}

	return &parsed, nil
}

func firstSearchValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func leadFromStorage(lead *storage2.Lead) Lead {
	return Lead{
		ID:          lead.Id,
		Name:        lead.ClientName,
		Title:       lead.ClientTitle,
		Description: lead.ClientDescription,
		Contact:     lead.ClientContact,
		Status:      lead.Status,
		CreatedAt:   lead.CreatedAt,
		UpdatedAt:   lead.UpdatedAt,
	}
}
