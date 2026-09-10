package model

type AppVersionExportStatus struct {
	Status              string `json:"status"`
	DownloadURL         string `json:"download_url,omitempty"`
	RelativeDownloadURL string `json:"relative_download_url,omitempty"`
	Message             string `json:"message,omitempty"`
}
