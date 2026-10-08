-- Los administradores pasan a la variable de entorno ADMIN_EMAILS, asi sus
-- correos no quedan en el repo.
-- +goose Up
-- +goose StatementBegin
alter table alumno drop column es_admin;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table alumno add column es_admin boolean not null default false;
-- +goose StatementEnd
