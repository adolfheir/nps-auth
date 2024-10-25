package logger

import (
	"fmt"
	"io"
	"nps-auth/configs"
	"os"
	"sync"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
	once   sync.Once
)

func GetLogger(moduleName string) zerolog.Logger {
	once.Do(func() {
		initLogger()
	})

	return Logger.With().Str("Pkg", moduleName).Logger()
}

// initLogger support:
// - output: empty (only to memory), stderr, stdout
// - format: empty (autodetect color support), color, json, text
// - time:   empty (disable timestamp), UNIXMS, UNIXMICRO, UNIXNANO
// - level:  disabled, trace, debug, info, warn, error...

func initLogger() {
	conf := configs.GetConfig()

	var modules = map[string]string{
		"format": "", // useless, but anyway
		"level":  conf.Logger.Level,
		"output": conf.Logger.Output,
		"time":   zerolog.TimeFormatUnixMs,
	}

	fmt.Printf("init logger:  %+v \n", modules)

	var writer io.Writer

	switch modules["output"] {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	case "file":
		writer = GetFileWriter()
	}

	timeFormat := modules["time"]

	if writer != nil && modules["output"] != "file" {
		if format := modules["format"]; format != "json" {
			console := &zerolog.ConsoleWriter{Out: writer}

			switch format {
			case "text":
				console.NoColor = true
			case "color":
				console.NoColor = false // useless, but anyway
			default:
				// autodetection if output support color
				// go-isatty - dependency for go-colorable - dependency for ConsoleWriter
				console.NoColor = !isatty.IsTerminal(writer.(*os.File).Fd())
			}

			if timeFormat != "" {
				console.TimeFormat = "2006/01/02 15:04:05.000"
			} else {
				console.PartsOrder = []string{
					zerolog.LevelFieldName,
					zerolog.CallerFieldName,
					zerolog.MessageFieldName,
				}
			}

			writer = console
		}
	}

	lvl, _ := zerolog.ParseLevel(modules["level"])
	Logger = zerolog.New(writer).Level(lvl)

	if timeFormat != "" {
		zerolog.TimeFieldFormat = timeFormat
		Logger = Logger.With().Timestamp().Logger()
	}
}
