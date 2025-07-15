-- +goose Up
-- +goose StatementBegin
alter table horario
add constraint fk_comision
foreign key (comision_id) references comision(id);

alter table horario
add constraint fk_dia
foreign key(dia_id) references dias(id) on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table horario drop constraint fk_comision;
alter table horario drop constraint fk_dia;
-- +goose StatementEnd
