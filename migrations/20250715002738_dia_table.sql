-- +goose Up
-- +goose StatementBegin
create table if not exists dias(
    id int generated always as identity primary key not null,
    nombre text not null check (char_length(nombre) <= 10)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table dias;
-- +goose StatementEnd
