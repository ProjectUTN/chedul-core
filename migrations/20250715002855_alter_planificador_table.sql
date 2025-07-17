-- +goose Up
-- +goose StatementBegin
alter table planificador
add constraint fk_alumno
foreign key (alumno_id) references alumno(id);

alter table planificador
add constraint fk_dia_inicio
foreign key (dia_inicio_id) references dias(id) on delete cascade;

alter table planificador
add constraint fk_dia_fin
foreign key (dia_fin_id) references dias(id) on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table planificador drop constraint fk_alumno;
alter table planificador drop constraint fk_dia_inicio;
alter table planificador drop constraint fk_dia_fin;
-- +goose StatementEnd
