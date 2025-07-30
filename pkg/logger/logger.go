package logger

import (
	// "net/http"
	"os"
	// "time"

	// "github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func SetupLogger() zerolog.Logger {
	os.MkdirAll("logs", os.ModePerm)

	file, err := os.OpenFile("/var/log/app/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	multi := zerolog.MultiLevelWriter(os.Stdout, file)

	logger := zerolog.New(multi).With().Timestamp().Logger()
	return logger
}
