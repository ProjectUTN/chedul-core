-- +goose Up
-- +goose StatementBegin

-- Un aporte puede ser un link (Drive, YouTube, etc), un archivo subido, o ambos.
alter table aporte drop column if exists puntaje;
alter table aporte add column if not exists creado_en timestamptz not null default now();
alter table aporte add column if not exists archivo_key text;
alter table aporte add column if not exists archivo_nombre text;
alter table aporte add column if not exists archivo_tipo text;
alter table aporte add column if not exists archivo_tamano bigint;
alter table aporte add constraint aporte_link_o_archivo check (link is not null or archivo_key is not null);
alter table aporte add constraint aporte_descripcion_largo check (char_length(descripcion) <= 2000);

-- Si se borra el alumno o la materia, se borran sus aportes
alter table aporte drop constraint if exists fk_alumno;
alter table aporte add constraint fk_alumno foreign key (alumno_id) references alumno(id) on delete cascade;
alter table aporte drop constraint if exists fk_materia;
alter table aporte add constraint fk_materia foreign key (materia_id) references materia(id) on delete cascade;

create index if not exists aporte_materia_idx on aporte (materia_id);
create index if not exists aporte_creado_en_idx on aporte (creado_en desc);

-- Un alumno marca un aporte como favorito una sola vez
alter table aportes_favoritos alter column fecha set default current_date;
alter table aportes_favoritos add constraint aportes_favoritos_unico unique (alumno_id, aporte_id);

alter table aporte_tag add constraint aporte_tag_nombre_unico unique (nombre);
insert into aporte_tag (nombre) values
    ('Resumen'),
    ('Apunte de clase'),
    ('Parcial'),
    ('Final'),
    ('Ejercicios'),
    ('Libro'),
    ('Video'),
    ('Otro')
on conflict (nombre) do nothing;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
delete from aportes_favoritos;
delete from aporte;
delete from aporte_tag;
alter table aporte_tag drop constraint if exists aporte_tag_nombre_unico;
alter table aportes_favoritos drop constraint if exists aportes_favoritos_unico;
alter table aportes_favoritos alter column fecha drop default;
drop index if exists aporte_creado_en_idx;
drop index if exists aporte_materia_idx;
alter table aporte drop constraint if exists fk_materia;
alter table aporte add constraint fk_materia foreign key (materia_id) references materia(id);
alter table aporte drop constraint if exists fk_alumno;
alter table aporte add constraint fk_alumno foreign key (alumno_id) references alumno(id);
alter table aporte drop constraint if exists aporte_descripcion_largo;
alter table aporte drop constraint if exists aporte_link_o_archivo;
alter table aporte drop column if exists archivo_tamano;
alter table aporte drop column if exists archivo_tipo;
alter table aporte drop column if exists archivo_nombre;
alter table aporte drop column if exists archivo_key;
alter table aporte drop column if exists creado_en;
alter table aporte add column puntaje float;
-- +goose StatementEnd
