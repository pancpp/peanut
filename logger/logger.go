package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"path"

	"github.com/pancpp/peanut/conf"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init() error {
	logFile := conf.GetString("log.path")
	logDir := path.Dir(logFile)

	// check log directory existence
	if _, err := os.Stat(logDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}

	fileLogWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    conf.GetInt("log.max_size"), // MB
		MaxBackups: conf.GetInt("log.max_backups"),
		LocalTime:  conf.GetBool("log.local_time"),
		Compress:   conf.GetBool("log.compress"),
	}
	if conf.GetBool("log.enable_console_log") {
		log.SetOutput(io.MultiWriter(fileLogWriter, os.Stderr))
	} else {
		log.SetOutput(fileLogWriter)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}
