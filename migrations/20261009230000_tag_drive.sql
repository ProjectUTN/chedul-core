-- Categoria "Drive" para los aportes que son una carpeta de Google Drive.
-- +goose Up
-- +goose StatementBegin
insert into aporte_tag (nombre) values ('Drive') on conflict (nombre) do nothing;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
delete from aporte_tag where nombre = 'Drive' and not exists (select 1 from aporte a where a.tag_id = aporte_tag.id);
-- +goose StatementEnd
