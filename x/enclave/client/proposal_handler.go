package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/virtengine/virtengine/x/enclave/client/cli"
)

// ProposalHandlerAddMeasurement is the governance proposal handler for adding measurements.
var ProposalHandlerAddMeasurement = govclient.NewProposalHandler(cli.NewCmdSubmitAddMeasurementProposal)

// ProposalHandlerRevokeMeasurement is the governance proposal handler for revoking measurements.
var ProposalHandlerRevokeMeasurement = govclient.NewProposalHandler(cli.NewCmdSubmitRevokeMeasurementProposal)
