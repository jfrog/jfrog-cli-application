package model

type RollbackAppVersionRequest struct {
	FromStage string `json:"from_stage"`
	Async     bool   `json:"async,omitempty"`
}

type RollbackAppVersionResponse struct {
	ApplicationKey    string `json:"application_key"`
	Version           string `json:"version"`
	ProjectKey        string `json:"project_key"`
	RollbackFromStage string `json:"rollback_from_stage"`
	RollbackToStage   string `json:"rollback_to_stage"`
}

func NewRollbackAppVersionRequest(fromStage string) *RollbackAppVersionRequest {
	return &RollbackAppVersionRequest{
		FromStage: fromStage,
		Async:     false, // Default to synchronous as per spec
	}
}
