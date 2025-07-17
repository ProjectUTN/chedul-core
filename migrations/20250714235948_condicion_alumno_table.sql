-- +goose Up
-- +goose StatementBegin
create table if not exists condicion_alumno(
    id int generated always as identity primary key not null,
    condicion_id int not null,
    nota int check (nota >= 1 and nota <= 10),
    alumno_id int not null references alumno(id) on delete cascade,
    materia_id int not null references materia(id) on delete cascade,
    unique(alumno_id, materia_id)
);

create table if not exists condicion(
    id int generated always as identity primary key not null,
    condicion text not null unique
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table condicion_alumno;
drop table condicion;
-- +goose StatementEnd
