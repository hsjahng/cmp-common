package sql

import (
	"cmp-common/sql/model"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB
var Dsn string

type DB struct {
	Dsn    string
	DbType model.DB_TYPE
	DB     *sqlx.DB
}

type Connection interface {
	Connect() (*sqlx.DB, error)
}

func NewDBConnection(dbType model.DB_TYPE) {

	switch dbType {
	case model.MARIADB:
		Dsn = "maestro:okestro2018@tcp(172.10.50.30:32006)/dp_common_pds_test?charset=utf8mb4&parseTime=True&loc=Local"
	//case model.PROMETHEUS:
	//	Dsn = "prom-prometheus-prometheus-test:9090"
	default:

	}
}

func (m *DB) GetDB() (*sqlx.DB, error) {
	//logger.GetSugaredLogger().Debugf("MariaDB Conntect 실행")
	//defer func() {
	//	if r := recover(); r != nil {
	//		errMsg := fmt.Sprintf("panic occurred: %v", r)
	//		logger.GetSugaredLogger().Error(errMsg)
	//		err = errors.New(errMsg)
	//		return
	//	}
	//}()
	//
	//dsn := server_config.GetCfg().DbInfo.Dsn
	//logger.GetSugaredLogger().Debugf("InitDB dsn: %v", dsn)
	//db, err := sqlx.Connect(server_config.GetCfg().DbInfo.DbType, dsn)
	//if err != nil {
	//	errMsg := fmt.Sprintf("failed to connect to MariaDB: %v", err)
	//	logger.GetSugaredLogger().Error(errMsg)
	//	return errors.New(errMsg)
	//}
	//
	//// 연결 풀 설정
	//db.SetMaxOpenConns(12)
	//db.SetMaxIdleConns(12)
	//db.SetConnMaxLifetime(30 * time.Minute)
	//
	//m.db = db
	//
	//logger.GetSugaredLogger().Info("Succeed to initialize MariaDB connection")
	//return m.
	return nil, nil
}
