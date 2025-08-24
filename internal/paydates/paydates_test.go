package paydates

import (
	"testing"
	"time"

	"github.com/leogtzr/payment-dates-advisor/internal/model"
)

// Test GetLastDayOfMonth for feb (non-leap and leap)
func TestGetLastDayOfMonth(t *testing.T) {
	loc := time.UTC
	d1 := GetLastDayOfMonth(2025, time.February, loc) // 2025 no es bisiesto
	if d1 != 28 {
		t.Fatalf("expected 28 got %d", d1)
	}

	d2 := GetLastDayOfMonth(2024, time.February, loc) // 2024 bisiesto
	if d2 != 29 {
		t.Fatalf("expected 29 got %d", d2)
	}
}

// Test AdjustIfWeekend
func TestAdjustIfWeekend(t *testing.T) {
	loc := time.UTC
	// 2025-11-01 is Saturday
	sat := time.Date(2025, 11, 1, 0, 0, 0, 0, loc)
	adj := AdjustIfWeekend(sat)
	if adj.Weekday() != time.Monday || adj.Day() != 3 {
		t.Fatalf("expected Monday 3rd, got %v", adj)
	}

	// 2025-11-03 is Monday -> should remain same
	mon := time.Date(2025, 11, 3, 0, 0, 0, 0, loc)
	if !AdjustIfWeekend(mon).Equal(mon) {
		t.Fatalf("expected same day for Monday")
	}
}

// Test PreviousFriday
func TestPreviousFriday(t *testing.T) {
	loc := time.UTC
	// Monday Nov 3 2025 -> previous Friday should be Nov 31? Actually Friday Oct 31, but test uses known
	mon := time.Date(2025, 11, 3, 0, 0, 0, 0, loc)
	f := PreviousFriday(mon)
	if f.Weekday() != time.Friday {
		t.Fatalf("expected Friday, got %v", f.Weekday())
	}
}

// Test GeneratePaymentDate (example for item day 1)
func TestGeneratePaymentDate(t *testing.T) {
	loc := time.UTC
	item := model.ConfigItem{Name: "Pago de Kia 🚗", EveryWhenDay: 1}
	pd, err := GeneratePaymentDate(item, 2025, time.November, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// original should be 2025-11-01 (Saturday)
	if pd.Original.Day() != 1 || pd.Original.Month() != time.November {
		t.Fatalf("unexpected original date: %v", pd.Original)
	}
	// adjusted should be Monday 3rd
	if pd.Adjusted.Day() != 3 || pd.Adjusted.Weekday() != time.Monday {
		t.Fatalf("unexpected adjusted date: %v", pd.Adjusted)
	}
}
