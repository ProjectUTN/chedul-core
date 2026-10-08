-- Los bloques del horario pueden ser clases u otras actividades (trabajo,
-- gimnasio...) y pueden repetirse solo hasta una fecha.
-- +goose Up
-- +goose StatementBegin
alter table clase
    add column tipo text not null default 'clase' check (tipo in ('clase', 'trabajo', 'otro')),
    add column hasta date;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table clase drop column hasta, drop column tipo;
-- +goose StatementEnd
