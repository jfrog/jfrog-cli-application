package model

const (
	ImportModePathMapping = "path_mapping"
	ImportModeUnpromoted  = "unpromoted"
)

type ImportAppVersionOptions struct {
	Mode         string                    `json:"mode,omitempty"`
	PathMappings []DistributionPathMapping `json:"path_mappings,omitempty"`
}
