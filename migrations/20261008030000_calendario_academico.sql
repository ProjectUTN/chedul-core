-- Calendario academico de la FRRe (grado, tecnicaturas y ciclos de
-- licenciatura), el mismo para todos los alumnos. Cada año se agrega el nuevo
-- en otra migracion; el 2026 sale de frre.utn.edu.ar/calendario/calendario.
-- +goose Up
-- +goose StatementBegin
create table fecha_academica(
    id int generated always as identity primary key,
    titulo text not null check (char_length(titulo) between 1 and 120),
    tipo text not null check (tipo in ('examen', 'cuatrimestre', 'feriado', 'receso', 'otro')),
    desde date not null,
    hasta date not null,
    check (hasta >= desde),
    unique (titulo, desde)
);

create index fecha_academica_desde_hasta_idx on fecha_academica(desde, hasta);

insert into fecha_academica (titulo, tipo, desde, hasta) values
    ('Inicio Primer Cuatrimestre', 'cuatrimestre', '2026-03-16', '2026-03-16'),
    ('Día No Laborable', 'feriado', '2026-03-23', '2026-03-23'),
    ('Feriado', 'feriado', '2026-03-24', '2026-03-24'),
    ('Feriado', 'feriado', '2026-04-02', '2026-04-03'),
    ('Exámenes Finales 1° Llamado', 'examen', '2026-04-13', '2026-04-18'),
    ('Feriado', 'feriado', '2026-05-01', '2026-05-01'),
    ('Asueto Docente', 'feriado', '2026-05-02', '2026-05-02'),
    ('Feriado', 'feriado', '2026-05-25', '2026-05-25'),
    ('Exámenes Finales 2° Llamado', 'examen', '2026-06-10', '2026-06-10'),
    ('Feriado', 'feriado', '2026-06-15', '2026-06-15'),
    ('Feriado', 'feriado', '2026-06-20', '2026-06-20'),
    ('Feriado', 'feriado', '2026-07-09', '2026-07-09'),
    ('Día No Laborable', 'feriado', '2026-07-10', '2026-07-10'),
    ('Recuperatorios y Fin Primer Cuatrimestre', 'cuatrimestre', '2026-07-13', '2026-07-18'),
    ('Receso invernal', 'receso', '2026-07-20', '2026-08-01'),
    ('Exámenes Finales 3° Llamado', 'examen', '2026-08-03', '2026-08-08'),
    ('Inicio Segundo Cuatrimestre', 'cuatrimestre', '2026-08-10', '2026-08-10'),
    ('Feriado', 'feriado', '2026-08-17', '2026-08-17'),
    ('Día de la UTN', 'otro', '2026-08-19', '2026-08-19'),
    ('Feriado', 'feriado', '2026-08-27', '2026-08-27'),
    ('Exámenes Finales 4° Llamado', 'examen', '2026-09-07', '2026-09-12'),
    ('Asueto Estudiantil', 'feriado', '2026-09-21', '2026-09-21'),
    ('Feriado', 'feriado', '2026-10-12', '2026-10-12'),
    ('Exámenes Finales 5° Llamado', 'examen', '2026-10-20', '2026-10-20'),
    ('Feriado', 'feriado', '2026-11-23', '2026-11-23'),
    ('Asueto No Docente', 'feriado', '2026-11-26', '2026-11-26'),
    ('Fin Segundo Cuatrimestre', 'cuatrimestre', '2026-12-05', '2026-12-05'),
    ('Día No Laborable', 'feriado', '2026-12-07', '2026-12-07'),
    ('Feriado', 'feriado', '2026-12-08', '2026-12-08'),
    ('Exámenes Finales 6° Llamado', 'examen', '2026-12-09', '2026-12-12'),
    ('Exámenes Finales 7° Llamado', 'examen', '2026-12-14', '2026-12-19'),
    ('Feriado', 'feriado', '2026-12-25', '2026-12-25'),
    ('Exámenes Finales 8° Llamado', 'examen', '2027-02-15', '2027-02-20'),
    ('Exámenes Finales 9° Llamado', 'examen', '2027-02-22', '2027-02-27'),
    ('Exámenes Finales 10° Llamado', 'examen', '2027-03-01', '2027-03-06');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table fecha_academica;
-- +goose StatementEnd
