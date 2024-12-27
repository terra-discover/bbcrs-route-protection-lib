package relation

import (
	"fmt"

	"gorm.io/gorm"
)

type modelColumn map[interface{}][]string

type IModelColumn interface {
	ToTableColumn(db *gorm.DB) (tc tableColumn, err error)
}

func (mc modelColumn) ToTableColumn(db *gorm.DB) (tc tableColumn, err error) {
	for model, listColumn := range mc {
		// Get fields of model
		assignModel := model
		schema, err := getSchema(db, assignModel)
		if err != nil {
			panic(err.Error())
		}

		table := schema.Table

		fields := schema.Fields

		for _, column := range listColumn {
			isMatch := false

		loopFields:
			for _, f := range fields {
				assignField := *f
				if assignField.DBName == column {
					isMatch = true
					break loopFields
				}
			}

			if !isMatch {
				panic(fmt.Sprintf("column %s is not listed on table %s", column, table))
			}
		}

		// Add to table column
		if err := tc.Add(table, listColumn); err != nil {
			panic(err.Error())
		}
	}

	if len(tc) == 0 {
		panic("table column not found")
	}

	return
}
