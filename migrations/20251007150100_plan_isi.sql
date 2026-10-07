-- Plan de estudios de Ingeniería en Sistemas de Información (UTN),
-- exportado de la base de la demo en Django (project-utn-demo/backend/db.sqlite3).
-- +goose Up
-- +goose StatementBegin
insert into carrera (nombre) values ('Ingeniería en Sistemas de Información') on conflict (nombre) do nothing;

insert into cuatrimestre (nombre) values ('1C'), ('2C'), ('Anual') on conflict (nombre) do nothing;

insert into condicion (condicion) values ('Cursando'), ('Regularizada'), ('Aprobada') on conflict (condicion) do nothing;

insert into dias (nombre) values ('Lunes'), ('Martes'), ('Miércoles'), ('Jueves'), ('Viernes'), ('Sábado');

insert into materia (codigo, nombre, area, bloque, nivel, cuatrimestre_id, carga_horaria, tipo, horas, programa)
select v.codigo, v.nombre, v.area, v.bloque, v.nivel, cu.id, v.carga_horaria, v.tipo, v.horas, ''
from (values
    ('isi-aed', 'Algoritmos y Estructura de Datos', 'Desarrollo de Software', 'Tecnologías Básicas', 1, 'Anual', 5, 'Obligatoria', 120),
    ('isi-am1', 'Análisis Matemático I', 'Matemática', 'Ciencias Básicas', 1, 'Anual', 5, 'Obligatoria', 120),
    ('isi-adc', 'Arquitectura de Computadoras', 'Computación y Comunicación de Datos', 'Tecnologías Básicas', 1, 'Anual', 4, 'Obligatoria', 96),
    ('isi-f1', 'Fisica', 'Fisica', 'Ciencias Básicas', 1, 'Anual', 4, 'Obligatoria', 120),
    ('isi-ing1', 'Inglés I', 'Ciencias Sociales', 'Ciencias Básicas', 1, '2C', 2, 'Obligatoria', 48),
    ('isi-led', 'Lógica y Estructuras Discretas', 'Desarrollo de Software', 'Ciencias Básicas de la Ingeniería', 1, '1C', 6, 'Obligatoria', 72),
    ('isi-sposn', 'Sistemas y Procesos de Negocio', 'Sistemas de Información', 'Sistemas de Información', 1, 'Anual', 3, 'Obligatoria', 72),
    ('isi-ayga', 'Álgebra y Geometría Analítica', 'Matemática', 'Ciencias Básicas', 1, 'Anual', 5, 'Obligatoria', 120),
    ('isi-am2', 'Análisis Matemático II', 'Matemática', 'Ciencias Básicas', 2, 'Anual', 5, 'Obligatoria', 120),
    ('isi-ads', 'Análisis de Sistemas de Información', 'Sistemas de Información', 'Tecnologías Aplicadas', 2, 'Anual', 6, 'Obligatoria', 144),
    ('isi-f2', 'Fisica II', 'Fisica', 'Ciencias Básicas', 2, '1C', 10, 'Obligatoria', 120),
    ('isi-iys', 'Ingeniería y Sociedad', 'Ciencias Sociales', 'Complementarias', 2, '1C', 2, 'Obligatoria', 48),
    ('isi-ing2', 'Inglés II', 'Ciencias Sociales', 'Ciencias Básicas', 2, 'Anual', 2, 'Obligatoria', 48),
    ('isi-pdp', 'Paradigmas de Programación', 'Desarrollo de Software', 'Tecnologías Básicas', 2, '2C', 8, 'Obligatoria', 96),
    ('isi-ssl', 'Sintaxis y Semántica de los Lenguajes', 'Desarrollo de Software', 'Tecnologías Básicas', 2, '1C', 8, 'Obligatoria', 96),
    ('isi-so', 'Sistemas Operativos', 'Computación y Comunicación de Datos', 'Tecnologías Aplicadas', 2, '2C', 8, 'Obligatoria', 96),
    ('isi-an', 'Análisis Numérico', 'Sistemas Inteligentes', 'Ciencias Básicas de la Ingeniería', 3, '2C', 6, 'Obligatoria', 72),
    ('isi-bdd', 'Base de Datos', 'Desarrollo de Software', 'Tecnologías Aplicadas', 3, '1C', 8, 'Obligatoria', 96),
    ('isi-com', 'Comunicación de Datos', 'Computación y Comunicación de Datos', 'Tecnologías Básicas', 3, '1C', 8, 'Obligatoria', 96),
    ('isi-dsw', 'Desarrollo de Software', 'Desarrollo de Software', 'Tecnologías Aplicadas', 3, '2C', 8, 'Obligatoria', 96),
    ('isi-dsi', 'Diseño de Sistemas de Información', 'Sistemas de Información', 'Sistemas de Información', 3, 'Anual', 6, 'Obligatoria', 144),
    ('isi-eco', 'Economía', 'Ciencias Sociales', 'Complementarias', 3, '1C', 3, 'Obligatoria', 72),
    ('isi-pye', 'Probabilidad y Estadística', 'Matemática', 'Ciencias Básicas', 3, '2C', 3, 'Obligatoria', 72),
    ('isi-admsi', 'Administración de Sistemas de Información', 'Sistemas de Información', 'Tecnologías Aplicadas', 4, 'Anual', 6, 'Obligatoria', 144),
    ('isi-iycs', 'Ingeniería y Calidad del Software', 'Desarrollo de Software', 'Tecnologías Aplicadas', 4, '1C', 6, 'Obligatoria', 72),
    ('isi-io', 'Investigación Operativa', 'Sistemas Inteligentes', 'Tecnologías Básicas', 4, '1C', 8, 'Obligatoria', 96),
    ('isi-leg', 'Legislación', 'Ciencias Sociales', 'Complementarias', 4, '1C', 2, 'Obligatoria', 48),
    ('isi-rdd', 'Redes de Datos', 'Computación y Comunicación de Datos', 'Tecnologías Aplicadas', 4, '2C', 8, 'Obligatoria', 96),
    ('isi-sim', 'Simulación', 'Sistemas Inteligentes', 'Tecnologías Básicas', 4, '1C', 6, 'Obligatoria', 72),
    ('isi-tpla', 'Tecnologías para la Automatización', 'Sistemas Inteligentes', 'Tecnologías Aplicadas', 4, '2C', 6, 'Obligatoria', 72),
    ('isi-cdd', 'Ciencia de Datos', 'Sistemas Inteligentes', 'Tecnologías Aplicadas', 5, '2C', 6, 'Obligatoria', 72),
    ('isi-gg', 'Gestión Gerencial', 'Gestión Ingenieril', 'Ciencias y Tecnologías Aplicadas', 5, '1C', 6, 'Obligatoria', 72),
    ('isi-ia', 'Inteligencia Artificial', 'Sistemas Inteligentes', 'Tecnologías Aplicadas', 5, '2C', 6, 'Obligatoria', 72),
    ('isi-pf', 'Proyecto Final', 'Sistemas de Información', 'Tecnologías Aplicadas', 5, 'Anual', 6, 'Obligatoria', 144),
    ('isi-ssi', 'Seguridad en los Sistemas de Información', 'Gestión Ingenieril', 'Tecnologías Aplicadas', 5, '1C', 6, 'Obligatoria', 72),
    ('isi-sg', 'Sistemas de Gestión', 'Gestión Ingenieril', 'Tecnologías Aplicadas', 5, '1C', 6, 'Obligatoria', 72)
) as v(codigo, nombre, area, bloque, nivel, cuatrimestre, carga_horaria, tipo, horas)
join cuatrimestre cu on cu.nombre = v.cuatrimestre
on conflict (codigo) do nothing;

insert into materiasPorcarrera (carrera_id, materia_id)
select c.id, m.id from carrera c, materia m
where c.nombre = 'Ingeniería en Sistemas de Información' and m.codigo like 'isi-%'
on conflict (carrera_id, materia_id) do nothing;

insert into correlativa (materia_id, requiere_id, tipo)
select m.id, r.id, v.tipo
from (values
    ('isi-pdp', 'isi-aed', 'regular'),
    ('isi-pdp', 'isi-led', 'regular'),
    ('isi-am2', 'isi-am1', 'regular'),
    ('isi-am2', 'isi-ayga', 'regular'),
    ('isi-f2', 'isi-am1', 'regular'),
    ('isi-f2', 'isi-f1', 'regular'),
    ('isi-ads', 'isi-aed', 'regular'),
    ('isi-ing2', 'isi-ing1', 'regular'),
    ('isi-ssl', 'isi-led', 'regular'),
    ('isi-ssl', 'isi-aed', 'regular'),
    ('isi-so', 'isi-adc', 'regular'),
    ('isi-ads', 'isi-sposn', 'regular'),
    ('isi-pye', 'isi-am1', 'regular'),
    ('isi-pye', 'isi-ayga', 'regular'),
    ('isi-eco', 'isi-ayga', 'aprobada'),
    ('isi-eco', 'isi-am1', 'aprobada'),
    ('isi-bdd', 'isi-led', 'aprobada'),
    ('isi-bdd', 'isi-aed', 'aprobada'),
    ('isi-bdd', 'isi-ssl', 'regular'),
    ('isi-bdd', 'isi-ads', 'regular'),
    ('isi-dsw', 'isi-ads', 'regular'),
    ('isi-dsw', 'isi-pdp', 'regular'),
    ('isi-dsw', 'isi-aed', 'aprobada'),
    ('isi-dsw', 'isi-led', 'aprobada'),
    ('isi-an', 'isi-am2', 'aprobada'),
    ('isi-com', 'isi-adc', 'aprobada'),
    ('isi-com', 'isi-f1', 'aprobada'),
    ('isi-an', 'isi-am1', 'aprobada'),
    ('isi-an', 'isi-ayga', 'aprobada'),
    ('isi-dsi', 'isi-pdp', 'regular'),
    ('isi-dsi', 'isi-ads', 'regular'),
    ('isi-dsi', 'isi-ing1', 'aprobada'),
    ('isi-dsi', 'isi-sposn', 'aprobada'),
    ('isi-dsi', 'isi-aed', 'aprobada'),
    ('isi-leg', 'isi-iys', 'regular'),
    ('isi-iycs', 'isi-bdd', 'regular'),
    ('isi-iycs', 'isi-dsw', 'regular'),
    ('isi-iycs', 'isi-dsi', 'regular'),
    ('isi-iycs', 'isi-pdp', 'aprobada'),
    ('isi-iycs', 'isi-ssl', 'aprobada'),
    ('isi-rdd', 'isi-com', 'regular'),
    ('isi-rdd', 'isi-so', 'regular'),
    ('isi-io', 'isi-an', 'regular'),
    ('isi-io', 'isi-pye', 'regular'),
    ('isi-sim', 'isi-pye', 'regular'),
    ('isi-sim', 'isi-am2', 'aprobada'),
    ('isi-tpla', 'isi-am2', 'aprobada'),
    ('isi-tpla', 'isi-f2', 'regular'),
    ('isi-tpla', 'isi-an', 'regular'),
    ('isi-admsi', 'isi-ads', 'aprobada'),
    ('isi-admsi', 'isi-dsi', 'regular'),
    ('isi-admsi', 'isi-eco', 'regular'),
    ('isi-ia', 'isi-sim', 'regular'),
    ('isi-ia', 'isi-pye', 'aprobada'),
    ('isi-ia', 'isi-an', 'aprobada'),
    ('isi-cdd', 'isi-pye', 'aprobada'),
    ('isi-cdd', 'isi-bdd', 'aprobada'),
    ('isi-cdd', 'isi-sim', 'regular'),
    ('isi-sg', 'isi-dsi', 'aprobada'),
    ('isi-sg', 'isi-io', 'regular'),
    ('isi-sg', 'isi-eco', 'regular'),
    ('isi-gg', 'isi-leg', 'regular'),
    ('isi-gg', 'isi-admsi', 'regular'),
    ('isi-gg', 'isi-eco', 'aprobada'),
    ('isi-ssi', 'isi-rdd', 'regular'),
    ('isi-ssi', 'isi-admsi', 'regular'),
    ('isi-ssi', 'isi-dsw', 'aprobada'),
    ('isi-ssi', 'isi-com', 'aprobada'),
    ('isi-pf', 'isi-ing2', 'aprobada'),
    ('isi-pf', 'isi-dsi', 'aprobada'),
    ('isi-pf', 'isi-dsw', 'aprobada'),
    ('isi-pf', 'isi-iycs', 'regular'),
    ('isi-pf', 'isi-rdd', 'regular'),
    ('isi-pf', 'isi-admsi', 'regular')
) as v(materia, requiere, tipo)
join materia m on m.codigo = v.materia
join materia r on r.codigo = v.requiere
on conflict (materia_id, requiere_id) do nothing;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
delete from correlativa where materia_id in (select id from materia where codigo like 'isi-%');
delete from materiasPorcarrera where materia_id in (select id from materia where codigo like 'isi-%');
delete from materia where codigo like 'isi-%';
delete from dias;
delete from condicion where condicion in ('Cursando', 'Regularizada', 'Aprobada');
delete from cuatrimestre where nombre in ('1C', '2C', 'Anual');
delete from carrera where nombre = 'Ingeniería en Sistemas de Información';
-- +goose StatementEnd
