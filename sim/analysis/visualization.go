package analysis

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Dashboard serves visualization for simulation outputs.
type Dashboard struct {
	Results     map[string]MonteCarloResult
	Sensitivity *SensitivityResult
	Port        int
}

// NewDashboard creates a dashboard.
func NewDashboard(port int) *Dashboard {
	return &Dashboard{Port: port}
}

// Serve starts the dashboard server.
func (d *Dashboard) Serve() error {
	if d.Port == 0 {
		d.Port = 8080
	}

	http.HandleFunc("/", d.handleIndex)
	http.HandleFunc("/api/monte-carlo", d.handleResults)
	http.HandleFunc("/api/sensitivity", d.handleSensitivity)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", d.Port),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}

func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexTemplate))
	_ = tmpl.Execute(w, map[string]interface{}{
		"HasResults":     d.Results != nil,
		"HasSensitivity": d.Sensitivity != nil,
	})
}

func (d *Dashboard) handleResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(d.Results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *Dashboard) handleSensitivity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(d.Sensitivity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>VirtEngine Tokenomics Dashboard</title>
  <style>
    body { font-family: "Helvetica Neue", Arial, sans-serif; margin: 2rem; background: #f7f7f4; color: #222; }
    h1 { margin-bottom: 0.5rem; }
    .card { background: #fff; padding: 1rem 1.5rem; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); margin-bottom: 1rem; }
    table { width: 100%; border-collapse: collapse; font-size: 0.95rem; }
    th, td { padding: 0.5rem; border-bottom: 1px solid #eee; text-align: left; }
  </style>
</head>
<body>
  <h1>VirtEngine Tokenomics Dashboard</h1>
  <p>Monte Carlo and sensitivity outputs for tokenomics simulations.</p>
  {{if .HasResults}}
  <div class="card">
    <h2>Monte Carlo Metrics</h2>
    <table id="mc-table">
      <thead>
        <tr><th>Metric</th><th>Mean</th><th>Std Dev</th><th>Median</th><th>95% CI</th></tr>
      </thead>
      <tbody></tbody>
    </table>
  </div>
  {{end}}
  {{if .HasSensitivity}}
  <div class="card">
    <h2>Sensitivity Analysis</h2>
    <table id="sens-table">
      <thead>
        <tr><th>Param</th><th>Elasticity</th><th>Points</th></tr>
      </thead>
      <tbody></tbody>
    </table>
  </div>
  {{end}}

  <script>
    async function loadMonteCarlo() {
      const res = await fetch("/api/monte-carlo");
      const data = await res.json();
      const tbody = document.querySelector("#mc-table tbody");
      if (!tbody) return;
      Object.values(data).forEach(function(metric) {
        const row = document.createElement("tr");
        row.innerHTML = "<td>" + metric.Metric + "</td><td>" + metric.Mean.toFixed(3) + "</td><td>" + metric.StdDev.toFixed(3) + "</td><td>" + metric.Median.toFixed(3) + "</td><td>" + metric.ConfidenceLower.toFixed(3) + " - " + metric.ConfidenceUpper.toFixed(3) + "</td>";
        tbody.appendChild(row);
      });
    }
    async function loadSensitivity() {
      const res = await fetch("/api/sensitivity");
      const data = await res.json();
      const tbody = document.querySelector("#sens-table tbody");
      if (!tbody || !data || !data.Param) return;
      const row = document.createElement("tr");
      row.innerHTML = "<td>" + data.Param + "</td><td>" + data.Elastic.toFixed(3) + "</td><td>" + data.Points.length + "</td>";
      tbody.appendChild(row);
    }
    loadMonteCarlo();
    loadSensitivity();
  </script>
</body>
</html>`
