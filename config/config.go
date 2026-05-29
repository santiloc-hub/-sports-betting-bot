package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port            string
	APIKey          string
	InitialBankroll float64
	KellyFraction   float64 // Coeficiente para fraccionar Kelly (ej. 0.25 para cuarto de Kelly)
	SimulationMode  bool
	SportsToScan    []string
	MinEVThreshold  float64
	SharpBookmaker  string
	MinOdds         float64
	MaxOdds         float64
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

	sportsStr := getEnv("SPORTS_TO_SCAN", "upcoming")
	var sports []string
	if sportsStr != "" {
		parts := strings.Split(sportsStr, ",")
		for _, p := range parts {
			trimmed := cleanEnvStr(p)
			if trimmed != "" {
				sports = append(sports, trimmed)
			}
		}
	}
	if len(sports) == 0 {
		sports = []string{"upcoming"}
	}

	minEVStr := getEnv("MIN_EV_THRESHOLD", "1.0")
	minEV, err := strconv.ParseFloat(minEVStr, 64)
	if err != nil {
		minEV = 1.0
	}

	sharpBM := getEnv("SHARP_BOOKMAKER", "Pinnacle")

	minOddsStr := getEnv("MIN_ODDS", "1.05")
	minOdds, err := strconv.ParseFloat(minOddsStr, 64)
	if err != nil {
		minOdds = 1.05
	}

	maxOddsStr := getEnv("MAX_ODDS", "10.0")
	maxOdds, err := strconv.ParseFloat(maxOddsStr, 64)
	if err != nil {
		maxOdds = 10.0
	}

	return &Config{
		Port:            port,
		APIKey:          apiKey,
		InitialBankroll: bankroll,
		KellyFraction:   kellyFraction,
		SimulationMode:  simMode,
		SportsToScan:    sports,
		MinEVThreshold:  minEV,
		SharpBookmaker:  sharpBM,
		MinOdds:         minOdds,
		MaxOdds:         maxOdds,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func init() {
	loadDotEnv()
}

func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // Si no existe, no pasa nada
	}
	defer file.Close()

	var lines []string
	var buf []byte
	tempBuf := make([]byte, 1)
	for {
		n, err := file.Read(tempBuf)
		if n > 0 {
			if tempBuf[0] == '\n' {
				lines = append(lines, string(buf))
				buf = []byte{}
			} else {
				buf = append(buf, tempBuf[0])
			}
		}
		if err != nil {
			if len(buf) > 0 {
				lines = append(lines, string(buf))
			}
			break
		}
	}

	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		for i := 0; i < len(line); i++ {
			if line[i] == '=' {
				key := line[:i]
				val := line[i+1:]
				key = cleanEnvStr(key)
				val = cleanEnvStr(val)
				_ = os.Setenv(key, val)
				break
			}
		}
	}
}

func cleanEnvStr(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	return s
}
