package service

import (
	"chedul-core/internals/handlers"
	"chedul-core/internals/domain"
	"context"
	"strings"
)

type alumnoService struct {
	alumnoRepo  domain.AlumnoRepository
	carreraRepo domain.CarreraRepository
}

func NewAlumnoService(alumnoRepo domain.AlumnoRepository, carreraRepo domain.CarreraRepository) domain.AlumnoService {
	return &alumnoService{
		alumnoRepo:  alumnoRepo,
		carreraRepo: carreraRepo,
	}
}

func (s *alumnoService) GetAll(ctx context.Context) ([]domain.Alumno, error) {
	return s.alumnoRepo.GetAll(ctx)
}

func (s *alumnoService) GetByID(ctx context.Context, id int64) (*domain.Alumno, error) {
	return s.alumnoRepo.GetByID(ctx, id)
}

func (s *alumnoService) Create(ctx context.Context, req domain.AlumnoRequest) (*domain.Alumno, error) {
	if err := s.validate(req); err != nil {
		return nil, err
	}

	if existing, _ := s.alumnoRepo.GetByEmail(ctx, req.Email); existing != nil {
		return nil, handlers.InvalidJSON()
	}

	carrera, err := s.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		return nil, handlers.InvalidJSON()
	}

	alumno := &domain.Alumno{
		Nombre:  strings.TrimSpace(req.Nombre),
		Email:   strings.ToLower(strings.TrimSpace(req.Email)),
		Carrera: carrera.ID,
	}

	if err := s.alumnoRepo.Create(ctx, alumno); err != nil {
		return nil, err
	}

	return alumno, nil
}

func (s *alumnoService) Update(ctx context.Context, id int64, req domain.AlumnoRequest) (*domain.Alumno, error) {
	alumno, err := s.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validate(req); err != nil {
		return nil, err
	}

	if req.Nombre != "" {
		alumno.Nombre = strings.TrimSpace(req.Nombre)
	}

	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if existing, _ := s.alumnoRepo.GetByEmail(ctx, email); existing != nil && existing.ID != id {
			return nil, handlers.InvalidJSON()
		}
		alumno.Email = email
	}

	if req.Carrera != "" {
		carrera, err := s.carreraRepo.GetByName(ctx, req.Carrera)
		if err != nil {
			return nil, handlers.InvalidJSON()
		}
		alumno.Carrera = carrera.ID
	}

	if err := s.alumnoRepo.Update(ctx, alumno); err != nil {
		return nil, err
	}

	return alumno, nil
}

func (s *alumnoService) Delete(ctx context.Context, id int64) error {
	if _, err := s.alumnoRepo.GetByID(ctx, id); err != nil {
		return err
	}

	return s.alumnoRepo.Delete(ctx, id)
}

func (s *alumnoService) validate(req domain.AlumnoRequest) error {
	if errors := req.Validate(); len(errors) > 0 {
		return handlers.InvalidRequestData(errors)
	}

	return nil
}
