-- +goose Up
-- +goose StatementBegin
alter table materia
add constraint fk_correlativa
foreign key (correlativa_id) references materia(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
