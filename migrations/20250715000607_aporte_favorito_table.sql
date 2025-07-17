-- +goose Up
-- +goose StatementBegin
create table if not exists aportes_favoritos(
    id int generated always as identity primary key not null,
    alumno_id int not null references alumno(id) on delete cascade,
    aporte_id int not null references aporte(id) on delete cascade,
    fecha date not null
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table aportes_favoritos;
-- +goose StatementEnd
