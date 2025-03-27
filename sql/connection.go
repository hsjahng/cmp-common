package sql

import (
	"errors"
	"github.com/hsjahng/cmp-common/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

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
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Factor      float64
	Jitter      float64
}

var DefaultRetryConfig = RetryConfig{
	MaxRetries:  5,
	InitialWait: 100 * time.Millisecond,
	MaxWait:     10 * time.Second,
	Factor:      2.0,
	Jitter:      0.1,
}

func Connection() {
	// retry backoff 를 추가해야함
	// connection 관련된 오류는 여기서 처리해야함

	// mock DB 를 만들어놓는것도 좋을 것 같다

}
