-- inicio: cuando empezo de verdad una sesion de estudio (con pausas incluidas).
--   Las sesiones de antes quedan en null y se muestran como fin menos duracion.
-- +goose Up
-- +goose StatementBegin
alter table sesion_estudio add column if not exists inicio timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table sesion_estudio drop column inicio;
-- +goose StatementEnd
