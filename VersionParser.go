package servicefoundation

import (
	"fmt"
	"time"
)

func ParseVersion(mainVersion string, minVersion string, parsedBuildDate time.Time, gitHash string) string {
	return fmt.Sprintf("version: %s%s - buildDate: %s - git hash: %s", mainVersion, minVersion, parsedBuildDate.Local(), gitHash)
}
