package model

type (
	BuildVersion struct {
		VersionNumber string `json:"version"`
		BuildDate     string `json:"buildDate"`
		GitHash       string `json:"gitHash"`
	}

	VersionBuilder interface {
		ToString() string
		ToMap() map[string]string
	}
)
