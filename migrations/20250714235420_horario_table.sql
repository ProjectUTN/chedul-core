-- +goose Up
-- +goose StatementBegin
create table if not exists horario(
    id int generated always as identity primary key not null,
    dia text,
    horainicio text,
    horafin text,
    comision_id int not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table horario;
-- +goose StatementEnd
