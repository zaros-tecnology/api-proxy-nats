package migration

import (
	"fmt"

	"github.com/zaros-tecnology/api-proxy-nats/internal/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"gorm.io/gorm"
)

// Migration interface
type Migration interface {
	Magration(db *gorm.DB, nc *request.NatsConn)
	Migrate(db *gorm.DB, nc *request.NatsConn, name string, versions []Versions) error
	Versions() []Versions
}

// Versions version migrate
type Versions interface {
	Save(tx *gorm.DB, nc *request.NatsConn, name string, version int) error
	Migrate(db *gorm.DB) error
}

// Base migration
type Base struct{}

// NewBase new instance migration
func NewBase() *Base {
	return &Base{}
}

// Migrate run migration
func (m *Base) Migrate(db *gorm.DB, nc *request.NatsConn, name string, versions []Versions) error {
	fmt.Println(">> AutoMigrate", name)
	defer fmt.Println("<< AutoMigrate", name)
	err := db.AutoMigrate(&models.SchemaVersion{}, &models.Service{})
	if err != nil {
		panic(err)
	}

	var schema models.SchemaVersion
	db.Where(&models.SchemaVersion{Service: name}).Order("version DESC").First(&schema)

	for index, item := range versions {
		if schema.Version < index+1 {
			err = item.Migrate(db)
			if err != nil {
				return err
			}
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for index, item := range versions {
			if schema.Version < index+1 {
				err = item.Save(tx, nc, name, index+1)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// BaseVersion base migration version
type BaseVersion struct{}

// Save base
func (b *BaseVersion) Save(tx *gorm.DB, nc *request.NatsConn, name string, version int) error {
	var schema models.SchemaVersion
	schema.Service = name
	schema.Version = version
	return tx.Create(&schema).Error
}
