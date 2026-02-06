package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/sim/analysis"
	"github.com/virtengine/virtengine/sim/core"
	"github.com/virtengine/virtengine/sim/scenarios"
)

func main() {
	root := &cobra.Command{
		Use:   "ve-sim",
		Short: "VirtEngine economic simulation",
	}

	root.AddCommand(runCmd())
	root.AddCommand(monteCarloCmd())
	root.AddCommand(sensitivityCmd())
	root.AddCommand(dashboardCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	var scenario string
	var output string
	var configPath string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a simulation scenario",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig(scenario, configPath)
			if err != nil {
				return err
			}
			engine := core.NewEngine(cfg)
			if err := engine.Initialize(context.Background()); err != nil {
				return err
			}
			result, err := engine.Run(context.Background())
			if err != nil {
				return err
			}

			return writeJSON(output, result)
		},
	}

	cmd.Flags().StringVar(&scenario, "scenario", "baseline", "Scenario name")
	cmd.Flags().StringVar(&output, "output", "simulation.json", "Output file")
	cmd.Flags().StringVar(&configPath, "config", "", "Path to JSON config (overrides scenario)")
	return cmd
}

func monteCarloCmd() *cobra.Command {
	var scenario string
	var output string
	var runs int
	var configPath string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run Monte Carlo analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig(scenario, configPath)
			if err != nil {
				return err
			}
			mc := analysis.NewMonteCarloAnalyzer(analysis.MonteCarloConfig{
				Runs:         runs,
				Parallelism:  4,
				Confidence:   0.95,
				ParamRanges:  defaultParamRanges(),
				BaseConfig:   cfg,
				ScenarioName: scenario,
			})
			results, err := mc.Run(context.Background())
			if err != nil {
				return err
			}
			return writeJSON(output, results)
		},
	}

	cmd.Flags().StringVar(&scenario, "scenario", "baseline", "Scenario name")
	cmd.Flags().StringVar(&output, "output", "monte-carlo.json", "Output file")
	cmd.Flags().IntVar(&runs, "runs", 100, "Number of Monte Carlo runs")
	cmd.Flags().StringVar(&configPath, "config", "", "Path to JSON config (overrides scenario)")
	return cmd
}

func sensitivityCmd() *cobra.Command {
	var scenario string
	var output string
	var param string
	var min float64
	var max float64
	var steps int
	var configPath string

	cmd := &cobra.Command{
		Use:   "sensitivity",
		Short: "Run parameter sensitivity analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig(scenario, configPath)
			if err != nil {
				return err
			}
			result, err := analysis.RunSensitivity(context.Background(), analysis.SensitivityConfig{
				BaseConfig: cfg,
				Steps:      steps,
				Param:      param,
				Min:        min,
				Max:        max,
			})
			if err != nil {
				return err
			}
			return writeJSON(output, result)
		},
	}

	cmd.Flags().StringVar(&scenario, "scenario", "baseline", "Scenario name")
	cmd.Flags().StringVar(&param, "param", "inflation_target_bps", "Parameter to sweep")
	cmd.Flags().Float64Var(&min, "min", 400, "Minimum parameter value")
	cmd.Flags().Float64Var(&max, "max", 1200, "Maximum parameter value")
	cmd.Flags().IntVar(&steps, "steps", 6, "Number of sweep steps")
	cmd.Flags().StringVar(&output, "output", "sensitivity.json", "Output file")
	cmd.Flags().StringVar(&configPath, "config", "", "Path to JSON config (overrides scenario)")
	return cmd
}

func dashboardCmd() *cobra.Command {
	var port int
	var monteCarloPath string
	var sensitivityPath string

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start visualization dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			dash := analysis.NewDashboard(port)
			if monteCarloPath != "" {
				var results map[string]analysis.MonteCarloResult
				if err := readJSON(monteCarloPath, &results); err != nil {
					return err
				}
				dash.Results = results
			}
			if sensitivityPath != "" {
				var sensitivity analysis.SensitivityResult
				if err := readJSON(sensitivityPath, &sensitivity); err != nil {
					return err
				}
				dash.Sensitivity = &sensitivity
			}
			return dash.Serve()
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "Port to serve dashboard")
	cmd.Flags().StringVar(&monteCarloPath, "monte-carlo", "", "Path to Monte Carlo results JSON")
	cmd.Flags().StringVar(&sensitivityPath, "sensitivity", "", "Path to sensitivity results JSON")
	return cmd
}

func resolveConfig(scenario, path string) (core.Config, error) {
	if path != "" {
		return loadConfig(path)
	}
	switch scenario {
	case "baseline":
		return scenarios.BaselineConfig(), nil
	case "bull":
		return scenarios.BullMarketConfig(), nil
	case "bear":
		return scenarios.BearMarketConfig(), nil
	case "attack":
		return scenarios.AttackConfig(), nil
	case "black_swan":
		return scenarios.BlackSwanConfig(), nil
	default:
		return core.Config{}, errors.New("unknown scenario")
	}
}

func loadConfig(path string) (core.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return core.Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg core.Config
	if err := decoder.Decode(&cfg); err != nil {
		return core.Config{}, err
	}
	if cfg.TimeStep == 0 {
		cfg.TimeStep = 24 * time.Hour
	}
	if cfg.EndTime.IsZero() {
		cfg.EndTime = cfg.StartTime.Add(365 * 24 * time.Hour)
	}
	return cfg, nil
}

func writeJSON(path string, v interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func readJSON(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(v)
}

func defaultParamRanges() map[string]analysis.ParameterRange {
	return map[string]analysis.ParameterRange{
		"inflation_target_bps": {Min: 400, Max: 1200, Dist: "normal"},
		"staking_target_bps":   {Min: 5500, Max: 7500, Dist: "normal"},
		"base_compute_price":   {Min: 0.01, Max: 0.05, Dist: "uniform"},
		"base_storage_price":   {Min: 0.003, Max: 0.01, Dist: "uniform"},
		"base_gpu_price":       {Min: 0.08, Max: 0.2, Dist: "uniform"},
		"base_gas_price":       {Min: 0.0005, Max: 0.003, Dist: "uniform"},
		"user_demand_mean":     {Min: 6, Max: 16, Dist: "normal"},
		"user_demand_stddev":   {Min: 1, Max: 6, Dist: "normal"},
		"token_price_usd":      {Min: 0.4, Max: 3.0, Dist: "lognormal"},
	}
}
