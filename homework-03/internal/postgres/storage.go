package postgres

import (
	"context"
	"main/internal/postgres/postgres"
)

type Interface interface {
	UpsertUrl(url string) (int64, error)
	GetUrl(id int64) (string, error)
	IncClicks(ctx context.Context, id int64, count int64) error
}

func New() (Interface, error) {
	return postgres.New()
}
