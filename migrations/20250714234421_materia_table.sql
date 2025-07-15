-- +goose Up
-- +goose StatementBegin
create table if not exists Materia (
    id int generated always as identity primary key not null,
    nombre text not null unique,
    carga_horaria int not null,
    correlativa_id int,
    nivel int,
    area text,
    tipo text,
    horas float,
    cuatrimestre text,
    bloque text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table Materia;
-- +goose StatementEnd
