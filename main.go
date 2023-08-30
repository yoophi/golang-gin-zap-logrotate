package main

import (
	"fmt"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	writer "github.com/utahta/go-cronowriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	baseDir := "./logs"
	w1 := writer.MustNew(baseDir+"/example.log.%Y%m%d_%H%M",
		writer.WithInit(), writer.WithMutex(),
	)
	w2 := writer.MustNew(baseDir+"/internal_error.log.%Y%m%d_%H%M",
		writer.WithInit(), writer.WithMutex(),
	)
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(w1),
			zapcore.InfoLevel,
		),
		zap.ErrorOutput(zapcore.AddSync(w2)),
	)

	// ---
	r := gin.New()

	//logger, _ := zap.NewProduction()

	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})

	r.Run(":8080")
}
