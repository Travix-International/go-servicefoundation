package site

import (
	"fmt"

	"github.com/Prutswonder/go-servicefoundation/env"
	. "github.com/Prutswonder/go-servicefoundation/model"
)

const (
	unknown = "?"
)

type versionBuilderImpl struct {
	version BuildVersion
}

func CreateBuildVersion() BuildVersion {
	return BuildVersion{
		VersionNumber: env.OrDefault("GO_PIPELINE_LABEL", unknown),
		BuildDate:     env.OrDefault("BUILD_DATE", unknown),
		GitHash:       env.OrDefault("GIT_HASH", unknown),
	}
}

func CreateDefaultVersionBuilder() VersionBuilder {
	version := CreateBuildVersion()
	return CreateVersionBuilder(version)
}

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
