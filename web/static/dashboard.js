// ARCHIVO JS PRINCIPAL DEL PANEL DE CONTROL EN TIEMPO REAL

let activeTab = 'live';
let sseSource = null;
let bankrollChart = null;

// Inicialización del Dashboard
document.addEventListener("DOMContentLoaded", () => {
    initSSE();
    initChart();
    loadRESTData();
    loadConfigData();
});

// Pestañas (Tab Switcher)
function switchTab(tab) {
    activeTab = tab;
    
    // Toggle botones activos
    document.getElementById("btn-live-tab").classList.toggle("active", tab === 'live');
    document.getElementById("btn-history-tab").classList.toggle("active", tab === 'history');
    document.getElementById("btn-settings-tab").classList.toggle("active", tab === 'settings');
    
    // Toggle contenido de pestañas
    document.getElementById("tab-live").classList.toggle("active", tab === 'live');
    document.getElementById("tab-history").classList.toggle("active", tab === 'history');
    document.getElementById("tab-settings").classList.toggle("active", tab === 'settings');

    // Forzar redibujo del gráfico al entrar a la pestaña de análisis histórico
    if (tab === 'history' && bankrollChart) {
        setTimeout(() => {
            bankrollChart.resize();
            bankrollChart.update();
        }, 100);
    }
}

// Conectar con el Servidor mediante Server-Sent Events (SSE) nativo
function initSSE() {
    const consoleStream = document.getElementById("console-stream");
    const cursorLine = consoleStream.querySelector(".cursor");

    // Conectar a la API SSE expuesta por Go
    sseSource = new EventSource("/events");

    // Evento 1: Recibir logs de actividad
    sseSource.addEventListener("log", (event) => {
        // Crear elemento de línea de log
        const logLine = document.createElement("div");
        logLine.className = "console-line";
        
        // Formatear colores especiales si hay palabras clave
        let text = event.data;
        if (text.includes("¡ARBITRAJE DETECTADO!")) {
            logLine.classList.add("positive");
            logLine.style.fontWeight = "700";
        } else if (text.includes("+EV Detectado")) {
            logLine.style.color = "hsl(205, 95%, 65%)";
        } else if (text.includes("GANADA!")) {
            logLine.classList.add("positive");
        } else if (text.includes("PERDIDA!")) {
            logLine.classList.add("negative");
        }

        logLine.textContent = text;
        
        // Insertar antes del cursor e inclinar scroll al final
        consoleStream.insertBefore(logLine, cursorLine);
        consoleStream.scrollTop = consoleStream.scrollHeight;
    });

    // Evento 2: Recibir estadísticas dinámicas de la banca
    sseSource.addEventListener("stats", (event) => {
        const stats = JSON.parse(event.data);
        updateKPIs(stats);
    });

    // Manejar desconexión
    sseSource.onerror = (err) => {
        console.error("Fallo de conexión SSE. Reintentando...", err);
        const errLine = document.createElement("div");
        errLine.className = "console-line text-accent";
        errLine.textContent = `[SISTEMA] Reestableciendo flujo asíncrono con el motor de Go...`;
        consoleStream.insertBefore(errLine, cursorLine);
        consoleStream.scrollTop = consoleStream.scrollHeight;
    };
}

// Actualizar los KPIs principales en pantalla
function updateKPIs(stats) {
    // 1. Capital Total
    const valBankroll = document.getElementById("val-bankroll");
    valBankroll.textContent = formatUSD(stats.currentBankroll);

    // 2. Ganancia Neta
    const valProfit = document.getElementById("val-profit");
    valProfit.textContent = (stats.totalProfit >= 0 ? "+" : "") + formatUSD(stats.totalProfit);
    valProfit.className = "kpi-value " + (stats.totalProfit >= 0 ? "positive" : "negative");

    // 3. Tasa de Aciertos
    const valWinrate = document.getElementById("val-winrate");
    valWinrate.textContent = stats.winRate.toFixed(1) + "%";
    
    const subWinrate = document.getElementById("sub-winrate");
    const resolved = stats.wonBets + stats.lostBets;
    subWinrate.textContent = `${stats.wonBets} ganadas / ${resolved} resueltas`;

    // 4. ROI
    const valRoi = document.getElementById("val-roi");
    valRoi.textContent = (stats.ROI >= 0 ? "+" : "") + stats.ROI.toFixed(2) + "%";
    valRoi.className = "kpi-value " + (stats.ROI >= 0 ? "positive" : "negative");

    // 5. Yield Promedio (Pestaña Histórico)
    const valYield = document.getElementById("val-yield");
    valYield.textContent = (stats.Yield >= 0 ? "+" : "") + stats.Yield.toFixed(2) + "%";
    valYield.className = "stats-subvalue " + (stats.Yield >= 0 ? "positive" : "negative");

    // 6. Drawdown Histórico (Pestaña Histórico)
    const valDrawdown = document.getElementById("val-drawdown");
    valDrawdown.textContent = stats.maxDrawdown.toFixed(2) + "%";

    // 7. Apuestas Totales y Resueltas
    document.getElementById("lbl-total-bets").textContent = `Total: ${stats.totalBets}`;
    document.getElementById("val-resolved-count").textContent = resolved;

    // Recargar tabla de apuestas y actualizar gráfico
    loadRESTData();
}

// Carga inicial y recarga REST de la lista de apuestas, gráfico y decisiones escaneadas
function loadRESTData() {
    // Cargar historial de apuestas para la tabla
    fetch("/api/history")
        .then(res => res.json())
        .then(data => {
            renderBetsTable(data.bets);
            updateChartData(data.bets);
        })
        .catch(err => console.error("Error al obtener historial REST:", err));

    // Cargar partidos escaneados hoy y decisiones
    fetch("/api/scanned")
        .then(res => res.json())
        .then(data => {
            renderScannedTable(data);
        })
        .catch(err => console.error("Error al obtener escaneados REST:", err));
}

// Renderizar dinámicamente la tabla de apuestas recientes
function renderBetsTable(bets) {
    const tbody = document.getElementById("bets-tbody");
    
    if (!bets || bets.length === 0) {
        tbody.innerHTML = `
            <tr id="row-empty">
                <td colspan="6" class="text-center text-muted">Aún no hay apuestas registradas. Esperando cuotas...</td>
            </tr>`;
        return;
    }

    // Limpiar tabla
    tbody.innerHTML = "";

    // Invertir lista para mostrar las más recientes arriba
    const sortedBets = [...bets].reverse();

    sortedBets.forEach(b => {
        const tr = document.createElement("tr");
        
        // Formatear Fecha
        const date = new Date(b.timestamp);
        const timeStr = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

        // Estado Badge
        let statusBadge = "";
        let profitCell = "";

        if (b.status === "PENDING") {
            statusBadge = `<span class="badge badge-pending">PENDIENTE</span>`;
            profitCell = `<span class="text-muted">-$${b.stake.toFixed(2)}</span>`;
        } else if (b.status === "WON") {
            statusBadge = `<span class="badge badge-won">GANADA</span>`;
            profitCell = `<span class="positive">+$${b.netProfit.toFixed(2)}</span>`;
        } else {
            statusBadge = `<span class="badge badge-lost">PERDIDA</span>`;
            profitCell = `<span class="negative">-$${b.stake.toFixed(2)}</span>`;
        }

        tr.innerHTML = `
            <td class="text-muted">${timeStr}</td>
            <td><strong>${b.eventName}</strong></td>
            <td><span class="text-muted" style="font-size: 11px;">${b.sport}</span></td>
            <td>${b.outcome} @ <strong>${b.odds.toFixed(2)}</strong> <br><small class="text-muted">${b.bookmaker}</small></td>
            <td>$${b.stake.toFixed(2)}</td>
            <td>${statusBadge}</td>
        `;
        tbody.appendChild(tr);
    });
}

// Inicializar el gráfico de balance de Chart.js con degradado neón
function initChart() {
    const ctx = document.getElementById("bankrollChart").getContext("2d");
    
    // Crear degradado para el área bajo la curva
    const gradient = ctx.createLinearGradient(0, 0, 0, 300);
    gradient.addColorStop(0, 'rgba(27, 137, 240, 0.3)');
    gradient.addColorStop(1, 'rgba(27, 137, 240, 0.0)');

    bankrollChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: ["Inicio"],
            datasets: [{
                label: 'Balance ($ USD)',
                data: [1000.00],
                borderColor: 'hsl(205, 95%, 55%)',
                borderWidth: 3,
                backgroundColor: gradient,
                fill: true,
                tension: 0.3,
                pointBackgroundColor: 'hsl(205, 95%, 55%)',
                pointBorderColor: '#ffffff',
                pointRadius: 4,
                pointHoverRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false }
            },
            scales: {
                x: {
                    grid: { color: 'rgba(255, 255, 255, 0.03)' },
                    ticks: { color: 'hsl(210, 10%, 60%)', font: { family: 'Inter' } }
                },
                y: {
                    grid: { color: 'rgba(255, 255, 255, 0.03)' },
                    ticks: { 
                        color: 'hsl(210, 10%, 60%)', 
                        font: { family: 'Inter' },
                        callback: function(value) { return '$' + value.toFixed(2); }
                    }
                }
            }
        }
    });
}

// Actualizar los datos del gráfico con el historial
function updateChartData(bets) {
    if (!bankrollChart || !bets || bets.length === 0) return;

    let current = 1000.00;
    const labels = ["Inicio"];
    const dataPoints = [current];

    bets.forEach((b, index) => {
        labels.push(`#${index + 1}`);
        dataPoints.push(b.bankrollAfter);
    });

    bankrollChart.data.labels = labels;
    bankrollChart.data.datasets[0].data = dataPoints;
    bankrollChart.update();
}

function formatUSD(value) {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
}

// Cargar la configuración actual desde el servidor
function loadConfigData() {
    fetch("/api/config")
        .then(res => res.json())
        .then(data => {
            if (data.apiKey) document.getElementById("txt-api-key").value = data.apiKey;
            if (data.initialBankroll !== undefined) document.getElementById("txt-bankroll").value = data.initialBankroll;
            if (data.kellyFraction !== undefined) document.getElementById("txt-kelly-fraction").value = data.kellyFraction;
            if (data.sportsToScan) document.getElementById("txt-sports").value = data.sportsToScan;
            if (data.minEVThreshold !== undefined) document.getElementById("txt-min-ev").value = data.minEVThreshold;
            if (data.sharpBookmaker) document.getElementById("txt-sharp-bm").value = data.sharpBookmaker;
            if (data.minOdds !== undefined) document.getElementById("txt-min-odds").value = data.minOdds;
            if (data.maxOdds !== undefined) document.getElementById("txt-max-odds").value = data.maxOdds;
            if (data.simulationMode !== undefined) document.getElementById("chk-sim-mode").checked = data.simulationMode;
            
            // Lanzar carga inicial de optimización de modelo
            triggerRetrain();
        })
        .catch(err => console.error("Error al cargar configuración:", err));
}

// Guardar la configuración en caliente
function saveConfig(event) {
    event.preventDefault();
    const apiKey = document.getElementById("txt-api-key").value.trim();
    const initialBankroll = parseFloat(document.getElementById("txt-bankroll").value) || 1000.0;
    const kellyFraction = parseFloat(document.getElementById("txt-kelly-fraction").value) || 0.25;
    const sportsToScan = document.getElementById("txt-sports").value.trim() || "upcoming";
    const minEVThreshold = parseFloat(document.getElementById("txt-min-ev").value) || 1.0;
    const sharpBookmaker = document.getElementById("txt-sharp-bm").value.trim() || "Pinnacle";
    const minOdds = parseFloat(document.getElementById("txt-min-odds").value) || 1.05;
    const maxOdds = parseFloat(document.getElementById("txt-max-odds").value) || 10.0;
    const simulationMode = document.getElementById("chk-sim-mode").checked;

    const alertBox = document.getElementById("cfg-alert");

    const payload = {
        apiKey,
        initialBankroll,
        kellyFraction,
        sportsToScan,
        minEVThreshold,
        sharpBookmaker,
        minOdds,
        maxOdds,
        simulationMode
    };

    fetch("/api/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === "success") {
            alertBox.style.display = "block";
            alertBox.style.background = "rgba(40, 167, 69, 0.2)";
            alertBox.style.color = "#28a745";
            alertBox.style.border = "1px solid rgba(40, 167, 69, 0.4)";
            alertBox.textContent = "¡Configuración y estrategia guardadas y actualizadas en caliente con éxito!";
            
            setTimeout(() => {
                alertBox.style.display = "none";
            }, 4000);
        } else {
            throw new Error(data.error || "Fallo al guardar");
        }
    })
    .catch(err => {
        alertBox.style.display = "block";
        alertBox.style.background = "rgba(220, 53, 69, 0.2)";
        alertBox.style.color = "#dc3545";
        alertBox.style.border = "1px solid rgba(220, 53, 69, 0.4)";
        alertBox.textContent = "Error al guardar la configuración: " + err.message;
    });
}

// Solicitar re-entrenamiento del modelo (Grid Search)
function triggerRetrain() {
    const tbody = document.getElementById("model-rules-tbody");
    tbody.innerHTML = `<tr><td colspan="4" class="text-center text-accent" style="padding: 15px 0;">Ejecutando Grid Search retrospectivo...</td></tr>`;

    fetch("/api/retrain", {
        method: "POST"
    })
    .then(res => res.json())
    .then(data => {
        // Actualizar KPIs de optimización
        document.getElementById("lbl-best-rule").textContent = data.bestPerformingRule;
        document.getElementById("lbl-best-ev").textContent = data.optimalEVThreshold.toFixed(1) + "%";
        document.getElementById("lbl-best-kelly").textContent = data.optimalKellyFraction.toFixed(2);
        
        const roiImprovement = document.getElementById("lbl-roi-improvement");
        roiImprovement.textContent = "+" + data.estimatedROIImprovement.toFixed(2) + "%";
        
        // Renderizar tabla de reglas
        if (!data.rulesEvaluated || data.rulesEvaluated.length === 0) {
            tbody.innerHTML = `<tr><td colspan="4" class="text-center text-muted" style="padding: 15px 0;">El re-entrenamiento requiere al menos 5 apuestas en el historial.</td></tr>`;
            return;
        }

        tbody.innerHTML = "";
        data.rulesEvaluated.forEach(r => {
            const tr = document.createElement("tr");
            tr.style.borderBottom = "1px solid rgba(255,255,255,0.05)";
            
            const profitClass = r.profit >= 0 ? "positive" : "negative";
            const profitPrefix = r.profit >= 0 ? "+" : "";

            tr.innerHTML = `
                <td style="padding: 8px 0;"><strong>${r.ruleName}</strong></td>
                <td style="padding: 8px 0; text-align: center;">${r.totalBets}</td>
                <td style="padding: 8px 0; text-align: center;">${r.accuracy.toFixed(1)}%</td>
                <td style="padding: 8px 0; text-align: right;" class="${profitClass}"><strong>${profitPrefix}$${r.profit.toFixed(2)}</strong></td>
            `;
            tbody.appendChild(tr);
        });
    })
    .catch(err => {
        console.error("Error al re-entrenar modelo:", err);
        tbody.innerHTML = `<tr><td colspan="4" class="text-center text-danger" style="padding: 15px 0;">Error al ejecutar el optimizador.</td></tr>`;
    });
}

// Renderizar la tabla de decisiones en vivo de partidos escaneados
function renderScannedTable(scanned) {
    const tbody = document.getElementById("scanned-tbody");
    if (!scanned || scanned.length === 0) {
        tbody.innerHTML = `
            <tr id="scanned-row-empty">
                <td colspan="8" class="text-center text-muted" style="padding: 20px 0;">Aún no se han analizado partidos en esta ronda. Esperando escaneo...</td>
            </tr>`;
        return;
    }

    tbody.innerHTML = "";
    scanned.forEach(s => {
        const tr = document.createElement("tr");

        const date = new Date(s.timestamp);
        const timeStr = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

        // Formatear decisión
        let decisionBadge = "";
        if (s.decision.includes("APOSTADO")) {
            decisionBadge = `<span class="badge badge-won" style="background: rgba(40,167,69,0.2); color: #28a745; border: 1px solid rgba(40,167,69,0.3); font-weight: bold; font-size: 11px;">${s.decision}</span>`;
        } else {
            decisionBadge = `<span class="badge badge-lost" style="background: rgba(220,53,69,0.15); color: #dc3545; border: 1px solid rgba(220,53,69,0.25); font-size: 11px;">DESCARTADO</span>`;
        }

        // Resaltar EV positivo
        const evHomeClass = s.evHome > 1.0 ? "positive font-bold" : "text-muted";
        const evAwayClass = s.evAway > 1.0 ? "positive font-bold" : "text-muted";

        tr.innerHTML = `
            <td class="text-muted" style="font-size: 12px;">${timeStr}</td>
            <td><strong>${s.eventName}</strong></td>
            <td><span class="text-muted" style="font-size: 11px;">${s.sport}</span></td>
            <td style="text-align: center;">${s.trueHomeProb.toFixed(1)}% / ${s.trueAwayProb.toFixed(1)}%</td>
            <td style="text-align: center;">${s.bestOddsHome.toFixed(2)} / ${s.bestOddsAway.toFixed(2)}</td>
            <td style="text-align: center;"><span class="${evHomeClass}">${s.evHome.toFixed(2)}%</span> / <span class="${evAwayClass}">${s.evAway.toFixed(2)}%</span></td>
            <td><span class="badge badge-pending" style="font-size: 11px; font-weight: 600;">${s.ruleUsed}</span></td>
            <td>${decisionBadge}</td>
        `;
        tbody.appendChild(tr);
    });
}

// Borrar historial y resetear banca
function resetHistory() {
    if (!confirm("¿Estás seguro de que deseas borrar todo el historial de apuestas y reiniciar la banca al valor configurado? Esta acción no se puede deshacer y limpiará la base de datos local.")) {
        return;
    }

    fetch("/api/reset", {
        method: "POST"
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === "success") {
            alert("¡Historial borrado con éxito! La banca se ha reiniciado.");
            loadRESTData(); // Recargar datos de la tabla y gráfico
        } else {
            alert("Error al borrar historial: " + data.error);
        }
    })
    .catch(err => {
        alert("Error de conexión: " + err.message);
    });
}
