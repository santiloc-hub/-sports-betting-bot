package strategy

// CalculateKelly calcula la fracción óptima de banca a apostar
// p: probabilidad real estimada (0.0 a 1.0)
// odds: cuota decimal disponible (ej. 2.10)
// fractionMultiplier: factor de atenuación de riesgo (ej. 0.25 para un cuarto de Kelly)
func CalculateKelly(p, odds, fractionMultiplier float64) float64 {
	if odds <= 1.0 || p <= 0.0 || p >= 1.0 {
		return 0.0
	}

	// b es las ganancias netas obtenidas por peso apostado (odds - 1)
	b := odds - 1.0
	
	// Fórmula de Kelly: f = (p * b - q) / b  donde q = 1 - p
	q := 1.0 - p
	f := (p*b - q) / b

	if f <= 0.0 {
		return 0.0 // No hay ventaja (+EV)
	}

	// Aplicar el multiplicador fraccional de Kelly para reducir la volatilidad
	safeF := f * fractionMultiplier

	// Criterio de prudencia: nunca arriesgar más del 10% de la banca en una sola apuesta
	if safeF > 0.10 {
		safeF = 0.10
	}

	return safeF
}

type ArbitrageOpp struct {
	IsArbitrage   bool    `json:"isArbitrage"`
	Outcome1      string  `json:"outcome1"`
	Bookmaker1    string  `json:"bookmaker1"`
	Odds1         float64 `json:"odds1"`
	Outcome2      string  `json:"outcome2"`
	Bookmaker2    string  `json:"bookmaker2"`
	Odds2         float64 `json:"odds2"`
	ProfitPercent float64 `json:"profitPercent"`
}

// CheckArbitrage evalúa si dos cuotas opuestas representan un arbitraje matemático
func CheckArbitrage(bm1 string, o1 float64, name1 string, bm2 string, o2 float64, name2 string) ArbitrageOpp {
	if o1 <= 1.0 || o2 <= 1.0 {
		return ArbitrageOpp{IsArbitrage: false}
	}

	// Probabilidades implícitas
	p1 := 1.0 / o1
	p2 := 1.0 / o2

	totalP := p1 + p2

	if totalP < 0.995 { // Margen de ganancia matemática (menos del 99.5% de probabilidad implícita combinada)
		profitPct := (1.0/totalP - 1.0) * 100
		return ArbitrageOpp{
			IsArbitrage:   true,
			Outcome1:      name1,
			Bookmaker1:    bm1,
			Odds1:         o1,
			Outcome2:      name2,
			Bookmaker2:    bm2,
			Odds2:         o2,
			ProfitPercent: profitPct,
		}
	}

	return ArbitrageOpp{IsArbitrage: false}
}

// CalculateArbitrageStakes calcula la división exacta de apuestas para un arbitraje
func CalculateArbitrageStakes(totalStake, odds1, odds2 float64) (float64, float64) {
	p1 := 1.0 / odds1
	p2 := 1.0 / odds2
	totalP := p1 + p2

	stake1 := totalStake * (p1 / totalP)
	stake2 := totalStake - stake1

	return stake1, stake2
}
