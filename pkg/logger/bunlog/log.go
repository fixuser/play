package bunlog

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
)

type QueryHookOptions struct {
	LogSlow time.Duration
}

// QueryHook wraps query hook
type QueryHook struct {
	opts QueryHookOptions
}

func NewQueryHook(opts QueryHookOptions) *QueryHook {
	return &QueryHook{opts: opts}
}

// BeforeQuery does nothing tbh
func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery convert a bun QueryEvent into a logrus message
func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !viper.GetBool("log.traced") {
		return
	}

	now := time.Now()
	dur := now.Sub(event.StartTime)

	logger := log.Ctx(ctx).With().Str("op", eventOperation(event)).Dur("duration", dur).Str("sql", event.Query).Logger()
	if event.Err != nil {
		logger.Error().Err(event.Err).Msg("query failed")
		return
	}
	if h.opts.LogSlow > 0 && dur > h.opts.LogSlow {
		logger.Warn().Msg("slow sql")
		return
	}
	logger.Debug().Msg("sql")
}

// taken from bun
func eventOperation(event *bun.QueryEvent) string {
	switch event.IQuery.(type) {
	case *bun.SelectQuery:
		return "SELECT"
	case *bun.InsertQuery:
		return "INSERT"
	case *bun.UpdateQuery:
		return "UPDATE"
	case *bun.DeleteQuery:
		return "DELETE"
	case *bun.CreateTableQuery:
		return "CREATE TABLE"
	case *bun.DropTableQuery:
		return "DROP TABLE"
	}
	return queryOperation(event.Query)
}

// taken from bun
func queryOperation(name string) string {
	if idx := strings.Index(name, " "); idx > 0 {
		name = name[:idx]
	}
	if len(name) > 16 {
		name = name[:16]
	}
	return string(name)
}
