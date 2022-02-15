package database

// Database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"nathanman/config"
	"nathanman/model"
)

// Connect

var zerotime, _ = time.Parse(time.RFC3339, "0000-00-00T00:00:00Z00:00")

var DbConfig = config.Configuration.Database

var Db *gorm.DB = initiateDB()

func initiateDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		DbConfig.Host, DbConfig.User, DbConfig.Password,
		DbConfig.Name, DbConfig.Port, DbConfig.SslMode,
		config.Configuration.Timezone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Couldn't connect to database")
	}
	return db
}

func Migrate() error {
	err := Db.AutoMigrate(&model.Entry{}, &model.Config{})
	var i int64
	if Db.Model(&model.Config{}).Where("bot = ?", "nathanman").Count(&i); i < 1 {
		Db.Create(&model.Config{
			Lasttime: zerotime,
			Bot:      "nathanman",
		})
	}
	return err
}
