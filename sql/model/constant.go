package model

type DB_TYPE string

const (
	MARIADB DB_TYPE = "MARIADB"
)

func (d DB_TYPE) String() string {
	return string(d)
}

type DB_DOMAIN string

const (
	DB_COMMON          DB_DOMAIN = "dp_common"
	DB_COMMPN_PDS_TEST DB_DOMAIN = "db_common_pds_test"
)
