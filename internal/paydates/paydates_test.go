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

// Test PreviousBusinessDay
func TestPreviousBusinessDay(t *testing.T) {
	loc := time.UTC
	// 2026-01-01 (Thursday, holiday) -> previous business day should be 2025-12-31 (Wednesday)
	newYear := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	f := PreviousBusinessDay(newYear)
	if f.Weekday() != time.Wednesday || f.Day() != 31 || f.Month() != time.December || f.Year() != 2025 {
		t.Fatalf("expected Wednesday 31st December 2025, got %v", f)
	}

	// 2025-11-03 (Monday, non-holiday) -> previous business day should be 2025-10-31 (Friday)
	mon := time.Date(2025, 11, 3, 0, 0, 0, 0, loc)
	f = PreviousBusinessDay(mon)
	if f.Weekday() != time.Friday || f.Day() != 31 || f.Month() != time.October {
		t.Fatalf("expected Friday 31st October, got %v", f)
	}
}

// Test PreviousFriday
func TestPreviousFriday(t *testing.T) {
	loc := time.UTC
	// 2025-11-03 (Monday) -> previous Friday should be 2025-10-31
	mon := time.Date(2025, 11, 3, 0, 0, 0, 0, loc)
	f := PreviousFriday(mon)
	if f.Weekday() != time.Friday || f.Day() != 31 || f.Month() != time.October {
		t.Fatalf("expected Friday 31st October, got %v", f)
	}
}

// Test GeneratePaymentDate
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

// TestGeneratePaymentDateWithHoliday prueba la generación de fechas con feriados
func TestGeneratePaymentDateWithHoliday(t *testing.T) {
	loc := time.UTC
	item := model.ConfigItem{Name: "Pago de Kia 🚗", EveryWhenDay: 1}
	// Prueba para enero 2026 (1 de enero es feriado)
	pd, err := GeneratePaymentDate(item, 2026, time.January, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd.Original.Day() != 1 || pd.Original.Month() != time.January {
		t.Fatalf("unexpected original date: %v", pd.Original)
	}
	if pd.Adjusted.Day() != 2 || pd.Adjusted.Weekday() != time.Friday {
		t.Fatalf("expected adjusted date to be 2nd January (Friday), got %v", pd.Adjusted)
	}

	// Prueba para octubre 2025 (1 de octubre es feriado)
	pd, err = GeneratePaymentDate(item, 2025, time.October, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd.Original.Day() != 1 || pd.Original.Month() != time.October {
		t.Fatalf("unexpected original date: %v", pd.Original)
	}
	if pd.Adjusted.Day() != 2 || pd.Adjusted.Weekday() != time.Thursday {
		t.Fatalf("expected adjusted date to be 2nd October (Thursday), got %v", pd.Adjusted)
	}

	// Prueba para febrero 2026 (1 de febrero es domingo, 2 de febrero es feriado móvil)
	pd, err = GeneratePaymentDate(item, 2026, time.February, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd.Original.Day() != 1 || pd.Original.Month() != time.February {
		t.Fatalf("unexpected original date: %v", pd.Original)
	}
	if pd.Adjusted.Day() != 3 || pd.Adjusted.Weekday() != time.Tuesday {
		t.Fatalf("expected adjusted date to be 3rd February (Tuesday), got %v", pd.Adjusted)
	}
}
