package service

import (
	"chedul-core/internals/domain"
	"context"
)

type carreraService struct {
	carreraRepo domain.CarreraRepository
}

func NewCarreraService(carreraRepo domain.CarreraRepository) domain.CarreraService {
	return &carreraService{
		carreraRepo: carreraRepo,
	}
}

func (s *carreraService) GetAll(ctx context.Context) ([]domain.Carrera, error) {
	return s.carreraRepo.GetAll(ctx)
}

func (s *carreraService) GetByID(ctx context.Context, id int64) (*domain.Carrera, error) {
	return s.carreraRepo.GetByID(ctx, id)
}
