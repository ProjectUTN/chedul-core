-- +goose Up
-- +goose StatementBegin
create table if not exists planificador_comision(
    id int generated always as identity primary key not null,
    planificador_id int not null references planificador(id) on delete cascade,
    comision_id int not null references comision(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table planificador_comision;
-- +goose StatementEnd
