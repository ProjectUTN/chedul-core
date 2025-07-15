-- +goose Up
-- +goose StatementBegin
alter table Materia
add constraint fk_correlativa
foreign key (correlativa_id) references Materia(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
