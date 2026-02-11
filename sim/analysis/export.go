package analysis

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

// WriteJSON writes a JSON payload to disk.
func WriteJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// WriteMonteCarloCSV exports Monte Carlo metrics into CSV.
func WriteMonteCarloCSV(path string, results map[string]MonteCarloResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"metric", "mean", "std_dev", "median", "p5", "p95", "ci_lower", "ci_upper"}); err != nil {
		return err
	}
	for _, result := range results {
		row := []string{
			result.Metric,
			formatFloat(result.Mean),
			formatFloat(result.StdDev),
			formatFloat(result.Median),
			formatFloat(result.Percentile5),
			formatFloat(result.Percentile95),
			formatFloat(result.ConfidenceLower),
			formatFloat(result.ConfidenceUpper),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// WriteDashboardHTML renders a static dashboard HTML file with embedded data.
func WriteDashboardHTML(path string, results map[string]MonteCarloResult, sensitivity *SensitivityResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := struct {
		Results     map[string]MonteCarloResult
		Sensitivity *SensitivityResult
		GeneratedAt time.Time
	}{
		Results:     results,
		Sensitivity: sensitivity,
		GeneratedAt: time.Now().UTC(),
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("dashboard").Funcs(template.FuncMap{
		"toJSON": func(v interface{}) template.JS {
			if v == nil {
				return template.JS("null")
			}
			bz, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(bz)
		},
	}).Parse(exportTemplate))
	return tmpl.Execute(file, payload)
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.6f", value)
}

const exportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>VirtEngine Economics Dashboard</title>
  <style>
    body { font-family: "Helvetica Neue", Arial, sans-serif; margin: 2rem; background: #f3f6f4; color: #222; }
    h1 { margin-bottom: 0.5rem; }
    .meta { color: #666; margin-bottom: 1.5rem; }
    .card { background: #fff; padding: 1rem 1.5rem; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); margin-bottom: 1rem; }
    table { width: 100%; border-collapse: collapse; font-size: 0.95rem; }
    th, td { padding: 0.5rem; border-bottom: 1px solid #eee; text-align: left; }
  </style>
</head>
<body>
  <h1>VirtEngine Economics Dashboard</h1>
  <div class="meta">Generated at {{ .GeneratedAt }}</div>

  <div class="card" id="mc-card">
    <h2>Monte Carlo Metrics</h2>
    <table id="mc-table">
      <thead>
        <tr><th>Metric</th><th>Mean</th><th>Std Dev</th><th>Median</th><th>5-95%</th></tr>
      </thead>
      <tbody></tbody>
    </table>
  </div>

  <div class="card" id="sens-card">
    <h2>Sensitivity Analysis</h2>
    <table id="sens-table">
      <thead>
        <tr><th>Param</th><th>Elasticity</th><th>Points</th></tr>
      </thead>
      <tbody></tbody>
    </table>
  </div>

  <script>
    const results = {{ .Results | toJSON }};
    const sensitivity = {{ .Sensitivity | toJSON }};

    function renderMonteCarlo() {
      const tbody = document.querySelector("#mc-table tbody");
      if (!tbody || !results) return;
      Object.values(results).forEach(function(metric) {
        const row = document.createElement("tr");
        row.innerHTML = "<td>" + metric.Metric + "</td><td>" + metric.Mean.toFixed(3) + "</td><td>" + metric.StdDev.toFixed(3) + "</td><td>" + metric.Median.toFixed(3) + "</td><td>" + metric.Percentile5.toFixed(3) + " - " + metric.Percentile95.toFixed(3) + "</td>";
        tbody.appendChild(row);
      });
    }

    function renderSensitivity() {
      const tbody = document.querySelector("#sens-table tbody");
      if (!tbody || !sensitivity || !sensitivity.Param) return;
      const row = document.createElement("tr");
      row.innerHTML = "<td>" + sensitivity.Param + "</td><td>" + sensitivity.Elastic.toFixed(3) + "</td><td>" + sensitivity.Points.length + "</td>";
      tbody.appendChild(row);
    }

    renderMonteCarlo();
    renderSensitivity();
  </script>
</body>
</html>`
