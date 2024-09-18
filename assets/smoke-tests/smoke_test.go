//go:build smoke

package main

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	TimeoutSeconds = 15
)

var endpoint = "http://localhost:4001/"

type apiWebEnv struct {
	Environment string `env:"SMOKE_TEST_ENVIRONMENT" env-default:"local"`
}

var (
	envVars apiWebEnv
)

func TestMain(m *testing.M) {
	err := cleanenv.ReadEnv(&envVars)
	if err != nil {
		slog.Error("error reading env vars", "error", err)
		os.Exit(1)
	}

	switch envVars.Environment {
	case "dev":
		endpoint = ""
	case "uat":
		endpoint = ""
	case "prod":
		endpoint = ""
	case "local":
	default:
		endpoint = "http://localhost:4001/"
	}

	os.Exit(m.Run())
}

func TestRoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutSeconds*time.Second)
	defer cancel()

	var resp string
	err := requests.
		URL(endpoint).
		ToString(&resp).
		Fetch(ctx)

	require.NoError(t, err)
	assert.Contains(t, resp, "tendersearch")
}

func TestHealthCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutSeconds*time.Second)
	defer cancel()

	var resp string
	err := requests.
		URL(endpoint + "v1/health-check/").
		ToString(&resp).
		Fetch(ctx)

	require.NoError(t, err)

	g := gjson.Get(resp, "ok")
	assert.True(t, g.IsBool())
	assert.Equal(t, true, g.Bool())

	g = gjson.Get(resp, "buildSha")
	assert.True(t, g.Exists())
	assert.Equal(t, 40, len(g.String()))

	g = gjson.Get(resp, "dbVersion")
	assert.True(t, g.Exists())
}
