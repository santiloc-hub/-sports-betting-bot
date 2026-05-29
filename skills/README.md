# 📂 Skills del Sports Betting Bot

Este directorio contiene las **Skills** personalizadas diseñadas específicamente para el desarrollo, pruebas, optimización y monitoreo del **Bot de Apuestas Deportivas en Go**.

## 🛠️ ¿Qué es una Skill?
Una **Skill** es un conjunto de instrucciones estructuradas, scripts y recursos que extienden mis capacidades (las del asistente de IA) para realizar tareas complejas y especializadas dentro de este repositorio de manera consistente y eficiente.

Cada carpeta de skill en este directorio está estructurada de la siguiente manera:
- **`SKILL.md`**: El archivo de instrucciones principal que contiene el frontmatter de YAML (nombre, descripción) y las directrices detalladas para mí.
- **`scripts/`** *(Opcional)*: Scripts auxiliares en bash, Go o Python para automatizar tareas.
- **`examples/`** *(Opcional)*: Ejemplos de uso o configuraciones de referencia.

---

## 🚀 Skills Disponibles en este Repositorio

Hemos inicializado las siguientes skills para ayudarte a operar y mejorar el bot de manera automática:

### 1. 🏃‍♂️ [run-and-test](file:///mnt/c/Users/SANTI/.gemini/antigravity/scratch/sports-betting-bot/skills/run-and-test/SKILL.md)
* **Propósito**: Compilar y ejecutar el bot en Go, verificar el estado del servidor web (puerto `8080`), comprobar la transmisión en vivo por Server-Sent Events (SSE) y ejecutar la suite de pruebas unitarias.
* **Uso**: Ideal cuando realizamos cambios en el servidor HTTP o en el motor de ejecución en segundo plano y necesitamos verificar que todo compile y funcione de forma segura.

### 2. 📊 [simulate-bets](file:///mnt/c/Users/SANTI/.gemini/antigravity/scratch/sports-betting-bot/skills/simulate-bets/SKILL.md)
* **Propósito**: Ejecutar simulaciones matemáticas avanzadas utilizando las fórmulas del Criterio de Kelly y Valor Esperado (+EV) para ajustar el balance riesgo/retorno.
* **Uso**: Ideal para afinar la fracción de Kelly, analizar el historial de apuestas almacenado en `db/history.json` y evaluar el Yield y el Drawdown acumulado.

### 3. 🧠 [explain-bot](file:///mnt/c/Users/SANTI/.gemini/antigravity/scratch/sports-betting-bot/skills/explain-bot/SKILL.md)
* **Propósito**: Explicación conceptual (para dummies) y rigurosa (para científicos de datos) sobre el flujo de datos del bot, matemáticas y arquitectura.
* **Uso**: Sirve como documentación de referencia viva para mí o cualquier otro agente que trabaje en este código, y la mantendremos actualizada ante cualquier cambio en las reglas de negocio del bot.

### 4. 📅 [daily-analysis](file:///mnt/c/Users/SANTI/.gemini/antigravity/scratch/sports-betting-bot/skills/daily-analysis/SKILL.md)
* **Propósito**: Escanear partidos programados del día de tenis y fútbol, cruzar datos de APIs estadísticas e históricas para estimar probabilidades e identificar apuestas +EV.
* **Uso**: Útil para realizar la fase de investigación previa a la jornada deportiva y emitir portafolios de apuestas optimizados.

---

## ✍️ Cómo Crear una Nueva Skill
Si deseas que desarrolle una nueva habilidad especializada (por ejemplo, "conectar una API de Telegram para notificaciones"), podemos crear una nueva carpeta aquí con su respectivo `SKILL.md`. La estructura básica de un `SKILL.md` es:

```markdown
---
name: nombre-de-la-skill
description: Breve descripción de lo que hace esta skill.
---

# Instrucciones de la Skill
Detalla aquí el paso a paso, los comandos que debo ejecutar y qué resultados se esperan obtener.
```
