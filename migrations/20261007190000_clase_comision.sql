-- +goose Up
-- +goose StatementBegin

-- Las clases que se agregan desde una comision (al marcar una materia como
-- cursando) guardan de cual vienen, para poder sacarlas si la materia deja de
-- estar en curso.
alter table clase add column comision_id int references comision(id) on delete set null;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table clase drop column if exists comision_id;
-- +goose StatementEnd
