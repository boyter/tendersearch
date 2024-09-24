package common

import "log/slog"

func UC(code string) slog.Attr {
	return slog.String("uniqueCode", code)
}

func Err(err error) slog.Attr {
	return slog.String("err", err.Error())
}
