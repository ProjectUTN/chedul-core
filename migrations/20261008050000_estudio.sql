-- Sesiones de estudio (pomodoro o cronometro) y el ranking semanal, en el
-- que solo aparecen los alumnos que se suman.
-- +goose Up
-- +goose StatementBegin
create table sesion_estudio(
    id int generated always as identity primary key,
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int references materia(id) on delete set null,
    modo text not null check (modo in ('pomodoro', 'libre')),
    minutos int not null check (minutos between 1 and 720),
    fin timestamptz not null default now()
);

create index sesion_estudio_alumno_fin_idx on sesion_estudio(alumno_id, fin);
create index sesion_estudio_fin_idx on sesion_estudio(fin);

alter table alumno add column en_ranking boolean not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table alumno drop column en_ranking;
drop table sesion_estudio;
-- +goose StatementEnd
