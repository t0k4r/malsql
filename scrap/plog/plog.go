package plog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Plog struct {
}

func NewPlog() *slog.Logger {
	return slog.New(&Plog{})
}

func (p *Plog) Enabled(context.Context, slog.Level) bool {
	return true
}
func (p *Plog) Handle(ctx context.Context, rec slog.Record) error {
	var out strings.Builder
	switch rec.Level {
	case slog.LevelInfo:
		out.WriteString(fmt.Sprintf("\033[0;32m%v:\033[0m\n", rec.Message))
		rec.Attrs(func(a slog.Attr) bool {
			out.WriteString(fmt.Sprintf("\t%v: %v\n", a.Key, a.Value))
			return true
		})
	case slog.LevelWarn:
		out.WriteString(fmt.Sprintf("\033[0;33m%v\033[0m\n", rec.Message))
		rec.Attrs(func(a slog.Attr) bool {
			out.WriteString(fmt.Sprintf("\t%v: %v\n", a.Key, a.Value))
			return true
		})
	case slog.LevelError:
		out.WriteString(fmt.Sprintf("\033[0;31m%v:\033[0m\n", rec.Message))
		rec.Attrs(func(a slog.Attr) bool {
			out.WriteString(fmt.Sprintf("\t%v: %v\n", a.Key, a.Value))
			return true
		})
	}
	fmt.Print(out.String())
	return nil
}
func (p *Plog) WithAttrs(attrs []slog.Attr) slog.Handler {
	return p
}
func (p *Plog) WithGroup(name string) slog.Handler {
	return p
}
