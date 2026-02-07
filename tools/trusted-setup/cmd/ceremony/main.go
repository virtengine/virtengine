package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/tools/trusted-setup/coordinator"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	"github.com/virtengine/virtengine/tools/trusted-setup/verify"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ceremony",
		Short: "VirtEngine ZK trusted setup ceremony tooling",
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(startPhase2Cmd())
	rootCmd.AddCommand(contributePhase1Cmd())
	rootCmd.AddCommand(contributePhase2Cmd())
	rootCmd.AddCommand(acceptPhase1Cmd())
	rootCmd.AddCommand(acceptPhase2Cmd())
	rootCmd.AddCommand(finalizeCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(participateCmd())
	rootCmd.AddCommand(verifyTranscriptCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	var (
		dir             string
		circuit         string
		minContributors int
		beacon          string
		notes           []string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a ceremony workspace",
		RunE: func(_ *cobra.Command, _ []string) error {
			state := coordinator.State{BaseDir: dir}
			_, err := coordinator.InitCeremony(state, circuit, minContributors, beacon, notes)
			return err
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	cmd.Flags().StringVar(&circuit, "circuit", "age-range", "circuit name (age-range|residency|score-range)")
	cmd.Flags().IntVar(&minContributors, "min-contributors", 20, "minimum contributors required")
	cmd.Flags().StringVar(&beacon, "beacon", "", "randomness beacon")
	cmd.Flags().StringArrayVar(&notes, "note", nil, "ceremony notes (repeatable)")
	_ = cmd.MarkFlagRequired("circuit")
	return cmd
}

func startPhase2Cmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "start-phase2",
		Short: "Start phase2 after enough phase1 contributions",
		RunE: func(_ *cobra.Command, _ []string) error {
			state := coordinator.State{BaseDir: dir}
			return coordinator.StartPhase2(state)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	return cmd
}

func contributePhase1Cmd() *cobra.Command {
	var (
		inPath        string
		outPath       string
		metaPath      string
		identity      string
		participantID string
		attestation   string
	)
	cmd := &cobra.Command{
		Use:   "contribute-phase1",
		Short: "Create a phase1 contribution offline",
		RunE: func(_ *cobra.Command, _ []string) error {
			payload, err := os.ReadFile(inPath)
			if err != nil {
				return err
			}

			id, err := participant.LoadOrCreateIdentity(identity, participantID)
			if err != nil {
				return err
			}
			client := participant.NewClient(id, attestation)
			output, signature, err := client.ContributePhase1(payload)
			if err != nil {
				return err
			}

			if err := os.WriteFile(outPath, output, 0o600); err != nil {
				return err
			}

			meta := coordinator.ContributionMeta{
				ParticipantID: id.ID,
				PublicKey:     id.PublicKey,
				Signature:     signature,
				Attestation:   attestation,
			}
			if err := writeMeta(metaPath, meta); err != nil {
				return err
			}

			fmt.Printf("Contribution hash: %s\n", transcript.HashBytes(output))
			fmt.Printf("Metadata written: %s\n", metaPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "phase1_latest.bin", "input phase1 transcript")
	cmd.Flags().StringVar(&outPath, "out", "phase1_contrib.bin", "output phase1 contribution")
	cmd.Flags().StringVar(&metaPath, "meta", "phase1_meta.json", "output metadata JSON")
	cmd.Flags().StringVar(&identity, "identity", "participant_identity.json", "identity file path")
	cmd.Flags().StringVar(&participantID, "participant", "participant", "participant identifier")
	cmd.Flags().StringVar(&attestation, "attestation", "", "participant attestation")
	_ = cmd.MarkFlagRequired("in")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func contributePhase2Cmd() *cobra.Command {
	var (
		inPath        string
		outPath       string
		metaPath      string
		identity      string
		participantID string
		attestation   string
	)
	cmd := &cobra.Command{
		Use:   "contribute-phase2",
		Short: "Create a phase2 contribution offline",
		RunE: func(_ *cobra.Command, _ []string) error {
			payload, err := os.ReadFile(inPath)
			if err != nil {
				return err
			}

			id, err := participant.LoadOrCreateIdentity(identity, participantID)
			if err != nil {
				return err
			}
			client := participant.NewClient(id, attestation)
			output, signature, err := client.ContributePhase2(payload)
			if err != nil {
				return err
			}

			if err := os.WriteFile(outPath, output, 0o600); err != nil {
				return err
			}

			meta := coordinator.ContributionMeta{
				ParticipantID: id.ID,
				PublicKey:     id.PublicKey,
				Signature:     signature,
				Attestation:   attestation,
			}
			if err := writeMeta(metaPath, meta); err != nil {
				return err
			}

			fmt.Printf("Contribution hash: %s\n", transcript.HashBytes(output))
			fmt.Printf("Metadata written: %s\n", metaPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "phase2_latest.bin", "input phase2 transcript")
	cmd.Flags().StringVar(&outPath, "out", "phase2_contrib.bin", "output phase2 contribution")
	cmd.Flags().StringVar(&metaPath, "meta", "phase2_meta.json", "output metadata JSON")
	cmd.Flags().StringVar(&identity, "identity", "participant_identity.json", "identity file path")
	cmd.Flags().StringVar(&participantID, "participant", "participant", "participant identifier")
	cmd.Flags().StringVar(&attestation, "attestation", "", "participant attestation")
	_ = cmd.MarkFlagRequired("in")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func acceptPhase1Cmd() *cobra.Command {
	var (
		dir     string
		payload string
		meta    string
	)
	cmd := &cobra.Command{
		Use:   "accept-phase1",
		Short: "Accept a phase1 contribution (coordinator)",
		RunE: func(_ *cobra.Command, _ []string) error {
			payloadBytes, err := os.ReadFile(payload)
			if err != nil {
				return err
			}
			metaValue, err := readMeta(meta)
			if err != nil {
				return err
			}
			state := coordinator.State{BaseDir: dir}
			return coordinator.AcceptPhase1Contribution(state, payloadBytes, metaValue)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	cmd.Flags().StringVar(&payload, "payload", "", "contribution payload file")
	cmd.Flags().StringVar(&meta, "meta", "", "contribution metadata JSON")
	_ = cmd.MarkFlagRequired("payload")
	_ = cmd.MarkFlagRequired("meta")
	return cmd
}

func acceptPhase2Cmd() *cobra.Command {
	var (
		dir     string
		payload string
		meta    string
	)
	cmd := &cobra.Command{
		Use:   "accept-phase2",
		Short: "Accept a phase2 contribution (coordinator)",
		RunE: func(_ *cobra.Command, _ []string) error {
			payloadBytes, err := os.ReadFile(payload)
			if err != nil {
				return err
			}
			metaValue, err := readMeta(meta)
			if err != nil {
				return err
			}
			state := coordinator.State{BaseDir: dir}
			return coordinator.AcceptPhase2Contribution(state, payloadBytes, metaValue)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	cmd.Flags().StringVar(&payload, "payload", "", "contribution payload file")
	cmd.Flags().StringVar(&meta, "meta", "", "contribution metadata JSON")
	_ = cmd.MarkFlagRequired("payload")
	_ = cmd.MarkFlagRequired("meta")
	return cmd
}

func finalizeCmd() *cobra.Command {
	var (
		dir     string
		version string
	)
	cmd := &cobra.Command{
		Use:   "finalize",
		Short: "Finalize ceremony and output proving/verifying keys",
		RunE: func(_ *cobra.Command, _ []string) error {
			state := coordinator.State{BaseDir: dir}
			_, _, err := coordinator.Finalize(state, version)
			return err
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	cmd.Flags().StringVar(&version, "version", "v1", "parameter version label")
	return cmd
}

func statusCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show ceremony status summary",
		RunE: func(_ *cobra.Command, _ []string) error {
			state := coordinator.State{BaseDir: dir}
			snapshot, err := coordinator.StatusSnapshot(state)
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(snapshot, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	return cmd
}

func serverCmd() *cobra.Command {
	var (
		dir  string
		addr string
	)
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run coordinator HTTP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			state := coordinator.State{BaseDir: dir}
			server := coordinator.Server{State: state}
			return server.Run(context.Background(), addr)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	cmd.Flags().StringVar(&addr, "addr", ":8080", "listen address")
	return cmd
}

func participateCmd() *cobra.Command {
	var (
		baseURL       string
		identity      string
		participantID string
		attestation   string
		phase         string
	)
	cmd := &cobra.Command{
		Use:   "participate",
		Short: "Participate in the ceremony via coordinator",
		RunE: func(_ *cobra.Command, _ []string) error {
			id, err := participant.LoadOrCreateIdentity(identity, participantID)
			if err != nil {
				return err
			}
			client := participant.NewClient(id, attestation)
			ctx := context.Background()

			payload, err := fetchBinary(ctx, baseURL, phase)
			if err != nil {
				return err
			}

			var contrib []byte
			var signature string
			if phase == "phase1" {
				contrib, signature, err = client.ContributePhase1(payload)
			} else {
				contrib, signature, err = client.ContributePhase2(payload)
			}
			if err != nil {
				return err
			}

			meta := coordinator.ContributionMeta{
				ParticipantID: id.ID,
				PublicKey:     id.PublicKey,
				Signature:     signature,
				Attestation:   attestation,
			}
			return submitContribution(ctx, baseURL, phase, contrib, meta)
		},
	}
	cmd.Flags().StringVar(&baseURL, "url", "http://localhost:8080", "coordinator URL")
	cmd.Flags().StringVar(&identity, "identity", "participant_identity.json", "identity file path")
	cmd.Flags().StringVar(&participantID, "participant", "participant", "participant identifier")
	cmd.Flags().StringVar(&attestation, "attestation", "", "participant attestation")
	cmd.Flags().StringVar(&phase, "phase", "phase1", "phase1 or phase2")
	return cmd
}

func verifyTranscriptCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "verify-transcript",
		Short: "Verify ceremony transcript and outputs",
		RunE: func(_ *cobra.Command, _ []string) error {
			result, err := verify.Verify(dir)
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "ceremony-data", "ceremony working directory")
	return cmd
}

func writeMeta(path string, meta coordinator.ContributionMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func readMeta(path string) (coordinator.ContributionMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return coordinator.ContributionMeta{}, err
	}
	var meta coordinator.ContributionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return coordinator.ContributionMeta{}, err
	}
	return meta, nil
}

func fetchBinary(ctx context.Context, baseURL, phase string) ([]byte, error) {
	endpoint := "api/v1/phase1/current"
	if phase == "phase2" {
		endpoint = "api/v1/phase2/current"
	}
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coordinator error: %s", string(body))
	}
	return io.ReadAll(resp.Body)
}

func submitContribution(ctx context.Context, baseURL, phase string, payload []byte, meta coordinator.ContributionMeta) error {
	endpoint := "api/v1/phase1/contribute"
	if phase == "phase2" {
		endpoint = "api/v1/phase2/contribute"
	}
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Participant-Id", meta.ParticipantID)
	req.Header.Set("X-Public-Key", meta.PublicKey)
	req.Header.Set("X-Signature", meta.Signature)
	req.Header.Set("X-Attestation", meta.Attestation)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("coordinator error: %s", string(body))
	}
	return nil
}
