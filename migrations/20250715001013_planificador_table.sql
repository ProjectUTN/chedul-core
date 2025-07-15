-- +goose Up
-- +goose StatementBegin
create table if not exists planificador(
    id int generated always as identity primary key not null,
    nombre text not null,
    alumno_id int not null,
    hora_inicio text not null,
    hora_fin text not null,
    dia_fin text not null,
    dia_inicio text not null,
    fecha date not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table planificador;
-- +goose StatementEnd
