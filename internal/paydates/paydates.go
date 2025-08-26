package paydates

import (
	"fmt"
	"time"

	"github.com/leogtzr/payment-dates-advisor/internal/model"
)

// PaymentDate represents a payment date with its original and adjusted values
type PaymentDate struct {
	Item     model.ConfigItem
	Original time.Time
	Adjusted time.Time
}

// FixedHoliday represents a fixed holiday (day and month)
type FixedHoliday struct {
	Day   int
	Month time.Month
}

// fixedHolidays lists the fixed bank holidays in Mexico
var fixedHolidays = []FixedHoliday{
	{1, time.January},    // Año Nuevo
	{1, time.May},        // Día del Trabajo
	{16, time.September}, // Día de la Independencia
	{1, time.October},    // Transmisión del Poder Ejecutivo (aplica en 2025)
	{2, time.November},   // Día de Muertos
	{12, time.December},  // Día del Empleado Bancario
	{25, time.December},  // Navidad
}

// GeneratePaymentDatesForMonth generates all payment dates for a given month
func GeneratePaymentDatesForMonth(items []model.ConfigItem, year int, month time.Month, loc *time.Location) []PaymentDate {
	var paymentDates []PaymentDate

	for _, item := range items {
		pd, err := GeneratePaymentDate(item, year, month, loc)
		if err != nil {
			continue
		}
		paymentDates = append(paymentDates, pd)
	}

	return paymentDates
}

// GetLastDayOfMonth returns the last day of the given month and year
func GetLastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// IsMobileHoliday checks if a date is a mobile holiday
func IsMobileHoliday(date time.Time) bool {
	// Día de la Constitución: primer lunes de febrero
	if date.Month() == time.February {
		firstMonday := time.Date(date.Year(), time.February, 1, 0, 0, 0, 0, date.Location())
		for firstMonday.Weekday() != time.Monday {
			firstMonday = firstMonday.AddDate(0, 0, 1)
		}
		if date.Equal(firstMonday) {
			return true
		}
	}
	return false
}

// IsFixedHoliday checks if a date is a fixed or mobile holiday
func IsFixedHoliday(date time.Time) bool {
	for _, holiday := range fixedHolidays {
		if date.Day() == holiday.Day && date.Month() == holiday.Month {
			return true
		}
	}
	return IsMobileHoliday(date)
}

// PreviousFriday returns the previous Friday from the given date
func PreviousFriday(d time.Time) time.Time {
	f := d
	for f.Weekday() != time.Friday {
		f = f.AddDate(0, 0, -1)
	}
	return f
}

// PreviousBusinessDay returns the previous business day before the given date, avoiding weekends and holidays
func PreviousBusinessDay(d time.Time) time.Time {
	previous := d.AddDate(0, 0, -1)
	for previous.Weekday() == time.Saturday || previous.Weekday() == time.Sunday || IsFixedHoliday(previous) {
		previous = previous.AddDate(0, 0, -1)
	}
	return previous
}

// AdjustIfWeekend adjusts the date to the next Monday if it falls on a Saturday or Sunday,
// and to the next business day if it falls on a fixed or mobile holiday
func AdjustIfWeekend(date time.Time) time.Time {
	adjusted := date
	for {
		if adjusted.Weekday() == time.Saturday {
			adjusted = adjusted.AddDate(0, 0, 2) // Trasladar al lunes
		} else if adjusted.Weekday() == time.Sunday {
			adjusted = adjusted.AddDate(0, 0, 1) // Trasladar al lunes
		} else if IsFixedHoliday(adjusted) {
			adjusted = adjusted.AddDate(0, 0, 1) // Trasladar al siguiente día si es feriado
		} else {
			break // Salir si no es fin de semana ni feriado
		}
	}
	return adjusted
}

// GeneratePaymentDate generates a payment date for an item in a specific month and year
func GeneratePaymentDate(item model.ConfigItem, year int, month time.Month, loc *time.Location) (PaymentDate, error) {
	var pd PaymentDate
	if item.EveryWhenDay <= 0 || item.EveryWhenDay > GetLastDayOfMonth(year, month, loc) {
		return pd, fmt.Errorf("invalid day %d for month %s", item.EveryWhenDay, month)
	}

	original := time.Date(year, month, item.EveryWhenDay, 0, 0, 0, 0, loc)
	adjusted := AdjustIfWeekend(original)

	return PaymentDate{
		Item:     item,
		Original: original,
		Adjusted: adjusted,
	}, nil
}

// IsAdjusted returns true if the payment date was moved due to weekend or holiday
func (pd PaymentDate) IsAdjusted() bool {
	return !pd.Original.Equal(pd.Adjusted)
}

// DaysUntil calculates the number of days from today until the adjusted payment date
func (pd PaymentDate) DaysUntil(today time.Time) int {
	return int(pd.Adjusted.Sub(today).Hours() / 24)
}

// IsUpcoming checks if the payment date is within the specified number of days
func (pd PaymentDate) IsUpcoming(today time.Time, daysThreshold int) bool {
	daysUntil := pd.DaysUntil(today)
	return daysUntil >= 0 && daysUntil <= daysThreshold
}
