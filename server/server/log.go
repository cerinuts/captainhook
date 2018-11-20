/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/


package server

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log *logrus.Logger

// InitLogger initalizes the logger with some settings
func InitLogger() {
	log = logrus.New()
	level, err := logrus.ParseLevel(viper.GetString("Loglevel"))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	if viper.GetBool("Debug") {
		log.SetReportCaller(true)
		log.SetLevel(logrus.DebugLevel)
	}
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	badgerLog := log.WithField("subsystem", "Badger")

	badger.SetLogger(badgerLog)
}

func getGinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {

		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)

		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := logrus.NewEntry(log).WithFields(logrus.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency,
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  userAgent,
			"subsystem":  "gin",
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(time.RFC3339), c.Request.Method, path, statusCode, dataLength, referer, userAgent, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
