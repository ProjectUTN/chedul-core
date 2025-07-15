-- +goose Up
-- +goose StatementBegin
create table if not exists materiasPorcarrera(
    id int generated always as identity primary key not null,
    carrera_id int not null references carrera(id) on delete cascade,
    materia_id int not null references materia(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table materiasPorcarrera;
-- +goose StatementEnd
