-- "No me interesa" para electivas que el alumno no piensa cursar, meta diaria
-- de estudio y lista de tareas.
-- +goose Up
-- +goose StatementBegin
insert into condicion (condicion) values ('No me interesa') on conflict (condicion) do nothing;

alter table alumno add column meta_diaria_minutos int not null default 120
    check (meta_diaria_minutos between 15 and 720);

create table tarea_estudio(
    id int generated always as identity primary key,
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int references materia(id) on delete set null,
    titulo text not null check (char_length(titulo) between 1 and 200),
    hecha boolean not null default false,
    creada timestamptz not null default now()
);

create index tarea_estudio_alumno_idx on tarea_estudio(alumno_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table tarea_estudio;
alter table alumno drop column meta_diaria_minutos;
delete from condicion_alumno where condicion_id = (select id from condicion where condicion = 'No me interesa');
delete from condicion where condicion = 'No me interesa';
-- +goose StatementEnd
