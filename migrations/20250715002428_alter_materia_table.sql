-- +goose Up
-- +goose StatementBegin
alter table materia
add constraint fk_correlativa
foreign key (correlativa_id) references materia(id);

alter table materia
add constraint fk_cuatrimestre
foreign key (cuatrimestre_id) references cuatrimestre(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table materia drop constraint fk_correlativa;
alter table materia drop constraint fk_cuatrimestre;
-- +goose StatementEnd
