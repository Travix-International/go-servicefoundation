package servicefoundation

import (
	"fmt"

	"github.com/Prutswonder/go-servicefoundation/env"
)

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

const (
	unknown = "?"
)

// CreateDefaultBuildVersion creates and returns a new BuildVersion based on conventional environment variables.
func CreateDefaultBuildVersion() BuildVersion {
	return BuildVersion{
		VersionNumber: env.OrDefault("GO_PIPELINE_LABEL", unknown),
		BuildDate:     env.OrDefault("BUILD_DATE", unknown),
		GitHash:       env.OrDefault("GIT_HASH", unknown),
	}
}

// CreateDefaultVersionBuilder creates and returns a VersionBuilder based on conventional environment variables.
func CreateDefaultVersionBuilder() VersionBuilder {
	version := CreateDefaultBuildVersion()
	return CreateVersionBuilder(version)
}

// CreateVersionBuilder creates and returns a VersionBuilder for the given BuildVersion.
func CreateVersionBuilder(version BuildVersion) VersionBuilder {
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
