-- Datos para el panel de administracion y para protegerse de cuentas truchas:
--   creado: cuando se registro (las cuentas de antes quedan en null = antiguas)
--   ultimo_acceso: ultima vez que uso la app
--   google_vinculado: ya entro alguna vez con Google
--   version_sesion: se sube al cambiar la clave; invalida los refresh tokens viejos
--   es_admin: puede ver el panel de administracion
-- acceso guarda cada inicio de sesion.
-- +goose Up
-- +goose StatementBegin
alter table alumno
    add column creado timestamptz,
    add column ultimo_acceso timestamptz,
    add column google_vinculado boolean not null default false,
    add column version_sesion int not null default 0,
    add column es_admin boolean not null default false;
alter table alumno alter column creado set default now();

create table acceso(
    id int generated always as identity primary key,
    alumno_id int not null references alumno(id) on delete cascade,
    fecha timestamptz not null default now(),
    metodo text not null
);
create index acceso_fecha_idx on acceso(fecha desc);

-- Eduardo, creador de Chedul
update alumno set es_admin = true where email = 'edu.ramirez645@gmail.com';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table acceso;
alter table alumno
    drop column creado,
    drop column ultimo_acceso,
    drop column google_vinculado,
    drop column version_sesion,
    drop column es_admin;
-- +goose StatementEnd
