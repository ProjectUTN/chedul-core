-- +goose Up
-- +goose StatementBegin
create table if not exists alumno (
    id int generated always as identity primary key not null,
    nombre text not null,
    email text unique,
    carrera_id int not null
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table alumno;
-- +goose StatementEnd
