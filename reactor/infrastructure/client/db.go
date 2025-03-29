package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Count struct {
	UserID int64     `db:"user_id"`
	Date   time.Time `db:"date"`
	Count  int       `db:"count"`
}

type db struct {
	Connection *sqlx.DB
}

type DB interface {
	Query(ctx context.Context, q string) ([]string, int64, error)
	UpdateCount(ctx context.Context, userID int64, date time.Time, count int) error
	Close() error
}

func NewDB(params map[string]string, maxLifetime time.Duration) (DB, error) {
	var dsn strings.Builder
	for k, v := range params {
		if v == "" {
			continue
		}
		dsn.WriteString(k)
		dsn.WriteString("=")
		dsn.WriteString(v)
		dsn.WriteString(" ")
	}

	conn, err := sqlx.Connect("pgx", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	conn.SetConnMaxLifetime(maxLifetime)

	return &db{
		Connection: conn,
	}, nil
}

func (d *db) Close() error {
	return d.Connection.Close()
}

func (d *db) query(ctx context.Context, conn *pgx.Conn, q string) (result []string, affected int64, err error) {
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return result, affected, err
	}
	defer func() {
		rows.Close()
		affected = rows.CommandTag().RowsAffected()
	}()

	columns := rows.FieldDescriptions()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return result, affected, fmt.Errorf("failed to get values: %w", err)
		}

		var sb strings.Builder
		for i, v := range values {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(columns[i].Name)
			sb.WriteString(": ")
			sb.WriteString(fmt.Sprint(v))
		}

		result = append(result, sb.String())
	}

	return
}

func (d *db) Query(ctx context.Context, q string) (result []string, affected int64, err error) {
	conn, err := d.Connection.Conn(ctx)
	if err != nil {
		return nil, affected, fmt.Errorf("failed to open: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	err = conn.Raw(func(driverConn any) error {
		conn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("failed to get conn from %T", driverConn)
		}

		result, affected, err = d.query(ctx, conn.Conn(), q)
		return err
	})

	return result, affected, err
}

func (d *db) UpdateCount(ctx context.Context, userID int64, date time.Time, count int) error {
	_, err := d.Connection.NamedExecContext(
		ctx,
		`INSERT INTO "counts" ("user_id", "date", "count") VALUES (:user_id, :date, :count) ON CONFLICT ("user_id", "date") DO UPDATE SET "count" = :count`,
		Count{
			UserID: userID,
			Date:   date,
			Count:  count,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update count on DB: %w", err)
	}

	return nil
}
