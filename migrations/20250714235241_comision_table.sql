-- +goose Up
-- +goose StatementBegin
create table if not exists comision(
    id int generated always as identity primary key not null,
    codigo text,
    materia_id int not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table comision;
-- +goose StatementEnd
