-- +goose Up
-- +goose StatementBegin

-- Fechas puntuales del alumno: parciales, finales, entregas, recordatorios
create table evento(
    id int generated always as identity primary key,
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int references materia(id) on delete set null,
    titulo text not null check (char_length(titulo) between 1 and 120),
    tipo text not null default 'otro'
        check (tipo in ('parcial', 'final', 'entrega', 'recordatorio', 'otro')),
    fecha date not null,
    hora time,
    descripcion text not null default '' check (char_length(descripcion) <= 1000),
    creado_en timestamptz not null default now()
);

create index evento_alumno_fecha_idx on evento(alumno_id, fecha);

-- Horario semanal de cursada del alumno. dia: 1 = lunes ... 7 = domingo
create table clase(
    id int generated always as identity primary key,
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int references materia(id) on delete set null,
    titulo text not null check (char_length(titulo) between 1 and 120),
    dia smallint not null check (dia between 1 and 7),
    hora_inicio time not null,
    hora_fin time not null,
    aula text not null default '' check (char_length(aula) <= 60),
    check (hora_fin > hora_inicio)
);

create index clase_alumno_idx on clase(alumno_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table clase;
drop table evento;
-- +goose StatementEnd
