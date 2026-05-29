package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	APIKey          string
	InitialBankroll float64
	KellyFraction   float64 // Coeficiente para fraccionar Kelly (ej. 0.25 para cuarto de Kelly)
	SimulationMode  bool
}

// LoadConfig carga la configuración desde el entorno o valores por defecto
func LoadConfig() *Config {
	port := getEnv("PORT", "8080")
	apiKey := getEnv("THE_ODDS_API_KEY", "") // Si está vacía, el bot usará datos simulados reales.
	
	bankrollStr := getEnv("INITIAL_BANKROLL", "1000.0")
	bankroll, err := strconv.ParseFloat(bankrollStr, 64)
	if err != nil {
		bankroll = 1000.0
	}

	kellyStr := getEnv("KELLY_FRACTION", "0.25")
	kellyFraction, err := strconv.ParseFloat(kellyStr, 64)
	if err != nil {
		kellyFraction = 0.25
	}

	simModeStr := getEnv("SIMULATION_MODE", "true")
	simMode, err := strconv.ParseBool(simModeStr)
	if err != nil {
		simMode = true
	}

	return &Config{
		Port:            port,
		APIKey:          apiKey,
		InitialBankroll: bankroll,
		KellyFraction:   kellyFraction,
		SimulationMode:  simMode,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
