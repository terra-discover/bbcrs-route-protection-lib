package relation

import (
	"fmt"

	"gorm.io/gorm"
)

type columnModel map[string]interface{}

type IColumnModel interface {
	Add(column string, model interface{}) (err error)
	ToColumnTable(db *gorm.DB) (ct ColumnTable, err error)
	GetData() (cm map[string]interface{})
}

func NewColumnModel(initial map[string]interface{}) (cm IColumnModel) {
	assignCm := make(columnModel)
	if initial != nil {
		assignCm = columnModel(initial)
	}
	cm = &assignCm
	return
}

func (cm *columnModel) Add(column string, model interface{}) (err error) {
	if model == nil {
		err = fmt.Errorf("model of column %s is nil", column)
		return
	}

	if existModel, ok := (*cm)[column]; ok {
		if existModel != model {
			err = fmt.Errorf("columnModel of column %s is exists with model: %+v, so we can't add new model: %+v",
				column,
				existModel,
				model,
			)
			return
		}
		// just return if exist model == new model
		return
	}

	// try adding
	(*cm)[column] = model

	return
}

func (cm columnModel) ToColumnTable(db *gorm.DB) (ct ColumnTable, err error) {
	// Init value
	ct = ColumnTable{}

	for column, model := range cm {
		assignModel := model
		schema, err := getSchema(db, assignModel)
		if err != nil {
			panic(err.Error())
		}

		table := schema.Table

		if existValue, ok := ct[column]; ok && existValue != table {
			panic(fmt.Sprintf("duplicated value of column %s with different table, %s or %s", column, existValue, table))
		}

		ct[column] = table
	}

	if len(ct) == 0 {
		panic("column table not found")
	}

	return
}

func (cm columnModel) GetData() (data map[string]interface{}) {
	data = cm
	return
}
