package upscale

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateJob(ctx context.Context, job Job) error {
	sql, args, err := psql.Insert("jobs").
		Columns("id", "status").
		Values(job.ID, job.Status).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	sql, args, err := psql.Update("jobs").
		Set("status", status).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) GetStatus(ctx context.Context, id string) (string, error) {
	sql, args, err := psql.Select("status").
		From("jobs").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return "", err
	}

	var status string
	err = r.db.QueryRow(ctx, sql, args...).Scan(&status)
	return status, err
}
