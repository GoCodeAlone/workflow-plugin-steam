// Command workflow-plugin-steam is a workflow engine external plugin that
// provides Steam platform integration (auth, achievements, leaderboards,
// rich presence, and friends).
// It runs as a subprocess and communicates with the host engine via
// the go-plugin protocol.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-steam/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.NewSteamPlugin())
}
