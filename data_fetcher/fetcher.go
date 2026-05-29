package data_fetcher

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type Outcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Market struct {
	Key      string    `json:"key"`
	Outcomes []Outcome `json:"outcomes"`
}

type Bookmaker struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	LastUpdate string   `json:"last_update"`
	Markets    []Market `json:"markets"`
}

type SportsEvent struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	SportTitle   string      `json:"sport_title"`
	CommenceTime string      `json:"commence_time"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	Bookmakers   []Bookmaker `json:"bookmakers"`
}

// FetchOdds obtiene cuotas de la API real o genera datos simulados de alta fidelidad si la API Key no está configurada
func FetchOdds(apiKey string) ([]SportsEvent, error) {
	if apiKey == "" {
		return GenerateMockOdds(), nil
	}

	url := fmt.Sprintf("https://api.the-odds-api.com/v4/sports/upcoming/odds/?regions=eu,us&markets=h2h&apiKey=%s", apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con The Odds API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("la API respondió con código de estado de error: %d", resp.StatusCode)
	}

	var events []SportsEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta de la API: %w", err)
	}

	return events, nil
}

// GenerateMockOdds genera partidos simulados de fútbol europeo y NBA con cuotas dinámicas y fluctuantes
func GenerateMockOdds() []SportsEvent {
	rand.Seed(time.Now().UnixNano())

	// Definir algunos partidos mock
	matchups := []struct {
		sport      string
		sportTitle string
		home       string
		away       string
	}{
		{"soccer_uefa_champions", "Champions League", "Real Madrid", "Manchester City"},
		{"soccer_uefa_champions", "Champions League", "Barcelona", "Bayern Munich"},
		{"basketball_nba", "NBA Basketball", "Los Angeles Lakers", "Boston Celtics"},
		{"basketball_nba", "NBA Basketball", "Golden State Warriors", "Miami Heat"},
		{"soccer_spain_la_liga", "La Liga Spain", "Atletico Madrid", "Sevilla"},
	}

	bookmakers := []string{"Bet365", "Pinnacle", "Polymarket", "DraftKings", "Bwin"}
	events := make([]SportsEvent, len(matchups))

	for i, m := range matchups {
		eventID := fmt.Sprintf("mock_match_%d", i+1)
		
		// Crear variabilidad de cuotas base según el partido
		var baseHomeOdds, baseAwayOdds float64
		if i == 0 { // Madrid vs City (Muy parejo)
			baseHomeOdds = 2.40 + (rand.Float64() * 0.20 - 0.10)
			baseAwayOdds = 2.60 + (rand.Float64() * 0.20 - 0.10)
		} else if i == 1 { // Barca vs Bayern (Bayern favorito)
			baseHomeOdds = 3.10 + (rand.Float64() * 0.40 - 0.20)
			baseAwayOdds = 1.95 + (rand.Float64() * 0.15 - 0.08)
		} else { // Otros partidos
			baseHomeOdds = 1.80 + (rand.Float64() * 0.30 - 0.15)
			baseAwayOdds = 2.10 + (rand.Float64() * 0.30 - 0.15)
		}

		bms := make([]Bookmaker, len(bookmakers))
		for j, bmTitle := range bookmakers {
			// Introducir ligeras diferencias entre casas para simular oportunidades de arbitraje ocasionales
			// Especialmente en Polymarket, las cuotas suelen fluctuar más rápido por la opinión pública
			var homeVar, awayVar float64
			if bmTitle == "Polymarket" {
				homeVar = (rand.Float64() * 0.35) - 0.15 // Fluctuación mayor
				awayVar = (rand.Float64() * 0.35) - 0.15
			} else {
				homeVar = (rand.Float64() * 0.12) - 0.06
				awayVar = (rand.Float64() * 0.12) - 0.06
			}

			homeOdds := baseHomeOdds + homeVar
			awayOdds := baseAwayOdds + awayVar

			// Asegurarse de que las cuotas no bajen de 1.05
			if homeOdds < 1.05 { homeOdds = 1.05 }
			if awayOdds < 1.05 { awayOdds = 1.05 }

			outcomes := []Outcome{
				{Name: m.home, Price: truncate(homeOdds)},
				{Name: m.away, Price: truncate(awayOdds)},
			}

			// Añadir empate para fútbol
			if m.sport != "basketball_nba" {
				baseDrawOdds := 3.20 + (rand.Float64() * 0.40 - 0.20)
				outcomes = append(outcomes, Outcome{Name: "Draw", Price: truncate(baseDrawOdds)})
			}

			bms[j] = Bookmaker{
				Key:        bmTitle,
				Title:      bmTitle,
				LastUpdate: time.Now().Format(time.RFC3339),
				Markets: []Market{
					{
						Key:      "h2h",
						Outcomes: outcomes,
					},
				},
			}
		}

		events[i] = SportsEvent{
			ID:           eventID,
			SportKey:     m.sport,
			SportTitle:   m.sportTitle,
			CommenceTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			HomeTeam:     m.home,
			AwayTeam:     m.away,
			Bookmakers:   bms,
		}
	}

	return events
}

func truncate(f float64) float64 {
	return float64(int(f*100)) / 100
}
