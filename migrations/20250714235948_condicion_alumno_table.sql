-- +goose Up
-- +goose StatementBegin
create table if not exists condicion_alumno(
    id int generated always as identity primary key not null,
    condicion text,
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int not null references materia(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table condicion_alumno;
-- +goose StatementEnd
