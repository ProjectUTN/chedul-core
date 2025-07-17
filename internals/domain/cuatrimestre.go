package domain

import (
	"context"
)

type Cuatrimestre struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

type CuatrimestreRequest struct {
	Nombre string `json:"nombre"`
}

type CuatrimestreRepository interface {
	GetAll(ctx context.Context) ([]Cuatrimestre, error)
	GetByID(ctx context.Context, id int64) (*Cuatrimestre, error)
	GetByNombre(ctx context.Context, nombre string) (*Cuatrimestre, error)
}

type CuatrimestreService interface {
	GetAll(ctx context.Context) ([]Cuatrimestre, error)
	GetByID(ctx context.Context, id int64) (*Cuatrimestre, error)
}
