-- +goose Up
-- +goose StatementBegin
alter table horario
add constraint fk_comision
foreign key (comision_id) references comision(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
