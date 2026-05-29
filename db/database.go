package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Bet struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	EventName     string    `json:"eventName"`
	Sport         string    `json:"sport"`
	League        string    `json:"league"`
	Bookmaker     string    `json:"bookmaker"`
	Outcome       string    `json:"outcome"`       // ej. "Real Madrid", "Empate", "Home Win"
	Odds          float64   `json:"odds"`          // Cuota decimal (ej. 2.10)
	Stake         float64   `json:"stake"`         // Monto apostado en USD
	Status        string    `json:"status"`        // "PENDING", "WON", "LOST"
	NetProfit     float64   `json:"netProfit"`     // Ganancia neta (positiva si gana, -stake si pierde)
	BankrollAfter float64   `json:"bankrollAfter"` // Balance restante tras registrar la apuesta
}

type History struct {
	CurrentBankroll float64 `json:"currentBankroll"`
	Bets            []Bet   `json:"bets"`
}

type Stats struct {
	CurrentBankroll float64 `json:"currentBankroll"`
	TotalBets       int     `json:"totalBets"`
	WonBets         int     `json:"wonBets"`
	LostBets        int     `json:"lostBets"`
	PendingBets     int     `json:"pendingBets"`
	WinRate         float64 `json:"winRate"`
	ROI             float64 `json:"roi"`
	Yield           float64 `json:"yield"`
	TotalProfit     float64 `json:"totalProfit"`
	MaxDrawdown     float64 `json:"maxDrawdown"`
}

var (
	historyFilePath = "history.json"
	mu              sync.RWMutex
)

// InitDatabase inicializa el archivo JSON si no existe
func InitDatabase(initialBankroll float64) error {
	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		h := &History{
			CurrentBankroll: initialBankroll,
			Bets:            []Bet{},
		}
		return saveHistoryToFile(h)
	}
	return nil
}

// GetHistory retorna todo el historial de manera segura
func GetHistory() (*History, error) {
	mu.RLock()
	defer mu.RUnlock()

	return loadHistoryFromFile()
}

// AddBet agrega una apuesta simulada y actualiza el balance de forma transaccional
func AddBet(bet *Bet) error {
	mu.Lock()
	defer mu.Unlock()

	h, err := loadHistoryFromFile()
	if err != nil {
		return err
	}

	// Descontar la apuesta de la banca en simulación si es pendiente
	h.CurrentBankroll -= bet.Stake
	bet.BankrollAfter = h.CurrentBankroll
	bet.Timestamp = time.Now()

	h.Bets = append(h.Bets, *bet)

	return saveHistoryToFile(h)
}

// ResolveBet liquida una apuesta pendiente y le otorga ganancias o pérdidas
func ResolveBet(id string, won bool) error {
	mu.Lock()
	defer mu.Unlock()

	h, err := loadHistoryFromFile()
	if err != nil {
		return err
	}

	found := false
	for i, b := range h.Bets {
		if b.ID == id && b.Status == "PENDING" {
			found = true
			if won {
				h.Bets[i].Status = "WON"
				h.Bets[i].NetProfit = b.Stake * (b.Odds - 1)
				// Devolver la apuesta + la ganancia neta a la banca
				h.CurrentBankroll += b.Stake * b.Odds
			} else {
				h.Bets[i].Status = "LOST"
				h.Bets[i].NetProfit = -b.Stake
				// No devolvemos nada (la banca ya tenía el stake descontado)
			}
			h.Bets[i].BankrollAfter = h.CurrentBankroll
			break
		}
	}

	if !found {
		return errors.New("apuesta pendiente no encontrada con el ID provisto")
	}

	return saveHistoryToFile(h)
}

// GetStats calcula las métricas de rendimiento basadas en el historial
func GetStats(initialBankroll float64) (*Stats, error) {
	mu.RLock()
	defer mu.RUnlock()

	h, err := loadHistoryFromFile()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		CurrentBankroll: h.CurrentBankroll,
		TotalBets:       len(h.Bets),
	}

	var totalInvested float64
	var maxBankroll float64 = initialBankroll
	var maxDrawdown float64 = 0.0

	currentRunBankroll := initialBankroll

	for _, b := range h.Bets {
		switch b.Status {
		case "WON":
			stats.WonBets++
			stats.TotalProfit += b.NetProfit
			totalInvested += b.Stake
			currentRunBankroll += b.NetProfit
		case "LOST":
			stats.LostBets++
			stats.TotalProfit += b.NetProfit
			totalInvested += b.Stake
			currentRunBankroll += b.NetProfit
		case "PENDING":
			stats.PendingBets++
			currentRunBankroll -= b.Stake
		}

		// Calcular Drawdown histórico
		if currentRunBankroll > maxBankroll {
			maxBankroll = currentRunBankroll
		}
		if maxBankroll > 0 {
			dd := (maxBankroll - currentRunBankroll) / maxBankroll * 100
			if dd > maxDrawdown {
				maxDrawdown = dd
			}
		}
	}

	resolvedCount := stats.WonBets + stats.LostBets
	if resolvedCount > 0 {
		stats.WinRate = float64(stats.WonBets) / float64(resolvedCount) * 100
	}

	if totalInvested > 0 {
		stats.ROI = (stats.TotalProfit / initialBankroll) * 100
		stats.Yield = (stats.TotalProfit / totalInvested) * 100
	}

	stats.MaxDrawdown = maxDrawdown

	return stats, nil
}

// Funciones auxiliares privadas (asumen bloqueo activo)

func loadHistoryFromFile() (*History, error) {
	file, err := os.Open(historyFilePath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo de base de datos: %w", err)
	}
	defer file.Close()

	var h History
	dec := json.NewDecoder(file)
	if err := dec.Decode(&h); err != nil {
		return nil, fmt.Errorf("error al decodificar datos del historial: %w", err)
	}

	return &h, nil
}

func saveHistoryToFile(h *History) error {
	file, err := os.OpenFile(historyFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error al guardar archivo de base de datos: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(h); err != nil {
		return fmt.Errorf("error al serializar datos del historial: %w", err)
	}

	return nil
}
