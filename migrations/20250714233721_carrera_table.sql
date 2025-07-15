-- +goose Up
-- +goose StatementBegin
create table if not exists carrera (
    id int generated always as identity primary key not null,
    nombre text not null unique
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table carrera;
-- +goose StatementEnd
