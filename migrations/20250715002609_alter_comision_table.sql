-- +goose Up
-- +goose StatementBegin
alter table Comision
add constraint fk_materia
foreign key (materia_id) references Materia(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
