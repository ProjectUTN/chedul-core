-- Token secreto del link de calendario (.ics) para suscribirse desde Google
-- Calendar u otra app. Se crea la primera vez que el alumno pide el link.
-- +goose Up
-- +goose StatementBegin
alter table alumno add column calendario_token text unique;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table alumno drop column calendario_token;
-- +goose StatementEnd
