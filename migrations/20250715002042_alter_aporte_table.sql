-- +goose Up
-- +goose StatementBegin
alter table aporte
add constraint fk_alumno
foreign key (alumno_id) references alumno(id);

alter table aporte
add constraint fk_materia
foreign key (materia_id) references materia(id);

alter table aporte
add constraint fk_tag
foreign key (tag_id) references aporte_tag(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table aporte drop constraint fk_alumno;
alter table aporte drop constraint fk_materia;
alter table aporte drop constraint fk_tag;
-- +goose StatementEnd
