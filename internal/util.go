package internal

// mergeConfigs merges current (runtime context) and config (step config), with current taking precedence.
func mergeConfigs(current, config map[string]any) map[string]any {
	merged := make(map[string]any, len(config)+len(current))
	for k, v := range config {
		merged[k] = v
	}
	for k, v := range current {
		merged[k] = v
	}
	return merged
}
