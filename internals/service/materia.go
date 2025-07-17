package service

import (
	"chedul-core/internals/domain"
	"context"
)

type materiaService struct {
	materiaRepo domain.MateriaRepository
}

func (s *materiaService) GetAll(ctx context.Context) ([]domain.Materia, error) {
	return s.materiaRepo.GetAll(ctx)
}

func (s *materiaService) GetByID(ctx context.Context, id int64) (*domain.Materia, error) {
	return s.materiaRepo.GetByID(ctx, id)
}

func NewMateriaService(materiaRepo domain.MateriaRepository) domain.MateriaService {
	return &materiaService{materiaRepo: materiaRepo}
}
