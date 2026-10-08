package handlers

import (
	"chedul-core/internals/domain"
	"fmt"
	"strings"
	"time"
)

// Calendario en formato iCalendar (RFC 5545) para suscribirse desde Google
// Calendar, Apple o Outlook. Lleva los eventos del alumno, su horario semanal
// como eventos que se repiten y las fechas de la facultad.

const icsZona = "America/Argentina/Buenos_Aires"

// Argentina no tiene horario de verano: con una zona fija alcanza y no hace
// falta tzdata en la imagen.
var zonaArgentina = time.FixedZone("ART", -3*60*60)

// Meses en que se cursa cada cuatrimestre, igual que en el frontend
var mesesDeCursada = map[string][2]time.Month{
	"1C": {time.March, time.July},
	"2C": {time.August, time.December},
}

var nombreTipoEvento = map[string]string{
	"parcial":      "Parcial",
	"final":        "Final",
	"entrega":      "Entrega",
	"recordatorio": "Recordatorio",
	"otro":         "Evento",
}

var diasICS = []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}

type icsWriter struct {
	b strings.Builder
}

// linea escribe una propiedad cortandola a 75 bytes como pide el formato
// (las lineas que siguen arrancan con un espacio, que tambien cuenta), sin
// partir un caracter UTF-8 a la mitad.
func (w *icsWriter) linea(s string) {
	limite := 75
	for len(s) > limite {
		corte := limite
		for corte > 0 && s[corte]&0xC0 == 0x80 {
			corte--
		}
		w.b.WriteString(s[:corte])
		w.b.WriteString("\r\n ")
		s = s[corte:]
		limite = 74
	}
	w.b.WriteString(s)
	w.b.WriteString("\r\n")
}

func escaparICS(s string) string {
	r := strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`, "\r", `\n`)
	return r.Replace(s)
}

func fechaICS(t time.Time) string { return t.Format("20060102") }

func fechaHoraICS(t time.Time) string { return t.Format("20060102T150405") }

func parsearFecha(s string) (time.Time, bool) {
	t, err := time.ParseInLocation(domain.FormatoFecha, s, zonaArgentina)
	return t, err == nil
}

// conHora devuelve el dia con la hora HH:MM en la zona de Argentina.
func conHora(dia time.Time, hora string) (time.Time, bool) {
	h, err := time.Parse(domain.FormatoHora, hora)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(dia.Year(), dia.Month(), dia.Day(), h.Hour(), h.Minute(), 0, 0, zonaArgentina), true
}

// diaSemana pasa de time.Weekday (0 = domingo) a 1 = lunes ... 7 = domingo.
func diaSemana(t time.Time) int {
	return (int(t.Weekday())+6)%7 + 1
}

// rangoDeClase devuelve entre que dias se repite una clase. Las de una
// comision se cursan en los meses de su cuatrimestre del año actual; las
// cargadas a mano arrancan unas semanas antes de hoy y no terminan, salvo que
// tengan fecha de fin.
func rangoDeClase(clase domain.Clase, hoy time.Time) (desde time.Time, hasta *time.Time) {
	desde = hoy.AddDate(0, 0, -28)
	if clase.Cuatrimestre != nil {
		if meses, ok := mesesDeCursada[*clase.Cuatrimestre]; ok {
			desde = time.Date(hoy.Year(), meses[0], 1, 0, 0, 0, 0, zonaArgentina)
			fin := time.Date(hoy.Year(), meses[1]+1, 0, 0, 0, 0, 0, zonaArgentina)
			hasta = &fin
		}
	}
	if clase.Hasta != nil {
		if fin, ok := parsearFecha(*clase.Hasta); ok && (hasta == nil || fin.Before(*hasta)) {
			hasta = &fin
		}
	}
	return desde, hasta
}

// diasSinClase son los feriados y recesos de la facultad, dia por dia.
func diasSinClase(academicas []domain.FechaAcademica) map[string]bool {
	dias := map[string]bool{}
	for _, f := range academicas {
		if f.Tipo != "feriado" && f.Tipo != "receso" {
			continue
		}
		desde, ok1 := parsearFecha(f.Desde)
		hasta, ok2 := parsearFecha(f.Hasta)
		if !ok1 || !ok2 {
			continue
		}
		for d := desde; !d.After(hasta); d = d.AddDate(0, 0, 1) {
			dias[d.Format(domain.FormatoFecha)] = true
		}
	}
	return dias
}

func (w *icsWriter) evento(e domain.Evento, sello string) {
	dia, ok := parsearFecha(e.Fecha)
	if !ok {
		return
	}
	w.linea("BEGIN:VEVENT")
	w.linea(fmt.Sprintf("UID:evento-%d@chedul", e.ID))
	w.linea("DTSTAMP:" + sello)

	inicio, conHorario := time.Time{}, false
	if e.Hora != nil {
		inicio, conHorario = conHora(dia, *e.Hora)
	}
	if conHorario {
		w.linea(fmt.Sprintf("DTSTART;TZID=%s:%s", icsZona, fechaHoraICS(inicio)))
		duracion := "PT1H"
		if e.Tipo == "parcial" || e.Tipo == "final" {
			duracion = "PT2H"
		}
		w.linea("DURATION:" + duracion)
	} else {
		w.linea("DTSTART;VALUE=DATE:" + fechaICS(dia))
		w.linea("DTEND;VALUE=DATE:" + fechaICS(dia.AddDate(0, 0, 1)))
	}

	titulo := e.Titulo
	if e.Materia != nil && !strings.Contains(strings.ToLower(titulo), strings.ToLower(e.Materia.Nombre)) {
		titulo += " · " + e.Materia.Nombre
	}
	w.linea("SUMMARY:" + escaparICS(titulo))

	descripcion := nombreTipoEvento[e.Tipo]
	if e.Descripcion != "" {
		descripcion += "\n" + e.Descripcion
	}
	w.linea("DESCRIPTION:" + escaparICS(descripcion))
	w.linea("CATEGORIES:" + escaparICS(nombreTipoEvento[e.Tipo]))
	w.linea("END:VEVENT")
}

func (w *icsWriter) clase(c domain.Clase, hoy time.Time, sinClase map[string]bool, sello string) {
	desde, hasta := rangoDeClase(c, hoy)
	// Primer dia del rango que cae en el dia de la clase
	primero := desde.AddDate(0, 0, (c.Dia-diaSemana(desde)+7)%7)
	if hasta != nil && primero.After(*hasta) {
		return
	}
	inicio, ok1 := conHora(primero, c.HoraInicio)
	fin, ok2 := conHora(primero, c.HoraFin)
	if !ok1 || !ok2 || c.Dia < 1 || c.Dia > 7 {
		return
	}

	w.linea("BEGIN:VEVENT")
	w.linea(fmt.Sprintf("UID:clase-%d@chedul", c.ID))
	w.linea("DTSTAMP:" + sello)
	w.linea(fmt.Sprintf("DTSTART;TZID=%s:%s", icsZona, fechaHoraICS(inicio)))
	w.linea(fmt.Sprintf("DTEND;TZID=%s:%s", icsZona, fechaHoraICS(fin)))

	regla := "RRULE:FREQ=WEEKLY;BYDAY=" + diasICS[c.Dia-1]
	if hasta != nil {
		// Con DTSTART en una zona, UNTIL va en UTC: el final del ultimo dia
		ultimo := time.Date(hasta.Year(), hasta.Month(), hasta.Day(), 23, 59, 59, 0, zonaArgentina)
		regla += ";UNTIL=" + ultimo.UTC().Format("20060102T150405Z")
	}
	w.linea(regla)

	// Los feriados y recesos no hay clase
	if c.Tipo == "clase" {
		for d := primero; hasta == nil || !d.After(*hasta); d = d.AddDate(0, 0, 7) {
			if d.After(hoy.AddDate(1, 0, 0)) {
				break
			}
			if sinClase[d.Format(domain.FormatoFecha)] {
				saltear, _ := conHora(d, c.HoraInicio)
				w.linea(fmt.Sprintf("EXDATE;TZID=%s:%s", icsZona, fechaHoraICS(saltear)))
			}
		}
	}

	w.linea("SUMMARY:" + escaparICS(c.Titulo))
	if c.Aula != "" {
		w.linea("LOCATION:" + escaparICS(c.Aula))
	}
	if c.Materia != nil && c.Materia.Nombre != c.Titulo {
		w.linea("DESCRIPTION:" + escaparICS(c.Materia.Nombre))
	}
	w.linea("END:VEVENT")
}

func (w *icsWriter) academica(f domain.FechaAcademica, sello string) {
	desde, ok1 := parsearFecha(f.Desde)
	hasta, ok2 := parsearFecha(f.Hasta)
	if !ok1 || !ok2 {
		return
	}
	w.linea("BEGIN:VEVENT")
	w.linea(fmt.Sprintf("UID:facultad-%d@chedul", f.ID))
	w.linea("DTSTAMP:" + sello)
	w.linea("DTSTART;VALUE=DATE:" + fechaICS(desde))
	w.linea("DTEND;VALUE=DATE:" + fechaICS(hasta.AddDate(0, 0, 1)))
	w.linea("SUMMARY:" + escaparICS(f.Titulo))
	w.linea("DESCRIPTION:Calendario de la facultad")
	// No ocupa: no te marca como ocupado todo el dia
	w.linea("TRANSP:TRANSPARENT")
	w.linea("END:VEVENT")
}

// ArmarICS arma el calendario del alumno. ahora define el año de cursada y el
// sello de los eventos.
func ArmarICS(eventos []domain.Evento, clases []domain.Clase, academicas []domain.FechaAcademica, ahora time.Time) string {
	ahora = ahora.In(zonaArgentina)
	hoy := time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 0, 0, 0, 0, zonaArgentina)
	sello := ahora.UTC().Format("20060102T150405Z")

	var w icsWriter
	w.linea("BEGIN:VCALENDAR")
	w.linea("VERSION:2.0")
	w.linea("PRODID:-//Chedul//Calendario//ES")
	w.linea("CALSCALE:GREGORIAN")
	w.linea("METHOD:PUBLISH")
	w.linea("X-WR-CALNAME:Chedul")
	w.linea("X-WR-CALDESC:Clases\\, parciales\\, finales y fechas de la facultad")
	w.linea("X-WR-TIMEZONE:" + icsZona)
	w.linea("REFRESH-INTERVAL;VALUE=DURATION:PT6H")
	w.linea("X-PUBLISHED-TTL:PT6H")

	w.linea("BEGIN:VTIMEZONE")
	w.linea("TZID:" + icsZona)
	w.linea("BEGIN:STANDARD")
	w.linea("DTSTART:19700101T000000")
	w.linea("TZOFFSETFROM:-0300")
	w.linea("TZOFFSETTO:-0300")
	w.linea("TZNAME:-03")
	w.linea("END:STANDARD")
	w.linea("END:VTIMEZONE")

	for _, e := range eventos {
		w.evento(e, sello)
	}
	sinClase := diasSinClase(academicas)
	for _, c := range clases {
		w.clase(c, hoy, sinClase, sello)
	}
	for _, f := range academicas {
		w.academica(f, sello)
	}

	w.linea("END:VCALENDAR")
	return w.b.String()
}
