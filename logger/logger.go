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
func InitLogger(level string) (*zap.Logger, error) {

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return nil, fmt.Errorf("location err :%v", err)
	}
	now := time.Now()
	kst := now.In(loc).Format("2006-01-02")
	// 최상위 경로에 로그파일 설정
	fmt.Println("InitLogger 로그 파일 패스 설정")
	logFilePath := filepath.Join(".", "pds-meta-collector-"+kst+".log")

	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	// 로컬 파일로 로그 저장 설정
	fmt.Println("InitLogger 로컬 파일로 로그 저장 설정")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
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
	return zapLogger, nil
}

// GetLogger: 생성된 로거 반환
func GetLogger() *zap.Logger {
	if zapLogger == nil {
		GetSugaredLogger().Error("Logger not initialized. Call InitLogger first")
	}
	return zapLogger
}

func GetSugaredLogger() *zap.SugaredLogger {
	if sugaredLogger == nil {
		fmt.Println("sugared logger not initialized. Call InitLogger first")
	}
	return sugaredLogger
}

// DailyUploadTask: 로그를 매일 업로드하고 초기화
func DailyUploadLogger(zapLogger **zap.Logger) {
	GetSugaredLogger().Infof("Reinitializing logger for date: %s", time.Now().Format("2006-01-02"))
	if _, err := InitLogger("debug"); err != nil {
		fmt.Printf("failed to reinitialize logger: %v\n", err)
	}
	loc, _ := time.LoadLocation("Asia/Seoul")
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		// 자정 실행 대기
		now := time.Now().In(loc)
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
		time.Sleep(time.Until(nextMidnight))

		// 새로운 로그 파일 초기화
		GetSugaredLogger().Infof("Reinitializing logger for date: %s", time.Now().Format("2006-01-02"))
		if newLogger, err := InitLogger("debug"); err != nil {
			fmt.Printf("failed to reinitialize logger: %v\n", err)
		} else {
			// 새로운 logger를 zapLogger에 바인딩
			*zapLogger = newLogger
		}
		// 일주일 이상 된 로그 파일 삭제
		deleteOldLogs(".", 7)
	}
}

// deleteOldLogs: 지정된 디렉토리에서 오래된 로그 파일 삭제
func deleteOldLogs(directory string, days int) {
	loc, _ := time.LoadLocation("Asia/Seoul")
	threshold := time.Now().In(loc).AddDate(0, 0, -days)

	entries, err := os.ReadDir(directory)
	if err != nil {
		GetSugaredLogger().Errorf("failed to read log directory: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !entry.Type().IsRegular() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		filePath := filepath.Join(directory, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			GetSugaredLogger().Errorf("failed to stat file %s: %v", filePath, err)
			continue
		}

		if info.ModTime().Before(threshold) {
			if err := os.Remove(filePath); err != nil {
				GetSugaredLogger().Errorf("failed to delete file %s: %v", filePath, err)
			} else {
				GetSugaredLogger().Infof("deleted old log file: %s", filePath)
			}
		}
	}
}
