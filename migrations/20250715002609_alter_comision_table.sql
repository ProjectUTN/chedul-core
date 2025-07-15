-- +goose Up
-- +goose StatementBegin
alter table comision
add constraint fk_materia
foreign key (materia_id) references materia(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table comision drop constraint fk_materia;
-- +goose StatementEnd
