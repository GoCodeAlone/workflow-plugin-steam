package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// steamPlugin implements sdk.PluginProvider.
type steamPlugin struct{}

// NewSteamPlugin returns the Steam platform plugin provider.
func NewSteamPlugin() sdk.PluginProvider {
	return &steamPlugin{}
}

func (p *steamPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-steam",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "Steam platform integration for workflow engine: auth, achievements, leaderboards, rich presence, and friends",
	}
}

func (p *steamPlugin) ModuleTypes() []string {
	return []string{}
}

func (p *steamPlugin) StepTypes() []string {
	return []string{
		// Auth
		"step.steam_auth",

		// Achievements & stats
		"step.steam_achievement_set",
		"step.steam_achievement_sync",
		"step.steam_stat_set",
		"step.steam_leaderboard_push",

		// Rich Presence & social
		"step.steam_presence_set",
		"step.steam_friends_list",
		"step.steam_invite_send",
	}
}

func (p *steamPlugin) CreateModule(typeName, _ string, _ map[string]any) (sdk.ModuleInstance, error) {
	return nil, fmt.Errorf("unknown module type %q", typeName)
}

func (p *steamPlugin) CreateStep(typeName, name string, _ map[string]any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.steam_auth":
		return newSteamAuthStep(name), nil
	case "step.steam_achievement_set":
		return newSteamAchievementSetStep(name), nil
	case "step.steam_achievement_sync":
		return newSteamAchievementSyncStep(name), nil
	case "step.steam_stat_set":
		return newSteamStatSetStep(name), nil
	case "step.steam_leaderboard_push":
		return newSteamLeaderboardPushStep(name), nil
	case "step.steam_presence_set":
		return newSteamPresenceSetStep(name), nil
	case "step.steam_friends_list":
		return newSteamFriendsListStep(name), nil
	case "step.steam_invite_send":
		return newSteamInviteSendStep(name), nil
	default:
		return nil, fmt.Errorf("unknown step type %q", typeName)
	}
}
