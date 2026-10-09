-- apellido: los aportes y el ranking muestran nombre e inicial del apellido.
--   Las cuentas de antes quedan con apellido vacio.
-- temporizador: el cronometro de estudio en curso de cada alumno, para que siga
--   igual al cambiar de dispositivo o cerrar la pantalla. rev sube con cada cambio
--   y sirve para que dos dispositivos no pisen el mismo fin de pomodoro.
-- +goose Up
-- +goose StatementBegin
alter table alumno add column if not exists apellido text not null default '';

create table if not exists temporizador(
    alumno_id int primary key references alumno(id) on delete cascade,
    estado jsonb not null,
    rev int not null default 1,
    actualizado timestamptz not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table temporizador;
alter table alumno drop column apellido;
-- +goose StatementEnd
