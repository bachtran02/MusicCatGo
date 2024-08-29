package musicbot

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(cfg DBConfig) (*DB, error) {

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := DB{}

	db.Pool, err = pgxpool.New(
		ctx,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			cfg.Host,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		))
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (d *DB) Close() {
	d.Pool.Close()
}
