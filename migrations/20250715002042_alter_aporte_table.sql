-- +goose Up
-- +goose StatementBegin
alter table aporte
add constraint fk_alumno
foreign key (alumno_id) references alumno(id);

alter table aporte
add constraint fk_materia
foreign key (materia_id) references materia(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
