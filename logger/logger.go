package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

var zapLogger *zap.Logger
var sugaredLogger *zap.SugaredLogger

// InitLogger: 로그 파일을 최상위 디렉토리에 날짜별로 저장
func InitLogger(level string) {

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Printf("location err :%v\n", err)
	}

	now := time.Now()
	kst := now.In(loc).Format("2006-01-02")
	// 최상위 경로에 로그파일 설정
	fmt.Println("InitLogger 로그 파일 패스 설정")
	logFilePath := filepath.Join(".", "pds-integration-service-"+kst+".log")

	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		fmt.Printf("ParseLevel err :%v\n", err)
	}

	// 로컬 파일로 로그 저장 설정
	fmt.Println("InitLogger 로컬 파일로 로그 저장 설정")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("OpenFile err :%v\n", err)
	}
	fileSync := zapcore.AddSync(logFile)

	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.TimeKey = "timestamp"
	consoleEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		kst := t.UTC().Add(9 * time.Hour)
		enc.AppendString(kst.Format("2006-01-02T15:04:05"))
	}

	// 콘솔 출력과 파일 출력 설정
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel), // 콘솔 출력
		zapcore.NewCore(consoleEncoder, fileSync, logLevel),                   // 파일 출력
	)

	// Zap 로거 생성
	zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugaredLogger = zapLogger.Sugar()
}

func GetSugaredLogger() *zap.SugaredLogger {
	if sugaredLogger == nil {
		GetSugaredLogger().Error("Sugared logger not initialized. Call InitLogger first")
	}
	return sugaredLogger
}
