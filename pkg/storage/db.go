package storage

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"strconv"
)

var defaultDb *bun.DB

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewDB(config1 DBConfig) *bun.DB {
	if defaultDb != nil {
		return defaultDb
	}
	//"postgres://postgres:@localhost:5432/test?sslmode=disable"
	// pgx connect by config
	config, err := pgx.ParseConfig("postgres://" + config1.User + ":" + config1.Password + "@" + config1.Host + ":" + strconv.Itoa(config1.Port) + "/" + config1.DBName + "?sslmode=" + config1.SSLMode)
	if err != nil {
		panic(err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	hsqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(hsqldb, pgdialect.New())
	defaultDb = db
	return defaultDb
}
