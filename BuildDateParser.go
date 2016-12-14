package servicefoundation

import (
	"fmt"
	"time"

	"github.com/Travix-International/logger"
)

const buildDateFormat = "Mon.January.2.2006.15:04:05.-0700.MST"

// Parses the build date
func ParseBuildDate(buildDate string, log *logger.Logger) time.Time {
	parsedBuildDate, err := time.Parse(buildDateFormat, buildDate)

	if err != nil {
		log.Warn("BuildDateParseError", fmt.Sprintf("Cant't parse build date %s -- %s", buildDate, err.Error()))
		parsedBuildDate = time.Now()
	}

	return parsedBuildDate
}
