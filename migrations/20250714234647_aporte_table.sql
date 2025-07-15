-- +goose Up
-- +goose StatementBegin
create table if not exists aporte (
    id int generated always as identity primary key not null,
    titulo text not null,
    alumno_id int not null,
    materia_id int not null,
    descripcion text,
    tag text,
    link text,
    puntaje float
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table aporte;
-- +goose StatementEnd
