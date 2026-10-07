-- +goose Up
-- +goose StatementBegin

-- Horarios del 2do cuatrimestre 2026 de ISI (UTN FRRe), copiados del PDF
-- oficial de horarios. Una comision puede tener la misma codigo en 1C y 2C,
-- por eso la unicidad pasa a incluir el cuatrimestre. Cada horario guarda su aula.

alter table horario add column if not exists aula text not null default '';
alter table comision drop constraint if exists comision_unica;
alter table comision add constraint comision_unica unique (materia_id, codigo, cuatrimestre);

insert into comision (codigo, cuatrimestre, materia_id)
select v.codigo, '2C', m.id
from (values
    ('K5.1', 'isi-aaari'),
    ('K4.1', 'isi-aacs'),
    ('K4.1', 'isi-admsi'),
    ('K2.1', 'isi-ads'),
    ('K2.2', 'isi-ads'),
    ('K1.1', 'isi-aed'),
    ('K1.2', 'isi-aed'),
    ('K1.3', 'isi-aed'),
    ('K1.4', 'isi-aed'),
    ('K4.1', 'isi-agavz'),
    ('K5.1', 'isi-aif'),
    ('K1.1', 'isi-am1'),
    ('K1.2', 'isi-am1'),
    ('K1.3', 'isi-am1'),
    ('K1.4', 'isi-am1'),
    ('K2.1', 'isi-am2'),
    ('K2.2', 'isi-am2'),
    ('K3.1', 'isi-an'),
    ('K1.1', 'isi-ayga'),
    ('K1.2', 'isi-ayga'),
    ('K1.3', 'isi-ayga'),
    ('K1.4', 'isi-ayga'),
    ('K5.1', 'isi-cdd'),
    ('K5.1', 'isi-devops'),
    ('K3.1', 'isi-dsi'),
    ('K3.1', 'isi-dsw'),
    ('K1.1', 'isi-f1'),
    ('K1.2', 'isi-f1'),
    ('K1.3', 'isi-f1'),
    ('K1.4', 'isi-f1'),
    ('K5.1', 'isi-fcs'),
    ('K4.1', 'isi-gis'),
    ('K5.1', 'isi-ia'),
    ('K2.1', 'isi-ing2'),
    ('K2.2', 'isi-ing2'),
    ('K2.1', 'isi-pdp'),
    ('K2.2', 'isi-pdp'),
    ('K5.1', 'isi-pf'),
    ('K3.1', 'isi-planif'),
    ('K5.1', 'isi-ps'),
    ('K3.1 C1', 'isi-pye'),
    ('K3.1 C2', 'isi-pye'),
    ('K4.1', 'isi-rdd'),
    ('K3.1', 'isi-sgbd'),
    ('K2.1', 'isi-so'),
    ('K2.2', 'isi-so'),
    ('K1.1', 'isi-sposn'),
    ('K1.2', 'isi-sposn'),
    ('K1.3', 'isi-sposn'),
    ('K1.4', 'isi-sposn'),
    ('K4.1', 'isi-tpla')
) as v(codigo, materia)
join materia m on m.codigo = v.materia
on conflict (materia_id, codigo, cuatrimestre) do nothing;

-- Las comisiones 2C que ya existian (Sistemas y Procesos de Negocio) traian
-- horarios viejos: se reemplazan por los del PDF.
delete from horario h
using comision co, materia m
where h.comision_id = co.id and co.materia_id = m.id and co.cuatrimestre = '2C'
  and (m.codigo, co.codigo) in (
    ('isi-aaari', 'K5.1'),
    ('isi-aacs', 'K4.1'),
    ('isi-admsi', 'K4.1'),
    ('isi-ads', 'K2.1'),
    ('isi-ads', 'K2.2'),
    ('isi-aed', 'K1.1'),
    ('isi-aed', 'K1.2'),
    ('isi-aed', 'K1.3'),
    ('isi-aed', 'K1.4'),
    ('isi-agavz', 'K4.1'),
    ('isi-aif', 'K5.1'),
    ('isi-am1', 'K1.1'),
    ('isi-am1', 'K1.2'),
    ('isi-am1', 'K1.3'),
    ('isi-am1', 'K1.4'),
    ('isi-am2', 'K2.1'),
    ('isi-am2', 'K2.2'),
    ('isi-an', 'K3.1'),
    ('isi-ayga', 'K1.1'),
    ('isi-ayga', 'K1.2'),
    ('isi-ayga', 'K1.3'),
    ('isi-ayga', 'K1.4'),
    ('isi-cdd', 'K5.1'),
    ('isi-devops', 'K5.1'),
    ('isi-dsi', 'K3.1'),
    ('isi-dsw', 'K3.1'),
    ('isi-f1', 'K1.1'),
    ('isi-f1', 'K1.2'),
    ('isi-f1', 'K1.3'),
    ('isi-f1', 'K1.4'),
    ('isi-fcs', 'K5.1'),
    ('isi-gis', 'K4.1'),
    ('isi-ia', 'K5.1'),
    ('isi-ing2', 'K2.1'),
    ('isi-ing2', 'K2.2'),
    ('isi-pdp', 'K2.1'),
    ('isi-pdp', 'K2.2'),
    ('isi-pf', 'K5.1'),
    ('isi-planif', 'K3.1'),
    ('isi-ps', 'K5.1'),
    ('isi-pye', 'K3.1 C1'),
    ('isi-pye', 'K3.1 C2'),
    ('isi-rdd', 'K4.1'),
    ('isi-sgbd', 'K3.1'),
    ('isi-so', 'K2.1'),
    ('isi-so', 'K2.2'),
    ('isi-sposn', 'K1.1'),
    ('isi-sposn', 'K1.2'),
    ('isi-sposn', 'K1.3'),
    ('isi-sposn', 'K1.4'),
    ('isi-tpla', 'K4.1')
  );

insert into horario (comision_id, dia_id, hora_inicio, hora_fin, aula)
select co.id, d.id, v.hora_inicio::time, v.hora_fin::time, v.aula
from (values
    ('isi-sposn', 'K1.1', 'Lunes', '07:45', '10:00', 'Aula 2.10'),
    ('isi-am1', 'K1.1', 'Lunes', '10:10', '12:25', 'Aula 2.10'),
    ('isi-am1', 'K1.1', 'Martes', '07:45', '09:15', 'Aula 2.10'),
    ('isi-ayga', 'K1.1', 'Martes', '09:15', '10:55', 'Aula 2.10'),
    ('isi-aed', 'K1.1', 'Martes', '10:55', '12:25', 'Aula 2.10'),
    ('isi-f1', 'K1.1', 'Miércoles', '07:45', '10:55', 'Aula 2.10'),
    ('isi-ayga', 'K1.1', 'Jueves', '07:45', '10:00', 'Aula 2.10'),
    ('isi-aed', 'K1.1', 'Viernes', '10:10', '12:25', 'Aula 2.10'),
    ('isi-f1', 'K1.2', 'Lunes', '07:45', '10:55', 'Aula 2.9'),
    ('isi-aed', 'K1.2', 'Lunes', '10:55', '12:25', 'Aula 2.9'),
    ('isi-sposn', 'K1.2', 'Martes', '07:45', '10:00', 'Aula 2.9'),
    ('isi-ayga', 'K1.2', 'Martes', '10:10', '11:40', 'Aula 2.9'),
    ('isi-ayga', 'K1.2', 'Miércoles', '07:45', '10:00', 'Aula 2.9'),
    ('isi-am1', 'K1.2', 'Miércoles', '10:10', '12:25', 'Aula 2.9'),
    ('isi-am1', 'K1.2', 'Jueves', '07:45', '09:15', 'Aula 2.9'),
    ('isi-aed', 'K1.2', 'Viernes', '07:45', '10:00', 'Aula 2.9'),
    ('isi-f1', 'K1.3', 'Miércoles', '10:10', '13:10', 'Aula 2.2'),
    ('isi-am1', 'K1.3', 'Lunes', '13:30', '17:20', 'Aula 2.9'),
    ('isi-ayga', 'K1.3', 'Martes', '14:15', '18:05', 'Aula 2.9'),
    ('isi-sposn', 'K1.3', 'Miércoles', '13:30', '15:45', 'Aula 2.9 + 2.2'),
    ('isi-aed', 'K1.3', 'Miércoles', '15:50', '18:05', 'Aula 2.9'),
    ('isi-aed', 'K1.3', 'Viernes', '16:35', '18:05', 'Aula 2.9'),
    ('isi-f1', 'K1.4', 'Jueves', '15:00', '18:05', 'Aula 2.5'),
    ('isi-am1', 'K1.4', 'Martes', '18:10', '22:00', 'Aula 2.9'),
    ('isi-aed', 'K1.4', 'Miércoles', '18:10', '20:25', 'Aula 2.9'),
    ('isi-sposn', 'K1.4', 'Miércoles', '20:30', '22:45', 'Aula 2.9'),
    ('isi-ayga', 'K1.4', 'Jueves', '18:10', '19:40', 'Aula 2.9'),
    ('isi-ayga', 'K1.4', 'Jueves', '19:40', '22:00', 'Aula 2.9'),
    ('isi-aed', 'K1.4', 'Lunes', '20:30', '22:00', 'Aula 2.9'),
    ('isi-ing2', 'K2.1', 'Lunes', '13:30', '16:35', 'Virtual'),
    ('isi-am2', 'K2.1', 'Lunes', '16:35', '18:05', 'Aula 2.10'),
    ('isi-am2', 'K2.1', 'Martes', '12:45', '15:00', 'Aula 2.10'),
    ('isi-so', 'K2.1', 'Martes', '15:00', '18:05', 'Aula 2.10'),
    ('isi-pdp', 'K2.1', 'Miércoles', '13:30', '15:45', 'Aula 2.10'),
    ('isi-ads', 'K2.1', 'Miércoles', '15:50', '18:05', 'Aula 2.10 + 2.2'),
    ('isi-pdp', 'K2.1', 'Jueves', '15:00', '18:55', 'Aula 2.10'),
    ('isi-so', 'K2.1', 'Viernes', '12:00', '15:00', 'Aula 2.1'),
    ('isi-ads', 'K2.1', 'Viernes', '15:00', '17:20', 'Aula 2.10'),
    ('isi-ing2', 'K2.2', 'Lunes', '13:30', '16:35', 'Virtual'),
    ('isi-am2', 'K2.2', 'Lunes', '18:10', '20:25', 'Aula 2.10'),
    ('isi-pdp', 'K2.2', 'Lunes', '20:30', '22:45', 'Aula 2.10'),
    ('isi-am2', 'K2.2', 'Martes', '18:10', '19:40', 'Aula 2.10'),
    ('isi-so', 'K2.2', 'Martes', '19:40', '22:45', 'Aula 2.10'),
    ('isi-ads', 'K2.2', 'Miércoles', '18:10', '20:25', 'Aula 2.10'),
    ('isi-pdp', 'K2.2', 'Jueves', '18:55', '22:45', 'Aula 2.10'),
    ('isi-ads', 'K2.2', 'Viernes', '17:20', '19:40', 'Aula 2.10'),
    ('isi-so', 'K2.2', 'Viernes', '19:40', '22:45', 'Aula 2.10'),
    ('isi-an', 'K3.1', 'Lunes', '15:50', '18:05', 'Aula 2.1'),
    ('isi-an', 'K3.1', 'Lunes', '20:30', '22:45', 'Aula 2.1'),
    ('isi-pye', 'K3.1 C1', 'Martes', '15:50', '18:05', 'Aula 2.1'),
    ('isi-pye', 'K3.1 C1', 'Viernes', '15:50', '18:05', 'Aula 2.1'),
    ('isi-pye', 'K3.1 C2', 'Lunes', '18:10', '20:25', 'Aula 2.1'),
    ('isi-pye', 'K3.1 C2', 'Viernes', '18:10', '20:25', 'Aula 2.1'),
    ('isi-sgbd', 'K3.1', 'Martes', '18:10', '22:45', 'Lab AI 5-6'),
    ('isi-planif', 'K3.1', 'Martes', '18:10', '22:45', 'Aula 2.1'),
    ('isi-dsi', 'K3.1', 'Miércoles', '18:10', '20:25', 'Aula Magna'),
    ('isi-dsi', 'K3.1', 'Miércoles', '20:30', '22:45', 'Aula 2.1 + 2.10'),
    ('isi-dsw', 'K3.1', 'Jueves', '16:35', '22:45', 'Aula 2.1'),
    ('isi-aacs', 'K4.1', 'Lunes', '15:50', '18:05', 'Aula 2.2'),
    ('isi-tpla', 'K4.1', 'Lunes', '18:10', '22:45', 'Aula 2.2'),
    ('isi-rdd', 'K4.1', 'Martes', '16:35', '18:05', 'Aula 2.2'),
    ('isi-rdd', 'K4.1', 'Martes', '18:10', '20:25', 'Aula 2.2'),
    ('isi-gis', 'K4.1', 'Martes', '20:30', '22:45', 'Aula 2.2'),
    ('isi-admsi', 'K4.1', 'Miércoles', '18:10', '22:45', 'Aula 2.2'),
    ('isi-gis', 'K4.1', 'Jueves', '18:10', '20:25', 'Aula 1.1'),
    ('isi-aacs', 'K4.1', 'Jueves', '18:10', '20:25', 'Aula 2.2'),
    ('isi-rdd', 'K4.1', 'Jueves', '20:30', '22:45', 'Aula 2.2'),
    ('isi-agavz', 'K4.1', 'Viernes', '18:10', '22:45', 'Aula 2.2'),
    ('isi-ps', 'K5.1', 'Lunes', '15:50', '18:05', 'Aula 1.2'),
    ('isi-aif', 'K5.1', 'Lunes', '18:10', '22:45', 'Aula 2.8'),
    ('isi-devops', 'K5.1', 'Lunes', '18:10', '22:45', 'Aula 1.4'),
    ('isi-pf', 'K5.1', 'Martes', '15:50', '18:05', 'Aula 1.4'),
    ('isi-cdd', 'K5.1', 'Martes', '18:10', '22:45', 'Aula 1.4'),
    ('isi-aif', 'K5.1', 'Miércoles', '16:35', '18:05', 'Aula 1.4'),
    ('isi-ia', 'K5.1', 'Miércoles', '18:10', '22:45', 'Aula 1.4'),
    ('isi-pf', 'K5.1', 'Jueves', '16:35', '18:55', 'Aula 1.4'),
    ('isi-fcs', 'K5.1', 'Jueves', '18:55', '21:10', 'Aula 1.4'),
    ('isi-aaari', 'K5.1', 'Jueves', '21:10', '22:45', 'Lab CISCO'),
    ('isi-devops', 'K5.1', 'Viernes', '17:20', '18:55', 'Aula 1.4'),
    ('isi-fcs', 'K5.1', 'Viernes', '18:55', '21:15', 'Aula 1.4')
) as v(materia, codigo, dia, hora_inicio, hora_fin, aula)
join materia m on m.codigo = v.materia
join comision co on co.materia_id = m.id and co.codigo = v.codigo and co.cuatrimestre = '2C'
join dias d on d.nombre = v.dia;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Se sacan las comisiones 2C que repiten codigo con otro cuatrimestre, para
-- poder volver a la unicidad vieja.
delete from comision co
where co.cuatrimestre = '2C'
  and exists (select 1 from comision o where o.materia_id = co.materia_id and o.codigo = co.codigo and o.id <> co.id);
alter table comision drop constraint if exists comision_unica;
alter table comision add constraint comision_unica unique (materia_id, codigo);
alter table horario drop column if exists aula;
-- +goose StatementEnd
