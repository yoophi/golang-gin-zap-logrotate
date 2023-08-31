package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	writer "github.com/utahta/go-cronowriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logDir := "./logs"
	archiveDir := "./archived"

	w1 := writer.MustNew(logDir+"/example.log.%Y%m%d_%H%M",
		writer.WithInit(), writer.WithMutex(),
	)
	w2 := writer.MustNew(logDir+"/internal_error.log.%Y%m%d_%H%M",
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

	logger.Info("hello, world!")

	s := gocron.NewScheduler(time.UTC)

	job, err := s.Every("1m").Do(func() {
		sourceDir := logDir
		destDir := archiveDir
		pattern := "example.log.*"

		err := os.MkdirAll(destDir, 0755)
		if err != nil {
			logger.Error(fmt.Sprintf("mkdir failed: %+v", err))
			return
		}

		matches, err := filepath.Glob(filepath.Join(sourceDir, pattern))
		if err != nil {
			logger.Error(fmt.Sprintf("error matching files: %+v", err))
			return
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i] > matches[j]
		})

		for i, match := range matches {
			if i == 0 {
				fmt.Println("first element:", match)
				continue
			}
			fileName := filepath.Base(match)
			destFilePath := filepath.Join(destDir, fileName)
			fmt.Println("$ mv", fileName, destFilePath)

			err := os.Rename(match, destFilePath)
			if err != nil {
				logger.Error(fmt.Sprintf("error moving file: %s", match))
			} else {
				logger.Info(fmt.Sprintf("moved %s to %s", match, destFilePath))
			}
		}
	})
	fmt.Printf("%+v\n", job)
	if err != nil {
		return
	}
	s.StartAsync()

	r := gin.New()
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
