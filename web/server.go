package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sports-betting-bot/config"
	"sports-betting-bot/db"
	"sports-betting-bot/engine"
	"sports-betting-bot/strategy"
)

// StartWebServer levanta el servidor HTTP de alta velocidad para servir el panel y los Server-Sent Events
func StartWebServer(cfg *config.Config) {
	// Servir archivos estáticos (CSS, JS)
	staticDir := filepath.Join("web", "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Ruta principal (Panel Web)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlPath := filepath.Join("web", "templates", "index.html")
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			// Si no existe, servir una plantilla integrada rápida
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<h3>Servidor Go activo. Creando archivos de plantilla... Refresca en 2 segundos.</h3>")
			return
		}
		http.ServeFile(w, r, htmlPath)
	})

	// Endpoints REST de la API
	http.HandleFunc("/api/history", handleGetHistory)
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		handleGetStats(w, r, cfg.InitialBankroll)
	})
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		handleConfig(w, r, cfg)
	})
	http.HandleFunc("/api/retrain", handleRetrain)
	http.HandleFunc("/api/scanned", handleGetScanned)
	http.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		handleReset(w, r, cfg)
	})

	// Canal de Server-Sent Events (SSE) para tiempo real ultra-rápido de logs y métricas
	http.HandleFunc("/events", handleSSE(cfg.InitialBankroll))

	addr := ":" + cfg.Port
	engine.AddLog("Servidor Web iniciado en http://localhost%s", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error al iniciar el servidor web: %v\n", err)
	}
}

func handleGetHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	history, err := db.GetHistory()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(history)
}

func handleGetStats(w http.ResponseWriter, r *http.Request, initialBankroll float64) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stats, err := db.GetStats(initialBankroll)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// handleSSE transmite logs y datos en tiempo real al panel mediante SSE nativo del navegador
func handleSSE(initialBankroll float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Encabezados SSE obligatorios
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Verificar que el ResponseWriter soporte streaming (Flusher)
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "El servidor no soporta streaming de eventos", http.StatusInternalServerError)
			return
		}

		// Enviar historial inicial de logs para poblar la pantalla
		logs := engine.GetLatestLogs()
		for _, logLine := range logs {
			fmt.Fprintf(w, "event: log\ndata: %s\n\n", logLine)
		}
		flusher.Flush()

		// Enviar estadísticas iniciales
		sendCurrentStats(w, flusher, initialBankroll)

		// Loop de escucha
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case logMsg := <-engine.LogChan:
				// Enviar log nuevo
				fmt.Fprintf(w, "event: log\ndata: %s\n\n", logMsg)
				flusher.Flush()
			case <-ticker.C:
				// Actualizar estadísticas cada 2 segundos
				sendCurrentStats(w, flusher, initialBankroll)
			case <-r.Context().Done():
				// Conexión del cliente cerrada
				return
			}
		}
	}
}

func sendCurrentStats(w http.ResponseWriter, flusher http.Flusher, initialBankroll float64) {
	stats, err := db.GetStats(initialBankroll)
	if err != nil {
		return
	}

	statsBytes, err := json.Marshal(stats)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: stats\ndata: %s\n\n", string(statsBytes))
	flusher.Flush()
}

func handleConfig(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodGet {
		res := map[string]interface{}{
			"apiKey":          cfg.APIKey,
			"initialBankroll": cfg.InitialBankroll,
			"kellyFraction":   cfg.KellyFraction,
			"simulationMode":  cfg.SimulationMode,
			"sportsToScan":    strings.Join(cfg.SportsToScan, ","),
			"minEVThreshold":  cfg.MinEVThreshold,
			"sharpBookmaker":  cfg.SharpBookmaker,
			"minOdds":         cfg.MinOdds,
			"maxOdds":         cfg.MaxOdds,
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	var req struct {
		APIKey          string  `json:"apiKey"`
		InitialBankroll float64 `json:"initialBankroll"`
		KellyFraction   float64 `json:"kellyFraction"`
		SimulationMode  bool    `json:"simulationMode"`
		SportsToScan    string  `json:"sportsToScan"`
		MinEVThreshold  float64 `json:"minEVThreshold"`
		SharpBookmaker  string  `json:"sharpBookmaker"`
		MinOdds         float64 `json:"minOdds"`
		MaxOdds         float64 `json:"maxOdds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
		return
	}

	// Actualizar en caliente
	cfg.APIKey = req.APIKey
	cfg.InitialBankroll = req.InitialBankroll
	cfg.KellyFraction = req.KellyFraction
	cfg.SimulationMode = req.SimulationMode

	// Parsear deportes
	var sports []string
	if req.SportsToScan != "" {
		parts := strings.Split(req.SportsToScan, ",")
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				sports = append(sports, t)
			}
		}
	}
	if len(sports) == 0 {
		sports = []string{"upcoming"}
	}
	cfg.SportsToScan = sports

	cfg.MinEVThreshold = req.MinEVThreshold
	cfg.SharpBookmaker = req.SharpBookmaker
	cfg.MinOdds = req.MinOdds
	cfg.MaxOdds = req.MaxOdds

	saveConfigToEnv(cfg)

	engine.AddLog("⚙️ Configuración del Bot y Estrategia actualizada en caliente.")
	if cfg.APIKey == "" {
		engine.AddLog("[Simulación Activa] Operando con generador de cuotas dinámicas de alta fidelidad.")
	} else {
		engine.AddLog("[API Real Activa] Conectado a The Odds API para cuotas y resultados.")
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "config": req})
}

func saveConfigToEnv(cfg *config.Config) {
	var sportsStr string
	if len(cfg.SportsToScan) > 0 {
		sportsStr = strings.Join(cfg.SportsToScan, ",")
	} else {
		sportsStr = "upcoming"
	}
	content := fmt.Sprintf(
		"THE_ODDS_API_KEY=\"%s\"\n"+
		"PORT=%s\n"+
		"INITIAL_BANKROLL=%.2f\n"+
		"KELLY_FRACTION=%.2f\n"+
		"SIMULATION_MODE=%t\n"+
		"SPORTS_TO_SCAN=\"%s\"\n"+
		"MIN_EV_THRESHOLD=%.2f\n"+
		"SHARP_BOOKMAKER=\"%s\"\n"+
		"MIN_ODDS=%.2f\n"+
		"MAX_ODDS=%.2f\n",
		cfg.APIKey,
		cfg.Port,
		cfg.InitialBankroll,
		cfg.KellyFraction,
		cfg.SimulationMode,
		sportsStr,
		cfg.MinEVThreshold,
		cfg.SharpBookmaker,
		cfg.MinOdds,
		cfg.MaxOdds,
	)
	_ = os.WriteFile(".env", []byte(content), 0644)
}

func handleRetrain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	history, err := db.GetHistory()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	report := strategy.RetrainModel(history.Bets)

	// Registrar en la consola del bot el re-entrenamiento exitoso
	engine.AddLog("🤖 [Modelo Optimizado] Grid Search finalizado. Regla estrella: %s. Umbral EV: %.1f%%. Kelly óptimo: %.2f", 
		report.BestPerformingRule, report.OptimalEVThreshold, report.OptimalKellyFraction)

	json.NewEncoder(w).Encode(report)
}

func handleGetScanned(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	scanned := engine.GetScannedEvents()
	json.NewEncoder(w).Encode(scanned)
}

func handleReset(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	err := db.ResetDatabase(cfg.InitialBankroll)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	engine.AddLog("🗑️ Base de datos e historial de apuestas reseteados exitosamente.")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Historial borrado"})
}
