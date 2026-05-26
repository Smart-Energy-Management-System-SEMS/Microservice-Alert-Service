package configuration

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}
