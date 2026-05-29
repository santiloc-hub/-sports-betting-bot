package engine

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	mrand "math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"sports-betting-bot/config"
	"sports-betting-bot/data_fetcher"
	"sports-betting-bot/db"
	"sports-betting-bot/strategy"
)

var (
	activityLogs []string
	logMu        sync.Mutex
	LogChan      = make(chan string, 100)
)

// AddLog agrega un log a la consola interna y lo envía al canal WebSocket/SSE
func AddLog(format string, v ...interface{}) {
	msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, v...))
	log.Println(msg)

	logMu.Lock()
	activityLogs = append(activityLogs, msg)
	if len(activityLogs) > 100 {
		activityLogs = activityLogs[1:] // Limitar a los últimos 100 logs
	}
	logMu.Unlock()

	// Enviar al canal sin bloquear si está lleno
	select {
	case LogChan <- msg:
	default:
	}
}

func GetLatestLogs() []string {
	logMu.Lock()
	defer logMu.Unlock()
	
	// Copiar logs
	logs := make([]string, len(activityLogs))
	copy(logs, activityLogs)
	return logs
}

func generateID() string {
	b := make([]byte, 4)
	crand.Read(b)
	return hex.EncodeToString(b)
}

// StartBotEngine arranca el escaneo y el motor en segundo plano de forma asíncrona
func StartBotEngine(cfg *config.Config) {
	AddLog("Iniciando motor de trading deportivo en Go...")
	AddLog("Modo Simulación: %v | Coeficiente de Kelly: %.2f", cfg.SimulationMode, cfg.KellyFraction)
	
	if cfg.APIKey == "" {
		AddLog("[Simulación Activa] No se detectó API Key. Usando generador de cuotas dinámicas de alta fidelidad.")
	} else {
		AddLog("[API Real Activa] Conectado a The Odds API.")
	}

	// Loop principal del bot (ejecución cada 8 segundos)
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			events, err := data_fetcher.FetchOdds(cfg.APIKey)
			if err != nil {
				AddLog("Error al obtener cuotas: %v", err)
				continue
			}

			AddLog("Escaneando %d eventos deportivos en tiempo real...", len(events))
			processEvents(events, cfg)
			
			// Si estamos usando datos simulados, resolver apuestas pendientes de forma acelerada (cada ronda)
			// para que el usuario pueda ver resultados y el análisis histórico en pocos segundos de prueba.
			if cfg.APIKey == "" {
				resolvePendingBetsSimulated()
			} else {
				// Resolver apuestas basadas en eventos reales consultando scores reales
				resolvePendingBetsReal(cfg.APIKey)
			}
		}
	}()
}

// processEvents escanea cuotas para detectar arbitrajes o apuestas de valor
func processEvents(events []data_fetcher.SportsEvent, cfg *config.Config) {
	history, err := db.GetHistory()
	if err != nil {
		AddLog("Error al leer base de datos: %v", err)
		return
	}

	for _, ev := range events {
		// Validar que tengamos al menos 2 casas de apuestas para comparar
		if len(ev.Bookmakers) < 2 {
			continue
		}

		// 1. ANÁLISIS DE APUESTAS DE VALOR (+EV)
		// Tomamos Pinnacle o el promedio general como "cuota justa/verdadera"
		// y buscamos si Polymarket u otra casa ofrece un precio mucho mayor (outlier).
		var sharpHomeOdds, sharpAwayOdds float64
		var sharpBM string

		// Intentamos buscar Pinnacle como casa de referencia sharp. Si no está, usamos la primera disponible.
		for _, bm := range ev.Bookmakers {
			if bm.Key == "Pinnacle" || sharpBM == "" {
				for _, m := range bm.Markets {
					if m.Key == "h2h" && len(m.Outcomes) >= 2 {
						sharpHomeOdds = m.Outcomes[0].Price
						sharpAwayOdds = m.Outcomes[1].Price
						sharpBM = bm.Title
					}
				}
			}
		}

		if sharpHomeOdds <= 1.0 {
			continue
		}

		// Probabilidades implícitas reales (estimación del modelo ajustado)
		trueHomeProb := 1.0 / sharpHomeOdds
		trueAwayProb := 1.0 / sharpAwayOdds
		totalProb := trueHomeProb + trueAwayProb
		trueHomeProb /= totalProb
		trueAwayProb /= totalProb

		// Pool de Decisiones: Aplicar reglas para estimar p en base al perfil del evento (Fútbol / Tenis)
		ruleHome := "CONSENSUS_EV"
		ruleAway := "CONSENSUS_EV"

		// Si es simulación local de Tenis, aplicamos reglas estadísticas dedicadas
		if cfg.APIKey == "" && (ev.SportKey == "tennis_atp" || ev.SportKey == "tennis_wta") {
			if ev.HomeTeam == "Novak Djokovic" && ev.AwayTeam == "Daniil Medvedev" {
				// Novak Djokovic domina en H2H a Medvedev
				trueHomeProb = 0.70
				trueAwayProb = 0.30
				ruleHome = "H2H_DOMINANCE"
				ruleAway = "H2H_DOMINANCE"
			} else if ev.HomeTeam == "Carlos Alcaraz" && ev.AwayTeam == "Jannik Sinner" {
				// Alcaraz es especialista en arcilla vs Sinner
				trueHomeProb = 0.65
				trueAwayProb = 0.35
				ruleHome = "SURFACE_SPECIALIST"
				ruleAway = "SURFACE_SPECIALIST"
			} else if ev.HomeTeam == "Iga Swiatek" && ev.AwayTeam == "Aryna Sabalenka" {
				// Swiatek es especialista suprema en arcilla vs Sabalenka
				trueHomeProb = 0.73
				trueAwayProb = 0.27
				ruleHome = "SURFACE_SPECIALIST"
				ruleAway = "SURFACE_SPECIALIST"
			}
		}

		// Comparar con el resto de casas (especialmente Polymarket) para buscar valor
		for _, bm := range ev.Bookmakers {
			if bm.Key == sharpBM {
				continue // No comparar con ella misma
			}

			for _, m := range bm.Markets {
				if m.Key == "h2h" && len(m.Outcomes) >= 2 {
					oddsHome := m.Outcomes[0].Price
					oddsAway := m.Outcomes[1].Price

					// Evaluar Victoria Local (Home)
					evalEV(ev, bm.Title, ev.HomeTeam, oddsHome, trueHomeProb, history.CurrentBankroll, ruleHome, cfg)
					// Evaluar Victoria Visitante (Away)
					evalEV(ev, bm.Title, ev.AwayTeam, oddsAway, trueAwayProb, history.CurrentBankroll, ruleAway, cfg)
				}
			}
		}

		// 2. DETECCIÓN DE ARBITRAJE MATEMÁTICO (SUREBETS)
		// Comparamos todas las casas entre sí para ver si podemos asegurar ganancia
		for i := 0; i < len(ev.Bookmakers); i++ {
			for j := i + 1; j < len(ev.Bookmakers); j++ {
				bm1 := ev.Bookmakers[i]
				bm2 := ev.Bookmakers[j]

				var o1Home, o2Away float64
				for _, m := range bm1.Markets {
					if m.Key == "h2h" && len(m.Outcomes) >= 2 {
						o1Home = m.Outcomes[0].Price
					}
				}
				for _, m := range bm2.Markets {
					if m.Key == "h2h" && len(m.Outcomes) >= 2 {
						o2Away = m.Outcomes[1].Price
					}
				}

				if o1Home > 1.0 && o2Away > 1.0 {
					opp := strategy.CheckArbitrage(bm1.Title, o1Home, ev.HomeTeam, bm2.Title, o2Away, ev.AwayTeam)
					if opp.IsArbitrage {
						AddLog("🔥 ¡ARBITRAJE DETECTADO! [%s vs %s] - ROI: %.2f%%", ev.HomeTeam, ev.AwayTeam, opp.ProfitPercent)
						AddLog("   Opción 1: %s en %s @ %.2f | Opción 2: %s en %s @ %.2f", 
							opp.Outcome1, opp.Bookmaker1, opp.Odds1, opp.Outcome2, opp.Bookmaker2, opp.Odds2)
					}
				}
			}
		}
	}
}

func evalEV(ev data_fetcher.SportsEvent, bmTitle, outcomeName string, odds, trueProb, currentBankroll float64, ruleName string, cfg *config.Config) {
	// Calcular fracción de Kelly
	kellyF := strategy.CalculateKelly(trueProb, odds, cfg.KellyFraction)
	
	if kellyF > 0.005 { // Si la apuesta sugerida es mayor al 0.5% de la banca, hay valor claro (+EV)
		evPercent := (trueProb*odds - 1.0) * 100
		stake := currentBankroll * kellyF

		// Asegurar apuesta mínima razonable
		if stake < 5.0 {
			stake = 5.0
		}
		if stake > currentBankroll {
			return // No hay fondos
		}

		// Evitar duplicar apuestas si ya está pendiente en el historial
		if isAlreadyPending(ev.ID, outcomeName) {
			return
		}

		// Registrar apuesta simulada
		bet := &db.Bet{
			ID:             generateID(),
			EventID:        ev.ID,
			EventName:      fmt.Sprintf("%s vs %s", ev.HomeTeam, ev.AwayTeam),
			Sport:          ev.SportTitle,
			League:         ev.SportKey,
			Bookmaker:      bmTitle,
			Outcome:        outcomeName,
			Odds:           odds,
			Stake:          truncate(stake),
			Status:         "PENDING",
			PredictionRule: ruleName,
			EstimatedProb:  trueProb,
		}

		if err := db.AddBet(bet); err != nil {
			AddLog("Error al guardar apuesta simulada: %v", err)
			return
		}

		AddLog("💸 +EV Detectado en [%s] en %s @ %.2f (Prob: %.1f%% | Ventaja: +.1f%%)", 
			bet.EventName, bet.Bookmaker, bet.Odds, trueProb*100, evPercent)
		AddLog("   -> Apuesta Simulada Registrada: $%.2f USD en '%s' (Kelly: %.1f%%)", 
			bet.Stake, bet.Outcome, kellyF*100)
	}
}

func isAlreadyPending(eventID, outcome string) bool {
	h, err := db.GetHistory()
	if err != nil {
		return false
	}
	for _, b := range h.Bets {
		if b.EventID == eventID && b.Outcome == outcome && b.Status == "PENDING" {
			return true
		}
	}
	return false
}

// resolvePendingBetsSimulated simula el final de los partidos mock cada ronda (acelerado)
func resolvePendingBetsSimulated() {
	h, err := db.GetHistory()
	if err != nil {
		return
	}

	for _, b := range h.Bets {
		if b.Status == "PENDING" && strings.HasPrefix(b.EventID, "mock_match_") {
			// Calcular probabilidad de ganar aproximada según la cuota
			winProb := 1.0 / b.Odds
			
			// Ajustar margen de la simulación
			randVal := mrand.Float64()
			won := randVal < winProb

			err := db.ResolveBet(b.ID, won)
			if err != nil {
				AddLog("Error al resolver apuesta mock %s: %v", b.ID, err)
				continue
			}

			if won {
				AddLog("✅ ¡APUESTA SIMULADA GANADA! [%s] -> Simulación recupera $%.2f USD (Neto: +$%.2f @ %.2f en %s)", 
					b.EventName, b.Stake*b.Odds, b.Stake*(b.Odds-1), b.Odds, b.Bookmaker)
			} else {
				AddLog("❌ ¡APUESTA SIMULADA PERDIDA! [%s] -> Simulación pierde -$%.2f USD (@ %.2f en %s)", 
					b.EventName, b.Stake, b.Odds, b.Bookmaker)
			}
		}
	}
}

var lastScoresCheck time.Time

// resolvePendingBetsReal consulta la API de resultados reales para liquidar las apuestas pendientes reales
func resolvePendingBetsReal(apiKey string) {
	if apiKey == "" {
		return
	}

	h, err := db.GetHistory()
	if err != nil {
		return
	}

	// 1. Recopilar apuestas reales pendientes
	var pendingRealBets []db.Bet
	uniqueSports := make(map[string]bool)

	for _, b := range h.Bets {
		if b.Status == "PENDING" && !strings.HasPrefix(b.EventID, "mock_match_") {
			pendingRealBets = append(pendingRealBets, b)
			// Guardar el identificador del deporte (ej: basketball_nba)
			if b.League != "" {
				uniqueSports[b.League] = true
			}
		}
	}

	// Si no hay apuestas reales pendientes, no consultamos la API (ahorramos peticiones)
	if len(pendingRealBets) == 0 {
		return
	}

	// Limitar consultas a la API de scores a un máximo de una vez cada 2 minutos para no agotar la cuota API
	if time.Since(lastScoresCheck) < 2*time.Minute {
		return
	}
	lastScoresCheck = time.Now()

	AddLog("🔄 Consultando resultados reales de partidos para resolver %d apuestas pendientes...", len(pendingRealBets))

	// 2. Consultar puntuaciones por cada deporte único
	for sportKey := range uniqueSports {
		scores, err := data_fetcher.FetchScores(sportKey, apiKey)
		if err != nil {
			AddLog("Error al obtener resultados para %s: %v", sportKey, err)
			continue
		}

		// Crear mapa rápido para buscar resultados de eventos por ID
		eventMap := make(map[string]data_fetcher.EventScore)
		for _, s := range scores {
			eventMap[s.ID] = s
		}

		// 3. Procesar las apuestas correspondientes a este deporte
		for _, b := range pendingRealBets {
			if b.League != sportKey {
				continue
			}

			evScore, exists := eventMap[b.EventID]
			if !exists {
				continue // Partido aún no finalizado o no reportado en las últimas 72 horas
			}

			if !evScore.Completed {
				continue // El partido sigue en juego o aún no comienza
			}

			// El partido ha finalizado, resolver apuesta según puntuación real
			if len(evScore.Scores) < 2 {
				continue
			}

			// Extraer puntajes
			scoreHomeVal := 0
			scoreAwayVal := 0
			for _, teamScore := range evScore.Scores {
				scoreInt, _ := strconv.Atoi(teamScore.Score)
				if teamScore.Name == evScore.HomeTeam {
					scoreHomeVal = scoreInt
				} else if teamScore.Name == evScore.AwayTeam {
					scoreAwayVal = scoreInt
				}
			}

			// Determinar resultado real
			var winner string
			if scoreHomeVal > scoreAwayVal {
				winner = evScore.HomeTeam
			} else if scoreAwayVal > scoreHomeVal {
				winner = evScore.AwayTeam
			} else {
				winner = "Draw" // Empate
			}

			// Evaluar si la apuesta fue ganadora
			won := false
			if b.Outcome == winner {
				won = true
			} else if (b.Outcome == "Draw" || b.Outcome == "Empate") && winner == "Draw" {
				won = true
			}

			// Resolver apuesta en la base de datos
			err = db.ResolveBet(b.ID, won)
			if err != nil {
				AddLog("Error al liquidar apuesta real %s: %v", b.ID, err)
				continue
			}

			if won {
				AddLog("🎉 ✅ ¡APUESTA REAL GANADA! [%s] -> Simulación recupera $%.2f USD reales (Neto: +$%.2f @ %.2f en %s)", 
					b.EventName, b.Stake*b.Odds, b.Stake*(b.Odds-1), b.Odds, b.Bookmaker)
			} else {
				AddLog("💔 ❌ ¡APUESTA REAL PERDIDA! [%s] -> Simulación pierde -$%.2f USD reales (@ %.2f en %s)", 
					b.EventName, b.Stake, b.Odds, b.Bookmaker)
			}
		}
	}
}

func truncate(f float64) float64 {
	return float64(int(f*100)) / 100
}
