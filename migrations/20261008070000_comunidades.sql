-- Comunidades de la carrera (grupos de WhatsApp, Discord, Telegram...) que
-- cargan los mismos alumnos. Con 3 reportes de distintos alumnos se ocultan.
-- +goose Up
-- +goose StatementBegin
create table comunidad(
    id int generated always as identity primary key,
    alumno_id int references alumno(id) on delete set null,
    materia_id int references materia(id) on delete set null,
    nombre text not null check (char_length(nombre) between 1 and 80),
    descripcion text not null default '' check (char_length(descripcion) <= 300),
    plataforma text not null check (plataforma in ('whatsapp', 'discord', 'telegram', 'instagram', 'otra')),
    link text not null unique check (char_length(link) <= 500),
    creada timestamptz not null default now()
);

create table comunidad_reporte(
    comunidad_id int not null references comunidad(id) on delete cascade,
    alumno_id int not null references alumno(id) on delete cascade,
    creado timestamptz not null default now(),
    primary key (comunidad_id, alumno_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table comunidad_reporte;
drop table comunidad;
-- +goose StatementEnd
