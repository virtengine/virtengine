package rest

import (
	"fmt"

	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"
)

const (
	DeploymentPathPrefix = "/deployment/{dseq}"
	LeasePathPrefix      = "/lease/{dseq}/{gseq}/{oseq}"
	HostnamePrefix       = "/hostname"
	EndpointPrefix       = "/endpoint"
	MigratePathPrefix    = "/migrate"
)

func VersionPath() string {
	return "version"
}

func StatusPath() string {
	return "status"
}

func ValidatePath() string {
	return "validate"
}

func LeasePath(id mtypes.LeaseID) string {
	return fmt.Sprintf("lease/%d/%d/%d", id.DSeq, id.GSeq, id.OSeq)
}

func SubmitManifestPath(dseq uint64) string {
	return fmt.Sprintf("deployment/%d/manifest", dseq)
}

func GetManifestPath(id mtypes.LeaseID) string {
	return fmt.Sprintf("%s/manifest", LeasePath(id))
}

func LeaseStatusPath(id mtypes.LeaseID) string {
	return fmt.Sprintf("%s/status", LeasePath(id))
}

func LeaseShellPath(lID mtypes.LeaseID) string {
	return fmt.Sprintf("%s/shell", LeasePath(lID))
}

func LeaseEventsPath(id mtypes.LeaseID) string {
	return fmt.Sprintf("%s/kubeevents", LeasePath(id))
}

func ServiceStatusPath(id mtypes.LeaseID, service string) string {
	return fmt.Sprintf("%s/service/%s/status", LeasePath(id), service)
}

func ServiceLogsPath(id mtypes.LeaseID) string {
	return fmt.Sprintf("%s/logs", LeasePath(id))
}
