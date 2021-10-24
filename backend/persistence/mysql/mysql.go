package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

const (
	ErrorCodeDuplicateEntry = 1062
)

func MustGetDB(setting Setting) *sql.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		setting.User,
		setting.Password,
		setting.Host,
		setting.Port,
		setting.Name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Panic(err)
	}
	if err := db.Ping(); err != nil {
		logrus.Panic(err)
	}

	db.SetMaxIdleConns(setting.MaxIdleConns)
	db.SetMaxOpenConns(setting.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(setting.ConnMaxLifetimeMs) * time.Millisecond)

	return db
}
