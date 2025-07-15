-- +goose Up
-- +goose StatementBegin
alter table Planificador
add constraint fk_alumno
foreign key (alumno_id) references Alumno(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
