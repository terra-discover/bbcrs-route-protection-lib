package migration

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	migration "github.com/terra-discover/bbcrs-migration-lib"
	seed "github.com/terra-discover/bbcrs-migration-lib/seed"
)

// testSupport__DBConnectTest test
func testSupport__DBConnectTest(database ...string) *gorm.DB {
	dbPath := "file::memory:"
	if len(database) > 0 {
		dbPath = database[0]
		if dbPath == "" {
			dbPath = "file::memory:"
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		QueryFields:                              false,
	})
	if nil != err {
		panic(err)
	}

	err = db.AutoMigrate(testSupport__GetModelMigrations()...)
	if nil != err {
		panic(err)
	}

	return db
}

// testSupport__DBConnectAndSeedTest connect and seed
func testSupport__DBConnectAndSeedTest(database ...string) *gorm.DB {
	db := testSupport__DBConnectTest(database...)
	testSupport__DBSeedTest(db)
	return db
}

// testSupport__DBSeedTest seed sample data
func testSupport__DBSeedTest(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		seeds := testSupport__DataSeeds()
		for i := range seeds {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(seeds[i]).Error; nil != err {
				return err
			}
		}

		return nil
	})
}

// testSupport__GetModelMigrations models to migrate
func testSupport__GetModelMigrations() []interface{} {
	return migration.ModelMigrations
}

// testSupport__DataSeeds data to seeds using random request
func testSupport__DataSeeds() []interface{} {
	req := seed.DataSeedsRequest{
		AgentID:        *lib.GenUUID(),
		UserID:         *lib.GenUUID(),
		SmtpEmail:      "abc@test.co",
		SmtpSenderName: "Bayu Buana",
		Salt:           lib.RandomChars(6),
		Aes:            lib.RandomChars(6),
	}
	return seed.DataSeeds(req)
}
