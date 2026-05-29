---
name: simulate-bets-sports-bot
description: Ejecuta simulaciones del motor de apuestas, evalúa el balance financiero del bot (ROI, Yield, Max Drawdown) y ajusta los parámetros del Criterio de Kelly.
---

# Skill: Simulación Matemática y Ajuste del Criterio de Kelly

Esta skill define el flujo de trabajo para monitorear el desempeño del bot, simular escenarios con cuotas sintéticas o reales, y refinar la estrategia matemática para optimizar el crecimiento de la banca y reducir el riesgo de quiebra.

---

## 📐 Parámetros de la Estrategia a Monitorear

### 1. Criterio de Kelly
El tamaño de la apuesta está determinado por la fórmula:
$$f^* = \frac{p \cdot b - q}{b} = \frac{p(b + 1) - 1}{b}$$
Donde:
- $f^*$: Fracción de la banca actual a apostar.
- $p$: Probabilidad real estimada de ganar (calculada a partir de nuestras métricas o valor esperado).
- $b$: Cuota neta recibida (Cuota Decimal - 1).
- $q$: Probabilidad de perder ($1 - p$).

*Nota: Se recomienda utilizar **Kelly Fraccionario** (por ejemplo, 0.25 o 0.5 de $f^*$) para suavizar la varianza extrema y evitar el Drawdown excesivo.*

---

## 📊 Evaluación del Historial de Apuestas (`history.json`)

Para medir la salud de la estrategia de apuestas, se deben inspeccionar y calcular de forma periódica las siguientes métricas clave a partir del archivo de persistencia `db/history.json`:

1. **ROI (Retorno de Inversión)**:
   $$\text{ROI} = \frac{\text{Ganancia Total}}{\text{Banca Inicial}} \times 100$$
2. **Yield (Rendimiento por unidad apostada)**:
   $$\text{Yield} = \frac{\text{Beneficio Neto Total}}{\text{Suma de todo el dinero apostado}} \times 100$$
3. **Win Rate (Tasa de acierto)**:
   $$\text{Win Rate} = \frac{\text{Apuestas Ganadas}}{\text{Total de Apuestas Resueltas}} \times 100$$
4. **Max Drawdown (Máxima caída de la banca)**:
   La mayor caída porcentual desde un pico máximo acumulado de la banca hasta el punto más bajo antes de un nuevo pico.

---

## 🛠️ Procedimiento de Ajuste de Simulación

Si el **Max Drawdown** supera el **25%** o el **Yield** es negativo tras 100 simulaciones, sigue este procedimiento:
1. **Reducir Fracción de Kelly**: Ve a `config/config.go` (o a través del panel) y reduce el multiplicador de Kelly (ej. de `0.5` a `0.2`).
2. **Subir el Umbral del Valor Esperado (+EV)**: Configura el motor para que solo emita apuestas donde el valor esperado de la cuota respecto a la probabilidad estimada sea significativamente superior (ej. de `+1%` a `+3%`).
3. **Reiniciar Historial de Pruebas**: Para simulaciones limpias en desarrollo, borra el archivo temporal del historial y reinicia el bot:
   ```bash
   rm db/history.json
   go run main.go
   ```
