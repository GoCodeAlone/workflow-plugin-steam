// Package workflowpluginsteam provides the Steam workflow plugin.
package workflowpluginsteam

import (
	"github.com/GoCodeAlone/workflow-plugin-steam/internal"
	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// NewSteamPlugin returns the Steam SDK plugin provider.
func NewSteamPlugin() sdk.PluginProvider {
	return internal.NewSteamPlugin()
}
