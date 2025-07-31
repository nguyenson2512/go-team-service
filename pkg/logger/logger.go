package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func SetupLogger() {
	os.MkdirAll("logs", os.ModePerm)

	file, err := os.OpenFile("/var/log/app/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	multi := zerolog.MultiLevelWriter(os.Stdout, file)

	Logger = zerolog.New(multi).With().Timestamp().Logger()
}
