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
