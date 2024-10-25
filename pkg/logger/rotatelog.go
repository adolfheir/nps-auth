package logger

import (
	"fmt"
	"io"
	"nps-auth/configs"
	"path"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var (
	writer         io.Writer
	onceInitWriter sync.Once
)

func initWriter() {
	conf := configs.GetConfig()

	fullPath := path.Join(conf.Path, "./data/log/log")
	logPath := fullPath + ".%Y%m%d"
	rl, err := rotatelogs.New(
		logPath,
		rotatelogs.WithLinkName(fullPath),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(fmt.Errorf("init log err: %w", err))
	}
	writer = rl
}

func GetFileWriter() io.Writer {

	onceInitWriter.Do(func() {
		initWriter()
	})

	return writer

}
