-- +goose Up
-- +goose StatementBegin
alter table Horario
add constraint fk_comision
foreign key (comision_id) references Comision(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
