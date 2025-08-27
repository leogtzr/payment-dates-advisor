package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/leogtzr/payment-dates-advisor/internal/config"
	"github.com/leogtzr/payment-dates-advisor/internal/ui"
)

func main() {
	monthsFlag := flag.Int("months", 4, "Número de meses hacia adelante (incluye mes actual)")
	configPath := flag.String("config", "config.yaml", "Ruta al archivo YAML de configuración")
	daysNotify := flag.Int("days", 10, "Número de días para notificaciones (default 10)")
	flag.Parse()

	monthsAhead := *monthsFlag
	if monthsAhead < 0 {
		log.Fatal("months >= 0 required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	loc := time.Local
	start := time.Now().In(loc)
	today := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)

	// Create printer with styling
	printer := ui.NewPrinter(os.Stdout, *daysNotify, today)

	// Render the range of months
	printer.RenderRange(cfg.Items, start, monthsAhead, loc)
}
