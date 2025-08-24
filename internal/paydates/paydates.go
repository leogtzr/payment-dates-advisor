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

// GeneratePaymentDatesForMonth generates all payment dates for a given month
func GeneratePaymentDatesForMonth(items []model.ConfigItem, year int, month time.Month, loc *time.Location) []PaymentDate {
	var paymentDates []PaymentDate

	for _, item := range items {
		pd, err := GeneratePaymentDate(item, year, month, loc)
		if err != nil {
			// In the original code, invalid days were printed directly
			// We'll skip them here and let the caller handle the error display
			continue
		}
		paymentDates = append(paymentDates, pd)
	}

	return paymentDates
}

// GetLastDayOfMonth returns the last day of the given month and year
func GetLastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	// Last day of month
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// AdjustDayForMonth adjusts the day if it exceeds the last day of the month
func AdjustDayForMonth(day int, year int, month time.Month, loc *time.Location) int {
	lastOfMonth := GetLastDayOfMonth(year, month, loc)
	d := day
	if d > lastOfMonth {
		d = lastOfMonth
	}
	return d
}

// Returns the previous Friday (or the same day if it's already Friday)
// Moves backwards until it finds Friday.
func PreviousFriday(d time.Time) time.Time {
	f := d
	for f.Weekday() != time.Friday {
		f = f.AddDate(0, 0, -1)
	}
	return f
}

// If it falls on a weekend, move to the next Monday
func AdjustIfWeekend(d time.Time) time.Time {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDate(0, 0, 2) // sábado -> lunes
	case time.Sunday:
		return d.AddDate(0, 0, 1) // domingo -> lunes
	default:
		return d
	}
}

// GeneratePaymentDate creates a PaymentDate for a given item, year, month, and location
func GeneratePaymentDate(item model.ConfigItem, year int, month time.Month, loc *time.Location) (PaymentDate, error) {
	day := item.EveryWhenDay
	if day <= 0 {
		return PaymentDate{}, fmt.Errorf("día inválido (%d) para %s", day, item.Name)
	}

	d := AdjustDayForMonth(day, year, month, loc)
	orig := time.Date(year, month, d, 0, 0, 0, 0, loc)
	adjusted := AdjustIfWeekend(orig)

	return PaymentDate{
		Item:     item,
		Original: orig,
		Adjusted: adjusted,
	}, nil
}

// IsAdjusted returns true if the payment date was moved due to weekend
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
