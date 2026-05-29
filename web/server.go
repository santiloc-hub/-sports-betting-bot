package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sports-betting-bot/config"
	"sports-betting-bot/db"
	"sports-betting-bot/engine"
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
		json.NewEncoder(w).Encode(map[string]string{"apiKey": cfg.APIKey})
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
		return
	}

	// Actualizar en caliente
	cfg.APIKey = req.APIKey
	saveAPIKeyToEnv(req.APIKey)

	engine.AddLog("🔑 API Key de The Odds API actualizada en vivo.")
	if req.APIKey == "" {
		engine.AddLog("[Simulación Activa] API Key removida. Operando con simulador acelerado.")
	} else {
		engine.AddLog("[API Real Activa] Conectado a The Odds API para cuotas y resultados.")
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "apiKey": req.APIKey})
}

func saveAPIKeyToEnv(apiKey string) {
	content := fmt.Sprintf("THE_ODDS_API_KEY=\"%s\"\nPORT=8080\nINITIAL_BANKROLL=10000.0\nKELLY_FRACTION=0.25\nSIMULATION_MODE=true\n", apiKey)
	_ = os.WriteFile(".env", []byte(content), 0644)
}
