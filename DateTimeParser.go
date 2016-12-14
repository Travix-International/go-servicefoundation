package servicefoundation

import (
	"strings"
	"time"
)

const (
	TEMPLATE_DATETIME                  string = "20060102 1504"
	TEMPLATE_DATETIME_SECONDS          string = "20060102 150405"
	TEMPLATE_DATETIME_SECONDS_TIMEZONE string = "20060102 150405 -07"
	TEMPLATE_DATETIME_TIMEZONE         string = "20060102 1504 -07"
	TEMPLATE_DATETIME_OUTPUT           string = "2006-01-02T15:04:05+00:00"
	TEMPLATE_DATETIME_FULL             string = "2006-01-02 15:04:05.999999999 -0700 MST"
)

type (
	DateTimeParser struct{}
)

func (p DateTimeParser) Parse(str string) time.Time {
	var parsed time.Time

	templates := []string{
		TEMPLATE_DATETIME_FULL,
		TEMPLATE_DATETIME_SECONDS_TIMEZONE,
		TEMPLATE_DATETIME_TIMEZONE,
		TEMPLATE_DATETIME_SECONDS,
		TEMPLATE_DATETIME,
	}

	// Because timezone can only be parsed as two digits we're prefixing the one-digit version here
	str = strings.Replace(str, " +", " +0", 1)
	str = strings.Replace(str, " +00", " +0", 1)
	str = strings.Replace(str, " -", " -0", 1)
	str = strings.Replace(str, " -00", " -0", 1)

	for _, t := range templates {
		var err error

		parsed, err = time.Parse(t, str)

		if err != nil {
			continue
		}

		parsed = parsed.UTC()
		break
	}

	return parsed
}

func (p DateTimeParser) ToString(dateTime time.Time) string {
	return dateTime.Format(TEMPLATE_DATETIME_OUTPUT)
}

func (p DateTimeParser) Convert(str string) string {
	return p.ToString(p.Parse(str))
}
