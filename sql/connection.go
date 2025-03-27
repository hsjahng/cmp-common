package sql

import (
	"errors"
	"github.com/hsjahng/cmp-common/logger"
	"github.com/hsjahng/cmp-common/sql/model"
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

// db 선택을 먼저 한다
// 스키마 선택을 한다
// 이는 다른 변수로 구성해야한다

func (d DbDsn) GetDsn() string {
	return string(d)
}

type Connection interface {
	Connect() (*gorm.DB, error)
}

func NewDBConnection(dbType model.DB_TYPE, dbDsn DbDsn, logMode gormLogger.LogLevel) (*gorm.DB, error) {
	var db *gorm.DB
	switch dbType {
	case model.MARIADB:
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
