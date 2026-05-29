package main

import (
	"log"

	"sports-betting-bot/config"
	"sports-betting-bot/db"
	"sports-betting-bot/engine"
	"sports-betting-bot/web"
)

func main() {
	log.Println("Iniciando aplicación Antigravity Sports Trading Bot...")

	// 1. Cargar Configuración
	cfg := config.LoadConfig()

	// 2. Inicializar Base de Datos JSON
	err := db.InitDatabase(cfg.InitialBankroll)
	if err != nil {
		log.Fatalf("Error crítico al inicializar base de datos: %v", err)
	}

	// 3. Arrancar Motor del Bot en segundo plano
	engine.StartBotEngine(cfg)

	// 4. Iniciar Servidor Web y de Eventos (bloqueante)
	web.StartWebServer(cfg)
}
