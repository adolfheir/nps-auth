package sql

import (
	"context"
	"time"

	pkgLogger "nps-auth/pkg/logger"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

// 自定义 GORM Logger 使用 zerolog
type ZerologGormLogger struct {
	logger zerolog.Logger
}

func newLogger() *ZerologGormLogger {
	return &ZerologGormLogger{
		logger: pkgLogger.GetLogger("gorm"),
	}
}

func (l *ZerologGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// 根据日志级别切换模式

	return l
}

func (l *ZerologGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Info().Msgf(msg, data...)
}

func (l *ZerologGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warn().Msgf(msg, data...)
}

func (l *ZerologGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Error().Msgf(msg, data...)
}

func (l *ZerologGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	l.logger.Info().
		Dur("elapsed", elapsed).
		Str("sql", sql).
		Int64("rows", rows).
		Err(err).
		Msg("SQL Trace")
}
