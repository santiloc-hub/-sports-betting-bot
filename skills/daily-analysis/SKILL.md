---
name: daily-sports-analysis
description: Escanea los eventos de tenis y fútbol programados para el día, estima la probabilidad real cruzando datos de APIs estadísticas, y genera un portafolio diario de apuestas recomendadas (+EV).
---

# Skill: Análisis Diario de Apuestas (Fútbol y Tenis)

Esta skill proporciona las directrices lógicas y matemáticas para que el bot genere un **Reporte Diario de Oportunidades** antes de que comiencen los partidos, evaluando la probabilidad estadística real y comparándola con las cuotas del mercado.

---

## 📅 Flujo de Trabajo para el Análisis Diario

### 1. Escaneo Pre-Partido (Fútbol y Tenis)
Al inicio del día (o a solicitud del usuario en el panel), el bot debe escanear la lista de eventos programados para las próximas 24 horas en las siguientes ligas prioritarias:
- **Fútbol:** UEFA Champions League (`soccer_uefa_champions`), La Liga España (`soccer_spain_la_liga`), Premier League Inglaterra (`soccer__league_1`).
- **Tenis:** Torneos ATP y WTA (`tennis_atp`, `tennis_wta`).

### 2. Consulta de Probabilidad Estadística (Estadísticas Históricas)
En lugar de depender únicamente de Pinnacle para estimar la probabilidad implícita, el bot consultará APIs estadísticas externas (como **API-Sports** o un modelo de Poisson interno) para evaluar:
- **Head-to-Head (H2H):** Historial de enfrentamientos directos entre ambos equipos o jugadores de tenis.
- **Forma Reciente:** Rendimiento de los últimos 5 partidos (puntos obtenidos, sets ganados/perdidos, goles a favor/en contra).
- **Métricas Avanzadas (Fútbol):** Goles Esperados (xG), posesión de balón promedio y efectividad de tiros a puerta.
- **Métricas Avanzadas (Tenis):** Rendimiento en el tipo de superficie (arcilla, césped, dura), porcentaje de quiebres de servicio y efectividad del primer saque.

El bot consolidará estas métricas para obtener la **Probabilidad Verdadera Estimada ($p_{\text{modelo}}$)**.

### 3. Detección de Discrepancias en el "Pool" del Mercado
El bot comparará la $p_{\text{modelo}}$ con las cuotas disponibles en los diferentes corredores de apuestas o pools predictivos (ej. Polymarket).
- Si la cuota disponible ($Odds$) es mayor que $1 / p_{\text{modelo}}$, existe un **Valor Esperado Positivo (+EV)**:
  $$\text{EV} = (p_{\text{modelo}} \times \text{Cuota} - 1.0) \times 100$$
- Se calculará el stake exacto recomendado mediante el **Criterio de Kelly**.

### 4. Generación del Reporte Diario (`daily_report.json`)
El bot guardará toda esta investigación en un archivo estructurado `daily_report.json` con la siguiente estructura:

```json
{
  "date": "2026-05-29",
  "generatedAt": "2026-05-29T08:00:00Z",
  "summary": {
    "totalEventsAnalyzed": 24,
    "valueBetsFound": 3,
    "recommendedTotalExposure": "7.5%"
  },
  "recommendations": [
    {
      "eventId": "real_event_id_123",
      "sport": "Tennis",
      "eventName": "Carlos Alcaraz vs Novak Djokovic",
      "market": "Match Winner",
      "prediction": "Carlos Alcaraz",
      "modelProbability": "58.5%",
      "marketProbability": "52.0%",
      "bestOdds": 1.92,
      "bookmaker": "Polymarket",
      "expectedValue": "+12.32%",
      "recommendedStake": "$250.00 (2.5% de la banca)"
    }
  ]
}
```

---

## 📊 Integración en el Panel Web
El archivo JSON generado será parseado en el frontend del panel y se expondrá en una sección interactiva dedicada dentro de la pestaña **"Ajustes / Análisis Diario"** para que el usuario pueda tomar la decisión final de en qué partidos apostar.
