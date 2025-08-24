package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/leogtzr/payment-dates-advisor/internal/model"
	"github.com/leogtzr/payment-dates-advisor/internal/paydates"
)

const FormatDate = "2006-01-02"

var (
	monthNamesES = []string{
		"ENERO", "FEBRERO", "MARZO", "ABRIL", "MAYO", "JUNIO",
		"JULIO", "AGOSTO", "SEPTIEMBRE", "OCTUBRE", "NOVIEMBRE", "DICIEMBRE",
	}

	weekDayEnToES = map[time.Weekday]string{
		time.Sunday:    "Domingo",
		time.Monday:    "Lunes",
		time.Tuesday:   "Martes",
		time.Wednesday: "Miércoles",
		time.Thursday:  "Jueves",
		time.Friday:    "Viernes",
		time.Saturday:  "Sábado",
	}
)

// Printer handles the rendering of payment dates with styling
type Printer struct {
	Out                   io.Writer
	SuggestionStyle       *color.Color
	UpcomingStyle         *color.Color
	UpcomingStyleOnMonday *color.Color
	DaysThreshold         int
	Today                 time.Time
}

// NewPrinter creates a new Printer with default styling
func NewPrinter(w io.Writer, daysThreshold int, today time.Time) *Printer {
	return &Printer{
		Out:                   w,
		SuggestionStyle:       color.New(color.FgHiWhite, color.BgGreen, color.Bold),
		UpcomingStyle:         color.New(color.FgHiYellow, color.Bold),             // Shining yellow in bold for upcoming dates
		UpcomingStyleOnMonday: color.New(color.FgHiBlue, color.Bold, color.Italic), // Shining blue in bold and italic for upcoming dates on Monday
		DaysThreshold:         daysThreshold,
		Today:                 today,
	}
}

// FormatPaymentDateLine formatea la información de la fecha de pago como una cadena
func FormatPaymentDateLine(pd paydates.PaymentDate) (string, time.Weekday) {
	if pd.IsAdjusted() {
		reason := "fin de semana"
		if paydates.IsFixedHoliday(pd.Original) {
			reason = "feriado"
		} else if paydates.IsFixedHoliday(pd.Adjusted) {
			reason = "feriado tras ajuste"
		} else if pd.Original.Weekday() == time.Saturday || pd.Original.Weekday() == time.Sunday {
			// Verificar si el lunes siguiente a un fin de semana es un feriado
			var mondayAfter time.Time
			if pd.Original.Weekday() == time.Saturday {
				mondayAfter = pd.Original.AddDate(0, 0, 2)
			} else if pd.Original.Weekday() == time.Sunday {
				mondayAfter = pd.Original.AddDate(0, 0, 1)
			}
			if !mondayAfter.IsZero() && paydates.IsFixedHoliday(mondayAfter) {
				reason = "feriado tras ajuste"
			}
		}
		return fmt.Sprintf("%s, %s (%s) -> pago en %s (%s) por %s",
			pd.Item.Name,
			pd.Original.Format(FormatDate), weekDayEnToES[pd.Original.Weekday()],
			pd.Adjusted.Format(FormatDate), weekDayEnToES[pd.Adjusted.Weekday()], reason), pd.Adjusted.Weekday()
	}
	return fmt.Sprintf("%s, %s (%s)", pd.Item.Name, pd.Original.Format(FormatDate), weekDayEnToES[pd.Original.Weekday()]), pd.Original.Weekday()
}

// FormatSuggestionMessage formats the suggestion message for adjusted dates or Monday payments
func FormatSuggestionMessage(pd paydates.PaymentDate) string {
	var suggestedDate time.Time
	if pd.IsAdjusted() {
		if paydates.IsFixedHoliday(pd.Original) {
			// Si se ajustó por feriado, sugerir el día hábil anterior a la fecha original
			suggestedDate = paydates.PreviousBusinessDay(pd.Original)
		} else {
			// Si se ajustó por fin de semana, sugerir el viernes anterior a la fecha ajustada
			suggestedDate = paydates.PreviousFriday(pd.Adjusted)
		}
	} else if pd.Original.Weekday() == time.Monday {
		// Si la fecha original es lunes (sin ajuste), sugerir el viernes anterior
		suggestedDate = paydates.PreviousFriday(pd.Original)
	} else {
		// Para fechas no ajustadas que no son lunes, no debería llamarse esta función
		return ""
	}

	return fmt.Sprintf("Deja fondos disponibles desde el %s %s",
		weekDayEnToES[suggestedDate.Weekday()], suggestedDate.Format(FormatDate))
}

// PrintPaymentDate prints a payment date with appropriate styling
func (p *Printer) PrintPaymentDate(pd paydates.PaymentDate) {
	line, weekDay := FormatPaymentDateLine(pd)
	isUpcoming := pd.IsUpcoming(p.Today, p.DaysThreshold)

	if isUpcoming && weekDay == time.Monday {
		_, _ = p.UpcomingStyleOnMonday.Fprintln(p.Out, line+", se recomienda dejar el pago desde el viernes anterior.")
	} else if isUpcoming {
		_, _ = p.UpcomingStyle.Fprintln(p.Out, line)
	} else {
		_, _ = fmt.Fprintln(p.Out, line)
	}

	if pd.IsAdjusted() || (weekDay == time.Monday) {
		msg := FormatSuggestionMessage(pd)
		if msg != "" {
			// Print suggestion with style
			_, _ = fmt.Fprint(p.Out, "   ")
			_, _ = p.SuggestionStyle.Fprintln(p.Out, msg)
		}
	}
}

// PrintInvalidDay prints an error message for invalid days
func (p *Printer) PrintInvalidDay(itemName string, day int) {
	_, _ = fmt.Fprintf(p.Out, "%s: día inválido (%d)\n", itemName, day)
}

// GetMonthName returns the Spanish name of the month
func GetMonthName(month time.Month) string {
	return monthNamesES[int(month)-1]
}

// PrintMonthHeader prints the month header
func (p *Printer) PrintMonthHeader(year int, month time.Month) {
	monthName := GetMonthName(month)
	_, _ = fmt.Fprintf(p.Out, "=== %d-%02d %s ===\n", year, month, monthName)
}

// RenderMonth renders all payment dates for a specific month
func (p *Printer) RenderMonth(items []model.ConfigItem, year int, month time.Month, loc *time.Location) {
	p.PrintMonthHeader(year, month)

	// Generate payment dates for this month
	paymentDates := paydates.GeneratePaymentDatesForMonth(items, year, month, loc)

	// Print valid payment dates
	for _, pd := range paymentDates {
		// Verify if the adjusted date is within the next days threshold
		p.PrintPaymentDate(pd)
	}

	// Print invalid days (items that couldn't generate payment dates)
	for _, item := range items {
		if item.EveryWhenDay <= 0 {
			p.PrintInvalidDay(item.Name, item.EveryWhenDay)
		}
	}

	_, _ = fmt.Fprintln(p.Out)
}

// RenderRange renders payment dates for a range of months
func (p *Printer) RenderRange(items []model.ConfigItem, start time.Time, monthsAhead int, loc *time.Location) {
	for i := 0; i <= monthsAhead; i++ {
		t := start.AddDate(0, i, 0)
		year, month := t.Year(), t.Month()
		p.RenderMonth(items, year, month, loc)
	}
}
