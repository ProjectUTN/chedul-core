package service

import (
	"chedul-core/internals/domain"
	"context"
)

type cuatrimestreService struct {
	cuatrimestreRepo domain.CuatrimestreRepository
}

func NewCuatrimestreService(cuatrimestreRepo domain.CuatrimestreRepository) domain.CuatrimestreService {
	return &cuatrimestreService{
		cuatrimestreRepo: cuatrimestreRepo,
	}
}

func (s *cuatrimestreService) GetAll(ctx context.Context) ([]domain.Cuatrimestre, error) {
	return s.cuatrimestreRepo.GetAll(ctx)
}

func (s *cuatrimestreService) GetByID(ctx context.Context, id int64) (*domain.Cuatrimestre, error) {
	return s.cuatrimestreRepo.GetByID(ctx, id)
}
