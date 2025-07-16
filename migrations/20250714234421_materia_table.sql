-- +goose Up
-- +goose StatementBegin
create table if not exists materia (
    id int generated always as identity primary key not null,
    nombre text not null unique,
    carga_horaria int not null,
    correlativa_id int not null,
    nivel int not null check(nivel between 1 and 5),
    area text not null,
    tipo text not null,
    cuatrimestre_id int,
    horas float,
    bloque text not null,
    programa text not null
    check(correlativa_id != id)
);

create table if not exists cuatrimestre(
    id int generated always as identity primary key not null,
    nombre text not null unique
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table materia;
drop table cuatrimestre;
-- +goose StatementEnd
