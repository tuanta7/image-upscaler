package upscale

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type JobRepository struct {
	dbClient     *pgx.Conn
	queryBuilder squirrel.StatementBuilderType
}

func NewJobRepository(conn *pgx.Conn) *JobRepository {
	return &JobRepository{
		dbClient:     conn,
		queryBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *JobRepository) Create(ctx context.Context, job Job) error {
	sql, args, err := r.queryBuilder.
		Insert("jobs").
		Columns("id", "status").
		Values(job.ID, job.Status).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.dbClient.Exec(ctx, sql, args...)
	return err
}

func (r *JobRepository) GetStatus(ctx context.Context, id string) (string, error) {
	sql, args, err := r.queryBuilder.
		Select("status").
		From("jobs").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return "", err
	}

	var status string
	err = r.dbClient.QueryRow(ctx, sql, args...).Scan(&status)
	return status, err
}

func (r *JobRepository) UpdateStatus(ctx context.Context, id, status string) error {
	sql, args, err := r.queryBuilder.
		Update("jobs").
		Set("status", status).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.dbClient.Exec(ctx, sql, args...)
	return err
}
