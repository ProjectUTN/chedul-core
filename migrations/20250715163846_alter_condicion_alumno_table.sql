-- +goose Up
-- +goose StatementBegin
alter table condicion_alumno
add constraint fk_estado
foreign key (condicion_id) references condicion(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table condicion_alumno drop constraint fk_estado;
-- +goose StatementEnd
