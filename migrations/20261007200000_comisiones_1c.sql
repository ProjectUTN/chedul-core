-- +goose Up
-- +goose StatementBegin

-- Horarios del 1er cuatrimestre de ISI (UTN FRRe), del PDF oficial
-- "Horario ISI 1er Cuat" (sin aulas). Las comisiones "Anual" que ya estaban
-- cargadas salieron de ese mismo PDF, asi que pasan a ser las del 1C: desde
-- que estan las del 2C, cada cuatrimestre tiene sus propias comisiones.
update comision set cuatrimestre = '1C' where cuatrimestre = 'Anual';

-- Lo que faltaba del PDF: Sistemas y Procesos de Negocio e Ingles I de cada
-- curso de 1er año, y el curso 6 (noche).
insert into comision (codigo, cuatrimestre, materia_id)
select v.codigo, '1C', m.id
from (values
    ('K1.1', 'isi-sposn'),
    ('K1.2', 'isi-sposn'),
    ('K1.3', 'isi-sposn'),
    ('K1.4', 'isi-sposn'),
    ('K1.5', 'isi-sposn'),
    ('K1.1', 'isi-ing1'),
    ('K1.2', 'isi-ing1'),
    ('K1.3', 'isi-ing1'),
    ('K1.4', 'isi-ing1'),
    ('K1.6', 'isi-f1'),
    ('K1.6', 'isi-aed'),
    ('K1.6', 'isi-am1'),
    ('K1.6', 'isi-ayga')
) as v(codigo, materia)
join materia m on m.codigo = v.materia
on conflict (materia_id, codigo, cuatrimestre) do nothing;

insert into horario (comision_id, dia_id, hora_inicio, hora_fin)
select co.id, d.id, v.hora_inicio::time, v.hora_fin::time
from (values
    ('isi-sposn', 'K1.1', 'Lunes', '07:45', '10:00'),
    ('isi-sposn', 'K1.2', 'Martes', '07:45', '10:00'),
    ('isi-sposn', 'K1.3', 'Miércoles', '13:30', '15:45'),
    ('isi-sposn', 'K1.4', 'Miércoles', '20:30', '22:45'),
    ('isi-sposn', 'K1.5', 'Miércoles', '13:30', '15:45'),
    ('isi-ing1', 'K1.1', 'Miércoles', '17:20', '20:25'),
    ('isi-ing1', 'K1.2', 'Miércoles', '17:20', '20:25'),
    ('isi-ing1', 'K1.3', 'Martes', '07:45', '10:55'),
    ('isi-ing1', 'K1.4', 'Martes', '07:45', '10:55'),
    ('isi-am1', 'K1.6', 'Martes', '20:30', '22:45'),
    ('isi-am1', 'K1.6', 'Miércoles', '19:40', '21:15'),
    ('isi-ayga', 'K1.6', 'Miércoles', '21:15', '22:45'),
    ('isi-ayga', 'K1.6', 'Viernes', '20:30', '22:45'),
    ('isi-f1', 'K1.6', 'Jueves', '18:10', '21:15'),
    ('isi-aed', 'K1.6', 'Sábado', '08:00', '12:00')
) as v(materia, codigo, dia, hora_inicio, hora_fin)
join materia m on m.codigo = v.materia
join comision co on co.materia_id = m.id and co.codigo = v.codigo and co.cuatrimestre = '1C'
join dias d on d.nombre = v.dia
where not exists (select 1 from horario h where h.comision_id = co.id and h.dia_id = d.id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
delete from comision co
using materia m
where co.materia_id = m.id and co.cuatrimestre = '1C'
  and (m.codigo, co.codigo) in (
    ('isi-sposn', 'K1.1'), ('isi-sposn', 'K1.2'), ('isi-sposn', 'K1.3'), ('isi-sposn', 'K1.4'), ('isi-sposn', 'K1.5'),
    ('isi-ing1', 'K1.1'), ('isi-ing1', 'K1.2'), ('isi-ing1', 'K1.3'), ('isi-ing1', 'K1.4'),
    ('isi-f1', 'K1.6'), ('isi-aed', 'K1.6'), ('isi-am1', 'K1.6'), ('isi-ayga', 'K1.6')
  );
-- +goose StatementEnd
