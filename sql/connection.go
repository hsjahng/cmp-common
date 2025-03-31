package sql

import (
	"errors"
	"github.com/hsjahng/cmp-common/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"math/rand/v2"
	"os"
	"time"
)

//todo: mock DB 구성해놓을것..
// sql 문 작성필요

// MariaDB
type DbType string

// db_common
type DbDsn string

const (
	MYSQL DbType = "MARIADB"
)

func (d DbType) String() string {
	return string(d)
}

const (
	DB_COMMON  DbDsn = "maestro:okestro2018@tcp(172.10.50.30:32006)/dp_common_pds_test?charset=utf8mb4&parseTime=True&loc=Local"
	DB_DEFAULT DbDsn = "maestro:okestro2018@tcp(172.10.50.30:32006)/dp_common_pds_test?charset=utf8mb4&parseTime=True&loc=Local"
)

func (d DbDsn) GetDsn() string {
	return string(d)
}

func NewDBConnection(dbType DbType, dbDsn DbDsn, logMode gormLogger.LogLevel) (*gorm.DB, error) {
	var db *gorm.DB
	switch dbType {
	case MYSQL:
		return GetDB(dbDsn, logMode)
	default:

	}
	return db, nil
}

// Maria DB 용
func GetDB(dbDsn DbDsn, logMode gormLogger.LogLevel) (*gorm.DB, error) {
	newLogger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logMode,
			Colorful:      false,
		},
	)
	sugaredLogger := logger.GetSugaredLogger()
	if sugaredLogger == nil {
		return nil, errors.New("sugared logger not initialized. Call InitLogger first")
	}
	sugaredLogger.Infof(dbDsn.GetDsn())

	db, err := gorm.Open(mysql.Open(dbDsn.GetDsn()), &gorm.Config{
		PrepareStmt: true,
		Logger:      newLogger,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

type RetryConfig struct {
	MaxRetries  int // 최대 재시도 횟수
	MaxInterval time.Duration
	InitialWait time.Duration // 초기 대기시간
	MaxWait     time.Duration // 최대 대기 시간
	Factor      float64       // 백오프 승수
	Jitter      float64       // 무작위성 추가
}

var DefaultRetryConfig = RetryConfig{
	MaxRetries:  10,
	InitialWait: 100 * time.Millisecond,
	MaxWait:     10 * time.Second,
	Factor:      2.0,
	Jitter:      0.1, // 무작위성
}

func RetryConnection(db *gorm.DB, config *RetryConfig) {
	//todo: connection 관련된 오류는 여기서 처리하는것으로
	var cfg RetryConfig
	if cfg.MaxRetries < 1 {
		cfg = DefaultRetryConfig
	}

	currentInterval := cfg.InitialWait

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// 첫 시도 또는 재시도 전에 지연 적용 (첫 시도는 지연 없음)
		if attempt > 0 {
			// 지터(랜덤성) 적용
			jitterRange := currentInterval.Seconds() * cfg.Jitter
			jitter := time.Duration((rand.Float64()*jitterRange*2 - jitterRange) * float64(time.Second))
			sleepTime := currentInterval + jitter
			time.Sleep(sleepTime)

			// 다음 대기 시간 계산 (지수 백오프)
			currentInterval = time.Duration(float64(currentInterval) * cfg.Factor)
			if currentInterval > cfg.MaxInterval {
				currentInterval = cfg.MaxInterval
			}
		}
	}

}

func IsNonRetryableError(err error) bool {
	// 무시해도 될 오류들은 지나치도록 한다
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	return false
}
