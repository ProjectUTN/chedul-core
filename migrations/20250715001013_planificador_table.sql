-- +goose Up
-- +goose StatementBegin
create table if not exists planificador(
    id int generated always as identity primary key not null,
    nombre text not null,
    alumno_id int not null,
    hora_inicio time not null default '06:00:00',
    hora_fin time not null default '23:59:59',
    dia_fin_id  int not null,
    dia_inicio_id int not null,
    fecha date not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table planificador;
-- +goose StatementEnd
