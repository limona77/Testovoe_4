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

type BidsRepository struct {
	*postgres.DB
}

func NewBidsRepository(db *postgres.DB) *BidsRepository {
	return &BidsRepository{db}
}

func (bR *BidsRepository) CreateBids(ctx context.Context, bids *model.Bids) (model.Bids, error) {
	path := "internal.repository.bids.NewBids"

	sql := `INSERT INTO bids (tender_id,
                  organization_id,
                  title,
                  description,
                  status,
                  creator_username)
					VALUES ($1, $2, $3, $4, $5, $6)
					RETURNING id,
                  title,
                  description,
                  status,
                  version,
                  creator_username,
									created_at`
	var res model.Bids
	err := bR.DB.Pool.QueryRow(ctx, sql,
		bids.TenderID,
		bids.OrganizationID,
		bids.Title,
		bids.Description,
		bids.Status,
		bids.CreatorUsername).Scan(&res.ID,
		&res.Title,
		&res.Description,
		&res.Status,
		&res.Version,
		&res.CreatorUsername,
		&res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return model.Bids{}, custom_errors.ErrBidsAlreadyExists
			}
		}
		return model.Bids{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return res, nil
}

func (bR *BidsRepository) IsUserAuthorizedToCreateBid(
	ctx context.Context,
	bids *model.Bids,
) (bool, error) {
	path := "internal.repository.bids.IsUserAuthorizedToCreateBid"
	query := `
		SELECT COUNT(*)
		FROM organization_responsible 
		WHERE user_id = (SELECT id FROM employee WHERE username = $1)
		AND organization_id = $2
		UNION
		SELECT COUNT(*)
		FROM employee 
		WHERE username = $1
	`

	var count int
	err := bR.DB.Pool.QueryRow(ctx, query, bids.CreatorUsername, bids.OrganizationID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf(path+".Scan, error: {%s}", err.Error())
	}
	return count > 0, nil
}

func (bR *BidsRepository) IsTenderValid(ctx context.Context, tenderID uuid.UUID) (bool, error) {
	path := "internal.repository.tender.IsTenderValid"
	query := `
		SELECT COUNT(*)
		FROM tender
		WHERE id = $1
		AND status <> 'Canceled';
	`

	var count int
	err := bR.DB.Pool.QueryRow(ctx, query, tenderID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf(path+".Scan, error: {%s}", err.Error())
	}

	return count > 0, nil
}

func (bR *BidsRepository) GetBids(ctx context.Context, user string, limit, offset int) ([]model.Bids, error) {
	path := "internal.repository.bids.GetBids"
	sql := `SELECT id, tender_id, organization_id, title, description, status, version, creator_username
					FROM bids WHERE creator_username = $1
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var res []model.Bids
	rows, err := bR.DB.Pool.Query(ctx, sql, user, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(path+".Query, error: {%s}", err.Error())
	}

	for rows.Next() {
		var bids model.Bids
		err = rows.Scan(&bids.ID,
			&bids.TenderID,
			&bids.OrganizationID,
			&bids.Title,
			&bids.Description,
			&bids.Status,
			&bids.Version,
			&bids.CreatorUsername)
		if err != nil {
			var pgErr *pgconn.PgError
			if ok := errors.As(err, &pgErr); ok {
				return []model.Bids{}, err
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return []model.Bids{}, custom_errors.ErrBidsNotFound
			}
			return []model.Bids{}, fmt.Errorf(path+".Scan, error: {%s}", err.Error())
		}
		res = append(res, bids)
	}
	if len(res) == 0 {
		return nil, custom_errors.ErrBidsNotFound
	}
	return res, nil
}

func (bR *BidsRepository) GetBidsByTenderId(
	ctx context.Context,
	user string,
	tenderId uuid.UUID,
	limit, offset int,
) ([]model.Bids, error) {
	path := "internal.repository.bids.GetBidsByTenderId"
	sql := `SELECT id,
                  title,
                  description,
                  status,
                  version,
                  creator_username,
									created_at
					FROM bids
					WHERE tender_id = $1 AND creator_username = $2
					ORDER BY created_at DESC
					LIMIT $3 OFFSET $4`

	var res []model.Bids
	rows, err := bR.DB.Pool.Query(ctx, sql, tenderId, user, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(path+".Query, error: {%s}", err.Error())
	}

	for rows.Next() {
		var bids model.Bids
		err = rows.Scan(&bids.ID,
			&bids.Title,
			&bids.Description,
			&bids.Status,
			&bids.Version,
			&bids.CreatorUsername,
			&bids.CreatedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if ok := errors.As(err, &pgErr); ok {
				return []model.Bids{}, err
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return []model.Bids{}, custom_errors.ErrBidsNotFound
			}
			return []model.Bids{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
		}
		res = append(res, bids)
	}
	if len(res) == 0 {
		return []model.Bids{}, custom_errors.ErrBidsNotFound
	}
	return res, nil
}

func (bR *BidsRepository) CheckUserExists(ctx context.Context, username string) (bool, error) {
	path := "internal.repository.user.CheckUserExists"
	sql := `SELECT 1 FROM employee WHERE username = $1 LIMIT 1`

	var exists int
	err := bR.DB.Pool.QueryRow(ctx, sql, username).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return exists == 1, nil
}

func (bR *BidsRepository) UpdateBids(ctx context.Context, bids *model.Bids) (model.Bids, error) {
	path := "internal.repository.bids.UpdateBids"

	checkUserSQL := `SELECT 1 FROM bids WHERE id = $1 AND creator_username = $2`

	var userExists int

	err := bR.DB.Pool.QueryRow(ctx, checkUserSQL, bids.ID, bids.CreatorUsername).Scan(&userExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bids{}, custom_errors.ErrBidsNotFound
		}
		return model.Bids{}, fmt.Errorf(path+".CheckUser, error: {%s}", err.Error())
	}
	if userExists == 0 {
		return model.Bids{}, custom_errors.ErrUserNotFound
	}
	query := "UPDATE bids SET updated_at = NOW(),version = version + 1, "
	params := []interface{}{}
	paramIndex := 1

	if bids.Title != "" {
		query += fmt.Sprintf("title = $%d, ", paramIndex)
		params = append(params, bids.Title)
		paramIndex++
	}
	if bids.Description != "" {
		query += fmt.Sprintf("description = $%d, ", paramIndex)
		params = append(params, bids.Description)
		paramIndex++
	}

	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d ", paramIndex)
	params = append(params, bids.ID)

	query += `RETURNING id,
                  title,
                  description,
                  status,
                  version,
                  creator_username,
									created_at`

	var res model.Bids
	err = bR.DB.Pool.QueryRow(ctx, query, params...).
		Scan(&res.ID,
			&res.Title,
			&res.Description,
			&res.Status,
			&res.Version,
			&res.CreatorUsername,
			&res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			return model.Bids{}, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bids{}, custom_errors.ErrBidsNotFound
		}
		return model.Bids{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}
	return res, nil
}

func (bR *BidsRepository) GetBidStatus(ctx context.Context, bidId uuid.UUID, user string) (string, error) {
	path := "internal.repository.bids.GetStatus"

	sql := `SELECT status FROM bids WHERE id = $1 AND creator_username = $2`

	var status string
	err := bR.DB.Pool.QueryRow(ctx, sql, bidId, user).Scan(&status)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			return "", err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// Возвращаем ошибку, если предложение не найдено
			return "", custom_errors.ErrBidsNotFound
		}
		return "", fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}

	return status, nil
}

func (bR *BidsRepository) UpdateBidsStatus(ctx context.Context, bids model.Bids) (model.Bids, error) {
	path := "internal.repository.bids.UpdateBidsStatus"

	sql := `UPDATE bids SET status = $1, updated_at = NOW() 
			WHERE id = $2 AND creator_username = $3 
			RETURNING id,
                  title,
                  description,
                  status,
                  version,
                  creator_username,
									created_at`

	var res model.Bids
	err := bR.DB.Pool.QueryRow(ctx, sql, bids.Status, bids.ID, bids.CreatorUsername).
		Scan(&res.ID,
			&res.Title,
			&res.Description,
			&res.Status,
			&res.Version,
			&res.CreatorUsername,
			&res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			return model.Bids{}, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bids{}, custom_errors.ErrBidsNotFound
		}
		return model.Bids{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}

	return res, nil
}

func (bR *BidsRepository) UpdateBidsDecision(
	ctx context.Context,
	bidId uuid.UUID,
	decision, username string,
) (model.Bids, error) {
	path := "internal.repository.bids.UpdateBidsDecision"

	checkBidSQL := `SELECT creator_username FROM bids WHERE id = $1`
	var creatorUsername string
	err := bR.DB.Pool.QueryRow(ctx, checkBidSQL, bidId).Scan(&creatorUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bids{}, custom_errors.ErrBidsNotFound
		}
		return model.Bids{}, fmt.Errorf(path+".CheckBid, error: {%s}", err.Error())
	}

	if creatorUsername != username {
		return model.Bids{}, custom_errors.ErrUserNotFound
	}

	query := `UPDATE bids
	          SET decision = $1
	          WHERE id = $2
	          RETURNING id,
                  title,
                  description,
                  status,
                  version,
                  creator_username,
									created_at`
	var res model.Bids
	err = bR.DB.Pool.QueryRow(ctx, query, decision, bidId).Scan(
		&res.ID,
		&res.Title, &res.Description, &res.Status,
		&res.Version, &res.CreatorUsername, &res.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bids{}, custom_errors.ErrBidsNotFound
		}
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			return model.Bids{}, err
		}
		return model.Bids{}, fmt.Errorf(path+".QueryRow, error: {%s}", err.Error())
	}

	return res, nil
}
