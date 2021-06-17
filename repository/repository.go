package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"time"
)

const (
	ForwardingRequestType = "ForwardingRequestType"
)

type Repository struct {
	db *sql.DB
}

type NotFound struct {
	ID   string
	Type string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("no %s with id %q", e.Type, e.ID)
}

type duplicateEntry struct {
}

func (e duplicateEntry) Error() string {
	return "duplicate "
}

var DuplicateEntry = duplicateEntry{}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveForwardingRequest(ctx context.Context, request *model.ForwardingRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO ForwardingRequests(id, requester_id, requester_name, duration, created_at, accepted_at, refused_at, expires_at) VALUES(?,?,?,?,?,?,?,?)",
		request.Id,
		request.RequesterId,
		request.RequesterName,
		request.Duration,
		request.CreatedAt,
		request.AcceptedAt,
		request.RefusedAt,
		request.ExpiresAt,
	)

	_ = res

	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return DuplicateEntry
		}
	}

	return err
}

func (r *Repository) AcceptForwardingRequest(ctx context.Context, requestId string, answeredBy string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE ForwardingRequests SET accepted_at = ?, expires_at = DATE_ADD(?, INTERVAL duration minute), answered_by = ? WHERE id = ?",
		now,
		now,
		answeredBy,
		requestId,
	)

	return err
}

func (r *Repository) RefuseForwardingRequest(ctx context.Context, requestId string, answeredBy string) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE ForwardingRequests SET refused_at = ?, answered_by = ? WHERE id = ?",
		time.Now().UTC(),
		answeredBy,
		requestId,
	)

	return err
}

func (r *Repository) GetForwardingRequest(ctx context.Context, requestId string) (*model.ForwardingRequest, error) {
	q := "SELECT id, requester_id, requester_name, duration, created_at, accepted_at, refused_at, expires_at, answered_by FROM ForwardingRequests WHERE id = ? LIMIT 1"
	row, err := r.db.QueryContext(ctx, q, requestId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, NotFound{ID: requestId, Type: ForwardingRequestType}
	}

	var result model.ForwardingRequest

	err = row.Scan(
		&result.Id,
		&result.RequesterId,
		&result.RequesterName,
		&result.Duration,
		&result.CreatedAt,
		&result.AcceptedAt,
		&result.RefusedAt,
		&result.ExpiresAt,
		&result.AnsweredBy,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func NewForwardingRequest(requesterId string, requesterName string, duration int) *model.ForwardingRequest {
	return &model.ForwardingRequest{
		Id:            uuid.New().String(),
		RequesterId:   requesterId,
		RequesterName: requesterName,
		Duration:      duration,
		CreatedAt:     time.Now().UTC(),
	}
}
