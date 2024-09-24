package common

import (
	"github.com/markusmobius/go-dateparser"
	"log/slog"
	"os"
	"strings"
	"time"
)

func GetEnvString(variable string, def string) string {
	val := strings.TrimSpace(os.Getenv(variable))
	if strings.TrimSpace(val) != "" {
		return ":" + val
	}

	return def
}

// GuessDate Note that calling this function is expensive from a CPU point of view
// so only do it if required or save the result if in a loop because we don't want to
// do it repeatedly
func GuessDate(input, fallback string) (string, int64) {
	// For guessing dates we always assume its in Sydney as that seems like a good default
	// making this a global speeds things up, but not by much...
	var guessDateLoc, _ = time.LoadLocation("Australia/Sydney")
	var guessDateCfg = &dateparser.Configuration{
		Locales:          []string{"en-AU"},
		DefaultLanguages: []string{"en"},
		CurrentTime:      time.Now().In(guessDateLoc), // needed so it can set the timezone to Australia
	}

	p, err := dateparser.Parse(guessDateCfg, input)
	if err == nil {
		return TimeToStringInt(p.Time)
	}

	_, res, err := dateparser.Search(guessDateCfg, input)
	if err == nil && len(res) != 0 {
		return TimeToStringInt(res[0].Date.Time)
	}

	slog.Error("error guessing date", "err", err, "input", input, UC("e66d371a"))
	return fallback, 0
}

func TimeToStringInt(p time.Time) (string, int64) {
	layout := "02 Jan 2006 15:04:05"
	layoutNoHour := "02 Jan 2006"

	if p.Hour() == 0 && p.Minute() == 0 {
		return p.Format(layoutNoHour), p.Unix()
	}
	return p.Format(layout), p.Unix()
}
