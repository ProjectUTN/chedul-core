-- +goose Up
-- +goose StatementBegin
create table if not exists CondicionAlumno(
    id int generated always as identity primary key not null,
    condicion text,
    alumno_id int not null references Alumno(id) on delete cascade,
    materia_id int not null references Materia(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table CondicionAlumno;
-- +goose StatementEnd
