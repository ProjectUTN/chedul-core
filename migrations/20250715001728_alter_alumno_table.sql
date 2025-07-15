-- +goose Up
-- +goose StatementBegin
alter table alumno
add constraint fk_carrera
foreign key (carrera_id) references carrera(id) on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table alumno drop constraint fk_carrera
-- +goose StatementEnd
