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
func FetchOdds(apiKey string, sports []string) ([]SportsEvent, error) {
	if apiKey == "" {
		return GenerateMockOdds(), nil
	}

	var allEvents []SportsEvent
	client := &http.Client{Timeout: 5 * time.Second}

	for _, sport := range sports {
		var url string
		if sport == "upcoming" {
			url = fmt.Sprintf("https://api.the-odds-api.com/v4/sports/upcoming/odds/?regions=eu,us&markets=h2h&apiKey=%s", apiKey)
		} else {
			url = fmt.Sprintf("https://api.the-odds-api.com/v4/sports/%s/odds/?regions=eu,us&markets=h2h&apiKey=%s", sport, apiKey)
		}

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error al conectar con The Odds API para %s: %w", sport, err)
		}

		if resp.StatusCode == http.StatusOK {
			var events []SportsEvent
			if err := json.NewDecoder(resp.Body).Decode(&events); err == nil {
				allEvents = append(allEvents, events...)
			} else {
				resp.Body.Close()
				return nil, fmt.Errorf("error al decodificar respuesta de la API para %s: %w", sport, err)
			}
		} else {
			resp.Body.Close()
			return nil, fmt.Errorf("la API respondió con código de estado de error %d para %s", resp.StatusCode, sport)
		}
		resp.Body.Close()
	}

	return allEvents, nil
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
		{"tennis_atp", "ATP Tour - Clay", "Carlos Alcaraz", "Jannik Sinner"},
		{"tennis_atp", "ATP Tour - Clay", "Novak Djokovic", "Daniil Medvedev"},
		{"tennis_wta", "WTA Tour - Clay", "Iga Swiatek", "Aryna Sabalenka"},
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
		} else if i == 2 { // Alcaraz vs Sinner (Alcaraz leve favorito en arcilla)
			baseHomeOdds = 1.72 + (rand.Float64() * 0.15 - 0.07)
			baseAwayOdds = 2.15 + (rand.Float64() * 0.20 - 0.10)
		} else if i == 3 { // Djokovic vs Medvedev (Djokovic favorito)
			baseHomeOdds = 1.45 + (rand.Float64() * 0.10 - 0.05)
			baseAwayOdds = 2.80 + (rand.Float64() * 0.30 - 0.15)
		} else { // Swiatek vs Sabalenka (Swiatek favorita en arcilla)
			baseHomeOdds = 1.55 + (rand.Float64() * 0.12 - 0.06)
			baseAwayOdds = 2.45 + (rand.Float64() * 0.25 - 0.12)
		}

		bms := make([]Bookmaker, len(bookmakers))
		for j, bmTitle := range bookmakers {
			// Introducir ligeras diferencias entre casas para simular oportunidades de arbitraje ocasionales
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

			// Añadir empate únicamente para fútbol
			if m.sport == "soccer_uefa_champions" || m.sport == "soccer_spain_la_liga" {
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

type Score struct {
	Name  string `json:"name"`
	Score string `json:"score"`
}

type EventScore struct {
	ID           string  `json:"id"`
	SportKey     string  `json:"sport_key"`
	Completed    bool    `json:"completed"`
	HomeTeam     string  `json:"home_team"`
	AwayTeam     string  `json:"away_team"`
	Scores       []Score `json:"scores"`
}

// FetchScores obtiene resultados reales de partidos desde The Odds API
func FetchScores(sport string, apiKey string) ([]EventScore, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("se requiere API Key para consultar resultados reales")
	}

	url := fmt.Sprintf("https://api.the-odds-api.com/v4/sports/%s/scores/?daysFrom=3&apiKey=%s", sport, apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con The Odds API scores: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("la API de resultados respondió con código de estado: %d", resp.StatusCode)
	}

	var scores []EventScore
	if err := json.NewDecoder(resp.Body).Decode(&scores); err != nil {
		return nil, fmt.Errorf("error al decodificar resultados de la API: %w", err)
	}

	return scores, nil
}

func truncate(f float64) float64 {
	return float64(int(f*100)) / 100
}
