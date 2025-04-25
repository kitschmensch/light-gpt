package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

type Logger struct {
	degubLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewLogger(logLevel, logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	debugLogger := log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger := log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	logger := &Logger{
		degubLogger: debugLogger,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
	}

	if strings.ToLower(logLevel) == "info" {
		logger.degubLogger.SetOutput(io.Discard)
	}

	if strings.ToLower(logLevel) == "error" {
		logger.degubLogger.SetOutput(io.Discard)
		logger.infoLogger.SetOutput(io.Discard)
	}

	return logger, nil
}

func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}
