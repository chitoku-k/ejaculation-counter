package client

import (
	"context"
	"fmt"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type db struct {
	Environment config.Environment
	Connection  *sqlx.DB
}

type DB interface {
	Query(ctx context.Context, q string) ([]string, error)
	UpdateCount(ctx context.Context, userID int64, date time.Time, count int) error
	Close() error
}

func NewDB(environment config.Environment) (DB, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s sslmode=%s",
		environment.DB.Username,
		environment.DB.Password,
		environment.DB.Host,
		environment.DB.Database,
		environment.DB.SSLMode,
	)
	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	conn.SetConnMaxLifetime(environment.DB.MaxLifetime)

	return &db{
		Environment: environment,
		Connection:  conn,
	}, nil
}

func (d *db) Close() error {
	return d.Connection.Close()
}

func (d *db) Query(ctx context.Context, q string) ([]string, error) {
	rows, err := d.Connection.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var r string
		rows.Scan(&r)
		result = append(result, r)
	}

	return result, nil
}

func (d *db) UpdateCount(ctx context.Context, userID int64, date time.Time, count int) error {
	var current int
	err := d.Connection.GetContext(
		ctx,
		&current,
		`SELECT COUNT(*) FROM "counts" WHERE "user" = ? AND "date" = ?`,
		userID,
		date.Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("failed to get current count: %w", err)
	}

	if current > 0 {
		_, err = d.Connection.ExecContext(
			ctx,
			`UPDATE "counts" SET "count" = ? WHERE "user" = ? AND "date" = ?`,
			count,
			userID,
			date.Format("2006-01-02"),
		)
	} else {
		_, err = d.Connection.ExecContext(
			ctx,
			`INSERT INTO "counts" ("user", "date", "count") VALUES (?, ?, ?)`,
			userID,
			date.Format("2006-01-02"),
			count,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to update count on DB: %w", err)
	}
	return nil
}
