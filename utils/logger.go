package utils

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	today := time.Now()
	year := today.Format("2006")
	month := today.Format("01")

	// Crear la estructura de carpetas logs/YYYY/MM/DD
	logDir := "logs/" + year + "/" + month
	logFileName := logDir + "/" + today.Format("2006-01-02") + ".log"

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic("Error creating log folder structure: " + err.Error())
	}

	// Configurar el formato del log
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.MessageKey = "msg"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 3:04:05 pm")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileWriter := zapcore.AddSync(openLogFile(logFileName))

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriter, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= zapcore.InfoLevel
		})),
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.DebugLevel),
	)

	logger := zap.New(core)

	return logger
}

func openLogFile(logFileName string) *os.File {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Error opening log file: " + err.Error())
	}
	return file
}
