package main

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"jan-server/services/llm-api/internal/infrastructure/database"
	_ "jan-server/services/llm-api/internal/infrastructure/database/dbschema"
)

var GormGenerator *gen.Generator

func init() {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable"
	}

	// Connect directly without table prefix for schema inspection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	})
	if err != nil {
		panic(err)
	}

	GormGenerator = gen.NewGenerator(gen.Config{
		OutPath:       "./internal/infrastructure/database/gormgen",
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		FieldNullable: true,
	})
	GormGenerator.UseDB(db)
}

func main() {
	for _, model := range database.SchemaRegistry {
		GormGenerator.ApplyBasic(model)
		type Querier interface {
		}
		GormGenerator.ApplyInterface(func(Querier) {}, model)
	}
	GormGenerator.Execute()
}
