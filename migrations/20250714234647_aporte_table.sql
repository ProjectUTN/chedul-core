-- +goose Up
-- +goose StatementBegin
create table if not exists aporte(
    id int generated always as identity primary key not null,
    materia_id int not null,
    alumno_id int not null,
    titulo text not null check (char_length(titulo) <= 100),
    descripcion text,
    fecha date not null default current_date,
    tag_id int not null, 
    link text,
    puntaje float
);

create table if not exists aporte_tag(
    id int generated always as identity primary key not null,
    nombre text not null check (char_length(nombre) <= 20)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table aporte;
drop table aporte_tag;
-- +goose StatementEnd
