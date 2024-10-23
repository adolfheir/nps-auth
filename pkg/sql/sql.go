package sql

import (
	"nps-auth/configs"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	once sync.Once
	db   *gorm.DB
)

func initDataBase() {
	conf := configs.GetConfig()

	log := newLogger()

	sqlite, err := gorm.Open(sqlite.Open(conf.DB.DSN), &gorm.Config{
		Logger: newLogger(),
	})

	if err != nil {
		log.logger.Panic().Err(err).Msg("failed to connect database")
	}
	db = sqlite

	if err := db.AutoMigrate(&Channel{}); err != nil {
		log.logger.Panic().Err(err).Msg("failed to migrate database")
	}
}

func GetDB() *gorm.DB {

	once.Do(func() {
		initDataBase()
	})

	return db
}
