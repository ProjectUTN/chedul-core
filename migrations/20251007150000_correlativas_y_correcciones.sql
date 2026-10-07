-- +goose Up
-- +goose StatementBegin

-- El dominio acepta nombres de hasta 256 caracteres (ver domain/username.go);
-- el check original (4 a 20) rechazaba nombres validos con un error 500.
alter table alumno drop constraint if exists alumno_nombre_check;
alter table alumno add constraint alumno_nombre_check check (char_length(nombre) between 1 and 256);

-- Una materia puede tener varias correlativas, cada una con su tipo.
-- La columna correlativa_id solo permitia una.
alter table materia drop constraint if exists fk_correlativa;
alter table materia drop column if exists correlativa_id;
alter table materia add column if not exists codigo text unique;
alter table materia alter column programa set default '';

create table if not exists correlativa (
    id int generated always as identity primary key not null,
    materia_id int not null references materia(id) on delete cascade,
    requiere_id int not null references materia(id) on delete cascade,
    -- 'regular': hay que tener regularizada (o aprobada) la materia requerida
    -- 'aprobada': hay que tener aprobada la materia requerida
    tipo text not null check (tipo in ('regular', 'aprobada')),
    unique (materia_id, requiere_id),
    check (materia_id <> requiere_id)
);

alter table materiasPorcarrera add constraint materiasporcarrera_unica unique (carrera_id, materia_id);

-- Al borrar un alumno se borra todo lo suyo
alter table planificador drop constraint if exists fk_alumno;
alter table planificador add constraint fk_alumno foreign key (alumno_id) references alumno(id) on delete cascade;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table planificador drop constraint if exists fk_alumno;
alter table planificador add constraint fk_alumno foreign key (alumno_id) references alumno(id);

alter table materiasPorcarrera drop constraint if exists materiasporcarrera_unica;

drop table if exists correlativa;

alter table materia alter column programa drop default;
alter table materia drop column if exists codigo;
alter table materia add column correlativa_id int;
alter table materia add constraint fk_correlativa foreign key (correlativa_id) references materia(id);

alter table alumno drop constraint if exists alumno_nombre_check;
alter table alumno add constraint alumno_nombre_check check (char_length(nombre) between 4 and 20);
-- +goose StatementEnd
