-- +goose Up
-- +goose StatementBegin
create table if not exists Carrera (
    id int generated always as identity primary key not null,
    nombre text not null unique
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table Carrera;
-- +goose StatementEnd
