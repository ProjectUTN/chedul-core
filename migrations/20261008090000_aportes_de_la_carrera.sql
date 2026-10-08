-- Aportes de toda la carrera (planes de estudio, guias de ingreso, links
-- utiles): no son de una materia en particular.
-- +goose Up
-- +goose StatementBegin
alter table aporte alter column materia_id drop not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
delete from aporte where materia_id is null;
alter table aporte alter column materia_id set not null;
-- +goose StatementEnd
