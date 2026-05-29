package strategy

import (
	"sports-betting-bot/db"
)

type RuleStats struct {
	RuleName  string  `json:"ruleName"`
	TotalBets int     `json:"totalBets"`
	WonBets   int     `json:"wonBets"`
	Accuracy  float64 `json:"accuracy"`
	Profit    float64 `json:"profit"`
}

type OptimizationReport struct {
	TotalBetsAnalyzed      int         `json:"totalBetsAnalyzed"`
	RulesEvaluated         []RuleStats `json:"rulesEvaluated"`
	BestPerformingRule     string      `json:"bestPerformingRule"`
	OptimalEVThreshold     float64     `json:"optimalEVThreshold"`
	OptimalKellyFraction   float64     `json:"optimalKellyFraction"`
	EstimatedROIImprovement float64     `json:"estimatedROIImprovement"`
}

// RetrainModel ejecuta un análisis retrospectivo tipo "Grid Search" sobre el historial de apuestas
// para optimizar las reglas de decisión, el umbral de EV y la fracción del Criterio de Kelly.
func RetrainModel(bets []db.Bet) OptimizationReport {
	report := OptimizationReport{
		TotalBetsAnalyzed:    0,
		RulesEvaluated:       []RuleStats{},
		BestPerformingRule:   "CONSENSUS_EV",
		OptimalEVThreshold:   1.5,
		OptimalKellyFraction: 0.25,
	}

	// 1. Agrupar estadísticas por regla de predicción
	ruleMap := make(map[string]*RuleStats)
	completedCount := 0

	for _, b := range bets {
		if b.Status == "PENDING" {
			continue
		}
		completedCount++

		rule := b.PredictionRule
		if rule == "" {
			rule = "CONSENSUS_EV"
		}

		if _, exists := ruleMap[rule]; !exists {
			ruleMap[rule] = &RuleStats{
				RuleName: rule,
			}
		}

		stats := ruleMap[rule]
		stats.TotalBets++
		if b.Status == "WON" {
			stats.WonBets++
			stats.Profit += b.NetProfit
		} else if b.Status == "LOST" {
			stats.Profit += b.NetProfit
		}
	}

	report.TotalBetsAnalyzed = completedCount

	// Si no hay suficientes datos para re-entrenar, retornar valores base
	if completedCount < 5 {
		// Inicializar estadísticas vacías de referencia para la UI
		report.RulesEvaluated = []RuleStats{
			{RuleName: "H2H_DOMINANCE", TotalBets: 0, WonBets: 0, Accuracy: 0, Profit: 0},
			{RuleName: "SURFACE_SPECIALIST", TotalBets: 0, WonBets: 0, Accuracy: 0, Profit: 0},
			{RuleName: "CONSENSUS_EV", TotalBets: 0, WonBets: 0, Accuracy: 0, Profit: 0},
		}
		return report
	}

	// 2. Calcular precisión de cada regla
	var bestRule string
	var maxProfit float64 = -999999.0

	for name, stats := range ruleMap {
		if stats.TotalBets > 0 {
			stats.Accuracy = (float64(stats.WonBets) / float64(stats.TotalBets)) * 100
		}
		report.RulesEvaluated = append(report.RulesEvaluated, *stats)

		// Buscar la mejor regla según beneficios netos
		if stats.Profit > maxProfit {
			maxProfit = stats.Profit
			bestRule = name
		}
	}

	if bestRule != "" {
		report.BestPerformingRule = bestRule
	}

	// 3. Simular "Grid Search" rápido de hiperparámetros sobre el histórico
	// Umbrales de EV a evaluar: 0.5%, 1.0%, 1.5%, 2.0%, 2.5%, 3.0%
	// Fracciones de Kelly a evaluar: 0.10, 0.20, 0.25, 0.30, 0.50
	var bestEV float64 = 1.0
	var bestKelly float64 = 0.25
	var maxSimulatedProfit float64 = -99999.0

	evThresholds := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0}
	kellyFractions := []float64{0.10, 0.20, 0.25, 0.30, 0.50}

	for _, ev := range evThresholds {
		for _, kf := range kellyFractions {
			simulatedProfit := 0.0
			for _, b := range bets {
				if b.Status == "PENDING" {
					continue
				}

				// Estimar el EV original a partir del logaritmo o aproximado
				oddsNeta := b.Odds - 1.0
				betEV := (b.EstimatedProb*b.Odds - 1.0) * 100

				// Solo "apostar" en la simulación si superamos el umbral de EV evaluado
				if betEV >= ev {
					// Recalcular stake simulado según la fracción evaluada
					// Kelly original = p - (q / b)
					q := 1.0 - b.EstimatedProb
					kellyOriginal := b.EstimatedProb - (q / oddsNeta)
					if kellyOriginal <= 0 {
						continue
					}
					simulatedStake := 1000.0 * kellyOriginal * kf // banca base de $1,000 para backtest

					if b.Status == "WON" {
						simulatedProfit += simulatedStake * oddsNeta
					} else {
						simulatedProfit -= simulatedStake
					}
				}
			}

			if simulatedProfit > maxSimulatedProfit {
				maxSimulatedProfit = simulatedProfit
				bestEV = ev
				bestKelly = kf
			}
		}
	}

	report.OptimalEVThreshold = bestEV
	report.OptimalKellyFraction = bestKelly
	report.EstimatedROIImprovement = (maxSimulatedProfit - maxProfit) / 1000.0 * 100

	if report.EstimatedROIImprovement < 0 {
		report.EstimatedROIImprovement = 0
	}

	return report
}
