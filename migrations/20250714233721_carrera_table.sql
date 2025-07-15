-- +goose Up
-- +goose StatementBegin
create table if not exists carrera (
    id int generated always as identity primary key not null,
    nombre text not null unique check (char_length(nombre) <= 100) 
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table carrera;
-- +goose StatementEnd
