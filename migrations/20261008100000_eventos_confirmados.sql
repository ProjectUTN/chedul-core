-- Parciales, entregas y finales confirmados por varios alumnos: si 3 o mas
-- de la misma comision (o con la materia regular, para los finales) cargan
-- el mismo evento, se le sugiere al resto. Aca se guarda quien dijo que no
-- es asi; con tantos "no" como confirmaciones deja de sugerirse.
-- comision_id es 0 en los finales, que no dependen de la comision.
-- +goose Up
-- +goose StatementBegin
create table evento_desmentido(
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int not null references materia(id) on delete cascade,
    comision_id int not null default 0,
    tipo text not null check (tipo in ('parcial', 'entrega', 'final')),
    fecha date not null,
    creado timestamptz not null default now(),
    primary key (alumno_id, materia_id, comision_id, tipo, fecha)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table evento_desmentido;
-- +goose StatementEnd
