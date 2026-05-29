# ⚡ ANTIGRAVITY SPORTS BOT v1.0

Un bot de deportes de ultra-alta velocidad escrito en **Go (Golang)**, diseñado para escaneo de cuotas de baja latencia, toma de decisiones matemáticas asíncronas y monitoreo en tiempo real desde una interfaz web premium (*glassmorphic dark-mode*) optimizada para móviles.

---

## 🚀 Características Clave

1.  **Motor Asíncrono de Alta Velocidad (Go):** Compilación nativa y procesamiento en microsegundos usando *Goroutines*.
2.  **Persistencia en Base de Datos Ligera (`history.json`):** Almacena y audita cada cuota analizada y apuesta realizada en modo simulación.
3.  **Análisis Histórico Integrado:** Calcula en tiempo real métricas financieras de rendimiento como **ROI**, **Yield** y **Win Rate**, y grafica la curva de crecimiento del capital (Bankroll) mediante Chart.js.
4.  **Flujo en Vivo mediante Server-Sent Events (SSE):** Conexión nativa unidireccional de bajísima latencia para empujar logs de la terminal directamente al navegador de tu celular con retraso cero.
5.  **Simulación Acelerada Inteligente:** Para facilitar las pruebas, si no se cuenta con claves de API, el bot simula cuotas fluctuantes de Champions League y NBA, y resuelve las apuestas pendientes en 15 segundos con base en probabilidades matemáticas reales.

---

## 🛠️ Requisitos de Instalación (2 Minutos)

Para ejecutar el bot en tu computadora, solo necesitas instalar Go:

1.  Descarga el instalador oficial de Go para Windows desde: [https://go.dev/dl/](https://go.dev/dl/)
2.  Ejecuta el archivo `.msi` descargado y sigue las instrucciones del instalador (haz clic en *Next*).
3.  Abre una terminal (PowerShell o CMD) y escribe para verificar la instalación:
    ```bash
    go version
    ```

---

## 🏃 Cómo Arrancar Localmente

Una vez instalado Go:

1.  Abre tu terminal en la carpeta del proyecto (`C:\Users\SANTI\.gemini\antigravity\scratch\sports-betting-bot`).
2.  Ejecuta el bot con el siguiente comando:
    ```bash
    go run main.go
    ```
3.  Abre tu navegador (en la PC o en el celular si están en la misma red) en:
    👉 **[http://localhost:8080](http://localhost:8080)**

---

## 📱 Cómo Monitorearlo y Controlarlo desde tu Celular

Para ver el comportamiento del bot en tu teléfono de forma rápida y gratuita, tienes dos excelentes opciones:

### Opción A: Túnel Instantáneo y Gratuito (Recomendado)
Puedes crear una URL pública segura `https://...` en 10 segundos apuntando a tu computadora local usando **Ngrok** o **Cloudflare Tunnels**:

*   **Usando Cloudflare (Totalmente gratis, sin registros):**
    1. Descarga el ejecutable rápido de Cloudflare para tu OS.
    2. Ejecuta en tu consola:
       ```bash
       cloudflared tunnel --url http://localhost:8080
       ```
    3. Copia la dirección web segura (`https://tucodigo.trycloudflare.com`) generada en la terminal y ábrela en tu celular. ¡Verás el panel actualizándose en vivo!

*   **Usando Ngrok:**
    1. Instala ngrok y corre en tu consola:
       ```bash
       ngrok http 8080
       ```
    2. Abre la URL HTTPS generada en tu celular.

### Opción B: Alojamiento Permanente Gratis en la Nube (Render.com)
Si deseas hostearlo en internet las 24/7 de forma permanente y gratis:

1.  Crea un repositorio en tu cuenta de GitHub (ej. `sports-betting-bot`) y sube este código.
2.  Regístrate de forma gratuita en **[Render.com](https://render.com/)**.
3.  Crea un nuevo **Web Service** y conecta tu repositorio de GitHub.
4.  Configura los siguientes parámetros en Render:
    *   **Runtime:** `Go`
    *   **Build Command:** `go build -o bin/main main.go`
    *   **Start Command:** `./bin/main`
5.  Haz clic en *Deploy*. ¡Render compilará el código en Go y te entregará una URL permanente gratis para abrirla cuando quieras desde tu móvil!

---

## ⚙️ Variables de Configuración (Entorno)

Puedes modificar el comportamiento del bot configurando variables de entorno en tu máquina o en tu panel de Render:

*   `PORT`: Puerto del servidor web (por defecto `8080`).
*   `INITIAL_BANKROLL`: Capital inicial en USD para la simulación (por defecto `1000.0`).
*   `THE_ODDS_API_KEY`: Tu clave de API de *The Odds API* para traer cuotas reales del mercado mundial. *Si la dejas vacía, el bot correrá en modo simulación autónomo con cuotas simuladas de alta fidelidad*.
*   `KELLY_FRACTION`: Fracción del criterio de Kelly a apostar (por defecto `0.25` para ¼ de Kelly, ideal para control de riesgo).
*   `SIMULATION_MODE`: `true` para operar solo con dinero ficticio (por defecto `true`).
