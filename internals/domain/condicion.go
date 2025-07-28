package domain

import "context"

type Condicion struct {
	ID        int64  `json:"id"`
	Condicion string `json:"condicion"`
}

type CondicionRepository interface {
	GetAll(ctx context.Context) ([]Condicion, error)
}
