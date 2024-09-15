package repository

import (
	"context"
	"errors"
	"fmt"
	custom_errors "zadanie-6105/internal/custom-errors"
	"zadanie-6105/internal/model"
	"zadanie-6105/pkg/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TenderRepository struct {
	*postgres.DB
}

func NewTenderRepository(db *postgres.DB) *TenderRepository {
	return &TenderRepository{db}
}

func (tR *TenderRepository) GetTenders(
	ctx context.Context,
	limit int,
	offset int,
	serviceTypesArr []string,
) ([]model.Tender, error) {
	path := "internal.repository.tender.GetTenders"
	sql := `SELECT id, title, description, service_type, status, version, created_at
	        FROM tender`

	args := []interface{}{limit, offset}
	if len(serviceTypesArr) > 0 {
		sql += ` WHERE service_type = ANY($3)`
		args = append(args, serviceTypesArr)
	}

	sql += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := tR.DB.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf(path+".Query, error: {%s}", err.Error())
	}
	defer rows.Close()
	tenders := make([]model.Tender, 0)
	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(&tender.ID,
			&tender.Title,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.Version,
			&tender.CreatedAt,
		)
		if err != nil {
			return []model.Tender{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
		}
		tenders = append(tenders, tender)
	}
	if len(tenders) == 0 {
		return nil, custom_errors.ErrTenderNotFound
	}

	return tenders, nil
}

func (tR *TenderRepository) CreateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	path := "internal.repository.tender.CreateTender"

	sql := `INSERT INTO tender 
    (
     organization_id,
     title,
     description,
     service_type,
     status,
		 creator_username) VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, title, description, service_type, version, status, created_at
		 `
	var res model.Tender
	err := tR.DB.Pool.QueryRow(ctx, sql,
		tender.OrganizationID,
		tender.Title,
		tender.Description,
		tender.ServiceType,
		tender.Status,
		tender.CreatorUsername,
	).Scan(
		&res.ID,
		&res.Title,
		&res.Description,
		&res.ServiceType,
		&res.Version,
		&res.Status,
		&res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return model.Tender{}, custom_errors.ErrTenderAlreadyExists
			}
		}
		return model.Tender{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}

	return res, nil
}

func (tR *TenderRepository) IsUserResponsibleForOrganization(ctx context.Context, tender model.Tender) (bool, error) {
	var count int
	query := `
        SELECT COUNT(*)
        FROM organization_responsible
        WHERE user_id = (SELECT id FROM employee WHERE username = $1)
        AND organization_id = $2
    `
	err := tR.DB.Pool.QueryRow(ctx, query, tender.CreatorUsername, tender.OrganizationID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (tR *TenderRepository) GetTender(ctx context.Context, user string, limit int, offset int) ([]model.Tender, error) {
	path := "internal.repository.tender.GetTender"
	sql := `SELECT * FROM tender WHERE creator_username = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := tR.DB.Pool.Query(ctx, sql, user, limit, offset)
	if err != nil {
		return []model.Tender{}, fmt.Errorf(path+".Query, error: {%s}", err.Error())
	}

	defer rows.Close()
	tenders := make([]model.Tender, 0)
	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(&tender.ID,
			&tender.OrganizationID,
			&tender.Title,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt,
			&tender.CreatorUsername,
		)
		if err != nil {
			return []model.Tender{}, fmt.Errorf(path+".Scan, error: {%s}", err.Error())
		}
		tenders = append(tenders, tender)
	}
	if len(tenders) == 0 {
		return []model.Tender{}, custom_errors.ErrTenderNotFound
	}
	return tenders, nil
}

func (tR *TenderRepository) UpdateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	path := "internal.repository.tender.UpdateTender"
	sql := `SELECT creator_username FROM tender WHERE id = $1`
	var creatorUsername string
	err := tR.DB.Pool.QueryRow(ctx, sql, tender.ID).Scan(&creatorUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tender{}, custom_errors.ErrUserNotFound
		}
		return model.Tender{}, fmt.Errorf(path+".QueryRow (creator check), error: {%s}", err.Error())
	}
	if creatorUsername != tender.CreatorUsername {
		return model.Tender{}, custom_errors.ErrAccessDenied
	}
	query := "UPDATE tender SET updated_at = NOW(), version = version + 1, "
	params := []interface{}{}
	paramIndex := 1

	if tender.Title != "" {
		query += fmt.Sprintf("title = $%d, ", paramIndex)
		params = append(params, tender.Title)
		paramIndex++
	}
	if tender.Description != "" {
		query += fmt.Sprintf("description = $%d, ", paramIndex)
		params = append(params, tender.Description)
		paramIndex++
	}
	if tender.Status != "" {
		query += fmt.Sprintf("status = $%d, ", paramIndex)
		params = append(params, tender.Status)
		paramIndex++
	}

	query = query[:len(query)-2] // Убираем последнюю запятую
	query += fmt.Sprintf(" WHERE id = $%d ", paramIndex)
	params = append(params, tender.ID)

	query += "RETURNING id, title, description, service_type, status, version, created_at"

	var res model.Tender
	err = tR.DB.Pool.QueryRow(ctx, query, params...).
		Scan(&res.ID,
			&res.Title,
			&res.Description,
			&res.ServiceType,
			&res.Status,
			&res.Version,
			&res.CreatedAt)
	if err != nil {
		return model.Tender{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return res, nil
}

func (tR *TenderRepository) GetStatus(ctx context.Context, tenderId uuid.UUID) (string, error) {
	path := "internal.repository.tender.GetStatus"
	sql := `SELECT status FROM tender WHERE id = $1`

	var status string
	err := tR.DB.Pool.QueryRow(ctx, sql, tenderId).Scan(&status)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			return "", err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return "", custom_errors.ErrTenderNotFound
		}
		return "", fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return status, nil
}

func (tR *TenderRepository) UpdateStatus(ctx context.Context, tender model.Tender) (model.Tender, error) {
	path := "internal.repository.tender.UpdateStatus"
	sql := `SELECT creator_username FROM tender WHERE id = $1`
	var creatorUsername string
	err := tR.DB.Pool.QueryRow(ctx, sql, tender.ID).Scan(&creatorUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tender{}, custom_errors.ErrUserNotFound
		}
		return model.Tender{}, fmt.Errorf(path+".QueryRow (creator check), error: {%s}", err.Error())
	}
	if creatorUsername != tender.CreatorUsername {
		return model.Tender{}, custom_errors.ErrAccessDenied
	}
	sql = `UPDATE tender SET status = $1, creator_username = $2, updated_at = NOW(),version = version + 1  
              WHERE id = $3 RETURNING id, title, description, service_type, status, version, created_at`

	var res model.Tender

	err = tR.DB.Pool.QueryRow(ctx, sql, tender.Status, tender.CreatorUsername, tender.ID).
		Scan(&res.ID,
			&res.Title,
			&res.Description,
			&res.ServiceType,
			&res.Status,
			&res.Version,
			&res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tender{}, custom_errors.ErrTenderNotFound
		}
		return model.Tender{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return res, nil
}
