package migration

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	model "github.com/terra-discover/bbcrs-migration-lib/model"
)

// Test__migrateRelation - validate migrate relation
// Skip this unit test is not recommended
func Test__migrateRelation(t *testing.T) {
	db := testSupport__DBConnectAndSeedTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Try migrateRelation
	err := migrateRelation(db, testSupport__GetModelMigrations(), false)
	utils.AssertEqual(t, nil, err, "validate err")

	var listRelationSchema []model.RelationSchema
	res := db.Find(&listRelationSchema)
	utils.AssertEqual(t, nil, res.Error, "validate err")
	utils.AssertEqual(t, true, res.RowsAffected > 0, "validate rows affected")

	// Check are all tables and fields in listRelationSchema valid
	modelMigrations := testSupport__GetModelMigrations()

	mapTableColumns := map[string][]string{}
	for _, model := range modelMigrations {
		assignModel := model
		mod := db.Model(assignModel).Take(assignModel)
		table := mod.Statement.Table
		utils.AssertEqual(t, false, lib.IsEmptyStr(table), "get migrations table")

		columns := mod.Statement.Schema.DBNames
		utils.AssertEqual(t, false, len(columns) == 0, "get migrations columns")

		_, ok := mapTableColumns[table]
		utils.AssertEqual(t, false, ok, "duplicate model")

		mapTableColumns[table] = columns
	}

	// Check all table_source, column_source, used_by_table, and used_by_column from relation_schema, comparing with mapTableColumns from modelMigrations directly
	errCollection := []error{}

	for _, rSchema := range listRelationSchema {
		utils.AssertEqual(t, false, rSchema.TableSource == nil || rSchema.ColumnSource == nil || rSchema.UsedByTable == nil || rSchema.UsedByColumn == nil, "validate pointer")

		// Validate table_source and column_source of relation_schema
		tableSource := *rSchema.TableSource
		columnSource := *rSchema.ColumnSource

		isTableSourceMatch := false

		for mTable, mColumns := range mapTableColumns {
			if mTable == tableSource {
				isTableSourceMatch = true
				_, isFound := lib.FindSlice(mColumns, columnSource)
				if !isFound {
					errCollection = append(errCollection, fmt.Errorf("column_source %s is not found in table_source %s", columnSource, tableSource))
				}
			}
		}

		if !isTableSourceMatch {
			errCollection = append(errCollection, fmt.Errorf("table_source %s is not found in any models (detail: used_by_column: %s , used_by_table: %s)", tableSource, *rSchema.UsedByColumn, *rSchema.UsedByTable))
		}

		// Validate used_by_table and used_by_column of relation_schema
		usedByTable := *rSchema.UsedByTable
		usedByColumn := *rSchema.UsedByColumn

		isUsedByTableMatch := false

		for mTable, mColumns := range mapTableColumns {
			if mTable == usedByTable {
				isUsedByTableMatch = true
				_, isFound := lib.FindSlice(mColumns, usedByColumn)
				if !isFound {
					errCollection = append(errCollection, fmt.Errorf("used_by_column %s is not found in used_by_table %s", usedByColumn, usedByTable))
				}
			}
		}

		if !isUsedByTableMatch {
			errCollection = append(errCollection, fmt.Errorf("used_by_table %s is not found in any models (detail: table_source: %s)", usedByTable, tableSource))
		}
	}

	showErr := ""
	for _, err := range errCollection {
		showErr += "- " + err.Error() + "\n"
	}
	utils.AssertEqual(t, false, !lib.IsEmptyStr(showErr), "Err desc: \n"+showErr)
}

func Test_mustDeleteOldData(t *testing.T) {
	db := testSupport__DBConnectAndSeedTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Case 1: (Success) no data found to delete
	err := mustDeleteOldData(db)
	utils.AssertEqual(t, nil, err, "response err")

	// Case 2: (Success) data found and (Not found) check after deleted
	relationSchema := model.RelationSchema{}
	relationSchema.TableSource = lib.Strptr("country")
	relationSchema.ColumnSource = lib.Strptr("id")
	relationSchema.UsedByTable = lib.Strptr("city")
	relationSchema.UsedByColumn = lib.Strptr("country_id")
	err = db.Create(&relationSchema).Error
	utils.AssertEqual(t, nil, err, "mock data")

	err = mustDeleteOldData(db)
	utils.AssertEqual(t, nil, err, "response err")

	err = db.First(&relationSchema).Error
	utils.AssertEqual(t, true, err != nil, "check exist data")
}

func Test_getLatestMigrateRelation(t *testing.T) {
	db := testSupport__DBConnectAndSeedTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Case 1: (Not Found) Data still empty
	lastUpdated, isFound, err := getLatestMigrateRelation(db)
	utils.AssertEqual(t, false, err != nil, "check err")
	utils.AssertEqual(t, false, isFound, "check is found")
	utils.AssertEqual(t, true, lib.IsZeroTime(lastUpdated), "check last updated")

	// Case 2: (Success) Data created and (Success) Found
	relationSchema := model.RelationSchema{}
	relationSchema.TableSource = lib.Strptr("country")
	relationSchema.ColumnSource = lib.Strptr("id")
	relationSchema.UsedByTable = lib.Strptr("city")
	relationSchema.UsedByColumn = lib.Strptr("country_id")
	err = db.Create(&relationSchema).Error
	utils.AssertEqual(t, nil, err, "mock data")

	lastUpdated, isFound, err = getLatestMigrateRelation(db)
	utils.AssertEqual(t, false, err != nil, "check err")
	utils.AssertEqual(t, true, isFound, "check is found")
	utils.AssertEqual(t, false, lib.IsZeroTime(lastUpdated), "check last updated")

	// Case 3: (Error) DB Closed
	sqlDB, _ := db.DB()
	sqlDB.Close()

	lastUpdated, isFound, err = getLatestMigrateRelation(db)
	utils.AssertEqual(t, true, err != nil, "check err")
	utils.AssertEqual(t, false, isFound, "check is found")
	utils.AssertEqual(t, true, lib.IsZeroTime(lastUpdated), "check last updated")
}

func Test_genTableRelationSchema(t *testing.T) {
	db := testSupport__DBConnectTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Case 1: Table not created (NOT ERROR)
	err := db.Migrator().DropTable(&model.RelationSchema{})
	utils.AssertEqual(t, false, err != nil, "check err")

	err = genTableRelationSchema(db)
	utils.AssertEqual(t, false, err != nil, "check err")

	// Case 2: Table has created (NOT ERROR)
	err = genTableRelationSchema(db)
	utils.AssertEqual(t, false, err != nil, "check err")

	// Case 3: DB connection close (ERROR)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	err = genTableRelationSchema(db)
	utils.AssertEqual(t, true, err != nil, "check err")
}
