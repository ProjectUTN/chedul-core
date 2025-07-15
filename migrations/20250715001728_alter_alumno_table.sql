-- +goose Up
-- +goose StatementBegin
alter table Alumno
add constraint fk_carrera
foreign key (carrera_id) references Carrera(id) on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SELECT 'down SQL query';
-- +goose StatementEnd
