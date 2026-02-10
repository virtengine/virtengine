package events

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/pubsub"
	tmquery "github.com/cometbft/cometbft/libs/pubsub/query"
	tmtypes "github.com/cometbft/cometbft/types"
)

func blkHeaderQuery() pubsub.Query {
	return tmquery.MustCompile(fmt.Sprintf("%s='%s'", tmtypes.EventTypeKey, tmtypes.EventNewBlockHeader))
}
