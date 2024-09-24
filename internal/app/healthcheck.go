package app

import (
	"encoding/json"
	"net/http"

	"github.com/boyter/tendersearch/internal/common"
)

type AppHealthCheckResponse struct {
	BuildSha  string `json:"buildSha"`
	DbVersion string `json:"dbVersion"`
	Env       string `json:"env"`
	Ok        bool   `json:"ok"`
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp := AppHealthCheckResponse{
		Ok:        true,
		BuildSha:  common.CommitSHA(),
		DbVersion: "TODO",
		Env:       "TODO",
	}

	marshalledResp, err := json.MarshalIndent(&resp, "", common.JsonIndent)
	if err != nil {
		return
	}
	_, err = w.Write(marshalledResp)
	if err != nil {
		return
	}
}
