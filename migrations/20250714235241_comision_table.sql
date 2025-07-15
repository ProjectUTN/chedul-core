-- +goose Up
-- +goose StatementBegin
create table if not exists Comision(
    id int generated always as identity primary key not null,
    codigo text,
    materia_id int not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table Comision;
-- +goose StatementEnd
