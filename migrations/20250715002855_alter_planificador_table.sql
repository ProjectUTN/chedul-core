-- +goose Up
-- +goose StatementBegin
alter table planificador
add constraint fk_alumno
foreign key (alumno_id) references alumno(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
