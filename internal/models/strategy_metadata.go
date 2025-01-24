package models

// StrategyMetadata represents available strategy information
type StrategyMetadata struct {
	Name       string          `json:"name"`
	Parameters []ParameterInfo `json:"parameters"`
	Flow       []string        `json:"strategy_flow"`
}

// ParameterInfo describes a strategy parameter
type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}
