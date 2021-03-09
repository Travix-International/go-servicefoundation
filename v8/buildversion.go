package v8

import "fmt"

type (
	// BuildVersion contains the version and build information of the application.
	BuildVersion struct {
		VersionNumber string `json:"version"`
		BuildDate     string `json:"buildDate"`
		GitHash       string `json:"gitHash"`
	}

	// VersionBuilder contains methods to output version information in string format.
	VersionBuilder interface {
		ToString() string
		ToMap() map[string]string
	}

	versionBuilderImpl struct {
		version BuildVersion
	}
)

// NewVersionBuilder creates and returns a VersionBuilder for the given BuildVersion.
func NewVersionBuilder(version BuildVersion) VersionBuilder {
	return &versionBuilderImpl{
		version: version,
	}
}

/* VersionBuilder implementation */

func (b *versionBuilderImpl) ToString() string {
	v := b.version
	return fmt.Sprintf("version: %s - buildDate: %s - git hash: %s", v.VersionNumber, v.BuildDate, v.GitHash)
}

func (b *versionBuilderImpl) ToMap() map[string]string {
	return map[string]string{
		"version":   b.version.VersionNumber,
		"buildDate": b.version.BuildDate,
		"gitHash":   b.version.GitHash,
	}
}
