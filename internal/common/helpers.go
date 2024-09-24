package common

import (
	"os"
	"strings"
)

func GetEnvString(variable string, def string) string {
	val := strings.TrimSpace(os.Getenv(variable))
	if strings.TrimSpace(val) != "" {
		return ":" + val
	}

	return def
}
