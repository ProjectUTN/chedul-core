-- +goose Up
-- +goose StatementBegin
create table if not exists AporteAlumno(
    id int generated always as identity primary key not null,
    alumno_id int not null references Alumno(id) on delete cascade,
    aporte_id int not null references Aporte(id) on delete cascade,
    fecha date not null
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table AporteAlumno;
-- +goose StatementEnd
