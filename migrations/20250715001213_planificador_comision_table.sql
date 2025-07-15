-- +goose Up
-- +goose StatementBegin
create table if not exists PlanificadorComision(
    id int generated always as identity primary key not null,
    planificador_id int not null references Planificador(id) on delete cascade,
    comision_id int not null references Comision(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table PlanificadorComision;
-- +goose StatementEnd
