package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/boyter/tendersearch/assets"
	"log/slog"
	"net/http"
	"os"

	"github.com/boyter/tendersearch/internal/app"
	"github.com/boyter/tendersearch/internal/common"
	"github.com/boyter/tendersearch/internal/service"
	_ "modernc.org/sqlite"
)

func main() {
	// setup logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	sqliteDb, err := connectSqliteDb("tenders")
	if err != nil {
		slog.Error("error", common.Err(err), common.UC("8457b854"))
		return
	}

	// create the tables and ignore the error because it shouldn't matter
	if _, err = sqliteDb.Exec(assets.Schema); err != nil {
		slog.Error("error", common.Err(err), common.UC("43af884d"))
	}

	ser := service.NewService(sqliteDb)
	application := app.NewApplication(ser)

	err = application.ParseTemplates()
	if err != nil {
		slog.Error("problem parsing templates", common.Err(err), common.UC("362e6e55"))
		return
	}

	addr := common.GetEnvString("HTTP_SERVER_PORT", ":8000")

	slog.Info("serving", "addr", addr, common.UC("1876ce1e"))
	errStr := http.ListenAndServe(addr, application.Router()).Error()
	slog.Error("error", "err", errStr, common.UC("69e6ef69"))
}

func connectSqliteDb(name string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("%s.sqlite?_busy_timeout=5000", name))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`pragma journal_mode = wal;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 268435456;
pragma foreign_keys = on;`)

	return db, err
}
