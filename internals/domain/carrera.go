package domain

import "context"

type Carrera struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

type CarreraRepository interface {
	GetAll(ctx context.Context) ([]Carrera, error)
	GetByID(ctx context.Context, id int64) (*Carrera, error)
	GetByName(ctx context.Context, name string) (*Carrera, error)
}
