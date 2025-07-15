-- +goose Up
-- +goose StatementBegin
create table if not exists MateriasPorCarrera(
    id int generated always as identity primary key not null,
    carrera_id int not null references Carrera(id) on delete cascade,
    materia_id int not null references Materia(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table MateriasPorCarrera;
-- +goose StatementEnd
