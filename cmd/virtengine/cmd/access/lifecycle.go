package access

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/virtengine/virtengine/pkg/security"
)

type lifecycleRequest struct {
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type lifecycleResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	OperationID string `json:"operation_id,omitempty"`
	State       string `json:"state,omitempty"`
	Error       string `json:"error,omitempty"`
}

// GetLifecycleCmd returns the lifecycle command group.
func GetLifecycleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "lifecycle",
		Short:                      "Manage resource lifecycles via provider portal API",
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newLifecycleActionCmd("start"),
		newLifecycleActionCmd("stop"),
		newLifecycleActionCmd("restart"),
		newLifecycleActionCmd("resize"),
	)

	return cmd
}

func newLifecycleActionCmd(action string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <deployment-id>", action),
		Short: fmt.Sprintf("%s a deployment", titleAction(action)),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := cmd.Flags().GetString("endpoint")
			if err != nil {
				return err
			}
			principal, err := cmd.Flags().GetString("principal")
			if err != nil {
				return err
			}
			secret, err := cmd.Flags().GetString("secret")
			if err != nil {
				return err
			}
			if endpoint == "" {
				return fmt.Errorf("endpoint is required")
			}
			if principal == "" || secret == "" {
				return fmt.Errorf("principal and secret are required")
			}

			parameters := map[string]interface{}{}
			if action == "resize" {
				if cpu, _ := cmd.Flags().GetUint32("cpu"); cpu > 0 {
					parameters["cpu_cores"] = cpu
				}
				if memory, _ := cmd.Flags().GetUint64("memory"); memory > 0 {
					parameters["memory_mb"] = memory
				}
				if storage, _ := cmd.Flags().GetUint64("storage"); storage > 0 {
					parameters["storage_gb"] = storage
				}
				if gpu, _ := cmd.Flags().GetUint32("gpu"); gpu > 0 {
					parameters["gpu_count"] = gpu
				}
				if len(parameters) == 0 {
					return fmt.Errorf("at least one resize parameter must be specified")
				}
			}

			request := lifecycleRequest{Action: action}
			if len(parameters) > 0 {
				request.Parameters = parameters
			}

			body, err := json.Marshal(request)
			if err != nil {
				return err
			}

			reqURL, err := buildLifecycleURL(endpoint, args[0])
			if err != nil {
				return err
			}

			resp, err := doHMACRequest(reqURL, principal, secret, body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var out lifecycleResponse
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				if out.Error != "" {
					return errors.New(out.Error)
				}
				return fmt.Errorf("request failed with status %s", resp.Status)
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(out)
		},
	}

	cmd.Flags().String("endpoint", "", "Provider portal API base URL")
	cmd.Flags().String("principal", "", "HMAC principal identifier")
	cmd.Flags().String("secret", "", "HMAC secret")

	if action == "resize" {
		cmd.Flags().Uint32("cpu", 0, "CPU cores")
		cmd.Flags().Uint64("memory", 0, "Memory in MB")
		cmd.Flags().Uint64("storage", 0, "Storage in GB")
		cmd.Flags().Uint32("gpu", 0, "GPU count")
	}

	return cmd
}

func titleAction(action string) string {
	if action == "" {
		return action
	}
	return strings.ToUpper(action[:1]) + action[1:]
}

func buildLifecycleURL(endpoint, deploymentID string) (string, error) {
	base, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/api/v1/deployments/" + deploymentID + "/actions"
	return base.String(), nil
}

func doHMACRequest(reqURL, principal, secret string, body []byte) (*http.Response, error) {
	parsed, err := url.Parse(reqURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := computeHMACSignature(req.Method, parsed.Path, parsed.RawQuery, principal, timestamp, secret)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-VE-Principal", principal)
	req.Header.Set("X-VE-Timestamp", timestamp)
	req.Header.Set("X-VE-Signature", signature)

	client := security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second))
	return client.Do(req)
}

func computeHMACSignature(method, path, query, principal, timestamp, secret string) string {
	payload := strings.Join([]string{method, path, query, principal, timestamp}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
