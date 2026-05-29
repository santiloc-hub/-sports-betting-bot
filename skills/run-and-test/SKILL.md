---
name: run-and-test-sports-bot
description: Compila y ejecuta el bot deportivo en Go, corre pruebas unitarias y valida la salud del servidor web local y sus endpoints SSE.
---

# Skill: Ejecución y Validación del Sports Betting Bot

Esta skill proporciona las directrices y flujos de trabajo para compilar, probar y verificar el funcionamiento correcto del bot deportivo.

## 📋 Requisitos Previos
- Tener **Go (Golang)** instalado y disponible en el PATH del sistema.
- Estar en el directorio raíz del repositorio: `/mnt/c/Users/SANTI/.gemini/antigravity/scratch/sports-betting-bot`

---

## 🛠️ Procedimiento de Compilación y Ejecución

### 1. Validación de Sintaxis y Compilación
Antes de iniciar el bot, asegúrate de que no haya errores de compilación ejecutando:
```bash
go build -o bin/sports-betting-bot main.go
```

### 2. Ejecución Local del Bot
Para iniciar el motor del bot y el servidor web integrado de forma interactiva:
```bash
go run main.go
```
*El servidor web debería levantarse por defecto en el puerto `8080` (o el puerto especificado en `config/config.go`).*

### 3. Verificación de Salud de Endpoints
Una vez iniciado, valida que los servicios respondan correctamente:
- **Dashboard Web**: Realizar un ping/GET simple a `http://localhost:8080/` para confirmar que el HTML responde con código `200 OK`.
- **Canal de Server-Sent Events (SSE)**: Validar la conexión a `http://localhost:8080/events` (canal de flujo de datos en tiempo real) y asegurar que transmite la estructura JSON de logs y estadísticas sin bloqueos.

---

## 🧪 Suite de Pruebas Unitarias

Para asegurar que los cálculos matemáticos (Kelly Criterion, Arbitraje, +EV) y la base de datos local funcionen de manera óptima, ejecuta:

```bash
go test -v ./...
```

### Áreas Críticas de Prueba:
- **`strategy/math_test.go`**: Validar que la fracción de Kelly nunca devuelva valores negativos bajo cuotas desfavorables y que los cálculos de Valor Esperado (+EV) sean correctos.
- **`db/database_test.go`**: Validar la concurrencia segura de lectura y escritura (`RWMutex`) en `history.json` para prevenir condiciones de carrera (*data races*).
