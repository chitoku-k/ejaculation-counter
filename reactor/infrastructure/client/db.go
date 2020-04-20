package client

import (
	"context"
	"fmt"
	"time"

	"github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type db struct {
	ctx         context.Context
	Environment config.Environment
	Connection  *sqlx.DB
}

type DB interface {
	Query(q string) ([]string, error)
	UpdateCount(userID int64, date time.Time, count int) error
}

func NewDB(ctx context.Context, environment config.Environment) (DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		environment.DB.Username,
		environment.DB.Password,
		environment.DB.Host,
		environment.DB.Database,
	)
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to DB")
	}

	return &db{
		ctx:         ctx,
		Environment: environment,
		Connection:  conn,
	}, nil
}

func (d *db) Query(q string) ([]string, error) {
	rows, err := d.Connection.QueryContext(d.ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
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

func (d *db) UpdateCount(userID int64, date time.Time, count int) error {
	var current int
	err := d.Connection.GetContext(
		d.ctx,
		&current,
		"SELECT COUNT(*) FROM `counts` WHERE `user` = ? AND `date` = ?",
		userID,
		date.Format("2006-01-02"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to get current count")
	}

	if current > 0 {
		_, err = d.Connection.ExecContext(
			d.ctx,
			"UPDATE `counts` SET `count` = ? WHERE `user` = ? AND `date` = ?",
			count,
			userID,
			date.Format("2006-01-02"),
		)
	} else {
		_, err = d.Connection.ExecContext(
			d.ctx,
			"INSERT INTO `counts` (`user`, `date`, `count`) VALUES (?, ?, ?)",
			userID,
			date.Format("2006-01-02"),
			count,
		)
	}
	return errors.Wrap(err, "failed to update count on DB")
}
