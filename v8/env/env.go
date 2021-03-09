package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const listSeparator = ","

// OrDefault returns the value of the environment variable (name). If empty, it returns defaultValue.
func OrDefault(name, defaultValue string) string {
	strValue := os.Getenv(name)

	if strValue == "" {
		return defaultValue
	}
	return strValue
}

// List returns the value of the environment variable (name) as a list.
func List(name string) []string {
	return strings.Split(os.Getenv(name), listSeparator)
}

// ListOrDefault returns the value of the environment variable (name) as a list. If not defined, returns a default.
func ListOrDefault(name string, defaultList []string) []string {
	value := os.Getenv(name)

	if value == "" {
		return defaultList
	}
	return strings.Split(os.Getenv(name), listSeparator)
}

// AsInt returns the value of the environment variable (name) as an int. If empty, it returns defaultValue.
func AsInt(name string, defaultValue int) int {
	strValue := os.Getenv(name)

	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		panic(fmt.Errorf("Failed parsing %s [%s]: %v", name, strValue, err))
	}
	return value
}
