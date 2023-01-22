package postgres

import (
	"context"
	"database/sql"
	"log"
	"main/internal/config"
	"main/internal/postgres/consts"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type storage struct {
	db *sql.DB
}

func New() (*storage, error) {
	db, err := sql.Open("pgx", config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &storage{
		db: db,
	}, nil
}

func (s *storage) UpsertUrl(url string) (int64, error) {
	log.Printf("UpsertUrl %v\n", url)

	rows, err := s.db.Query(`
		WITH ins AS (
		INSERT INTO urls(url) VALUES($1) 
		ON CONFLICT DO NOTHING 
		RETURNING urls.id 
		) 
		SELECT id FROM urls WHERE url = $1 
		UNION ALL 
		SELECT id FROM ins;`, url)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, consts.InternalDBError
	}

	var id int64
	err = rows.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *storage) GetUrl(id int64) (string, error) {
	log.Printf("GetUrl %v\n", id)

	rows, err := s.db.Query("SELECT url FROM urls WHERE id = $1;", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", consts.ErrNotFound
	}

	var url string
	err = rows.Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *storage) IncClicks(ctx context.Context, id int64, count int64) error {
	log.Printf("IncClicks %v %v\n", id, count)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := s.db.Query("SELECT clicks FROM urls WHERE id = $1;", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return consts.ErrNotFound
	}

	var clicks int64
	err = rows.Scan(&clicks)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE urls SET clicks = $1 WHERE id = $2;`, clicks+count, id)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return err
}
