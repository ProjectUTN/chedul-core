-- +goose Up
-- +goose StatementBegin
alter table Aporte
add constraint fk_alumno
foreign key (alumno_id) references Alumno(id);

alter table Aporte
add constraint fk_materia
foreign key (materia_id) references Materia(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
