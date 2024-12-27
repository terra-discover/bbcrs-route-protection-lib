package relation

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type MigrationModel struct {
	db                *gorm.DB
	listModelPointers []interface{}
	requirement       Requirement
	err               error
}

type IMigrationModel interface {
	ToColumnTable() (columnSource ColumnTable, columnUsedBy ColumnTables, err error)
	setDB(db *gorm.DB)
	setListModelPointers(listModelPointers []interface{})
	setRequirement(requirement Requirement)
	setErr()
}

func NewMigrationModel(db *gorm.DB, listModelPointers []interface{}, requirement Requirement) (m IMigrationModel) {
	m = new(MigrationModel)
	m.setDB(db)
	m.setListModelPointers(listModelPointers)
	m.setRequirement(requirement)
	m.setErr()
	return
}

func (m *MigrationModel) setDB(db *gorm.DB) {
	m.db = db
}

func (m *MigrationModel) setListModelPointers(listModelPointers []interface{}) {
	m.listModelPointers = listModelPointers
}

func (m *MigrationModel) setRequirement(requirement Requirement) {
	m.requirement = requirement
}

func (m *MigrationModel) setErr() {
switchCond:
	switch {
	case m.db == nil:
		m.err = errors.New("db is nil")
		break switchCond
	case len(m.listModelPointers) == 0:
		m.err = errors.New("length listModelPointers = 0")
		break switchCond
	case len(m.listModelPointers) > 0:
		{
		loopModelPointers:
			for _, modelPointer := range m.listModelPointers {
				if modelPointer == nil {
					m.err = errors.New("cannot address model of listModelPointers")
					break loopModelPointers
				}
			}
			break switchCond
		}
	case !m.requirement.isFulfilled():
		m.err = errors.New("requirement is not fulfilled")
		break switchCond
	default:
		m.err = nil
		break switchCond
	}
}

func (m MigrationModel) ToColumnTable() (columnSource ColumnTable, columnUsedBy ColumnTables, err error) {
	// Validate migration model
	if m.err != nil {
		err = m.err
		return
	}

	// Init value
	columnSource = ColumnTable{}
	columnUsedBy = ColumnTables{}

	arrMessage := []string{}

loopModelPointers:
	for _, model := range m.listModelPointers {
		// Get fields of model
		assignModel := model
		schema, errSchema := getSchema(m.db, assignModel)
		if errSchema != nil {
			arrMessage = append(arrMessage, errSchema.Error())
			continue loopModelPointers
		}
		fields := schema.Fields
		currentTable := schema.Table

		// Get column table match requirement
	loopFields:
		for _, f := range fields {
			assignField := *f
			newField := field(assignField)
			column, table, isValid := newField.GetColumnTable(m.requirement)
			if !isValid {
				continue loopFields
			}

			// Add to column source
			columnSource[column] = table

			// Add to column used by
			if existValue, ok := columnUsedBy[column]; !ok {
				columnUsedBy[column] = []string{currentTable}
			} else {
				existValue = append(existValue, currentTable)
				columnUsedBy[column] = existValue
			}
		}
	}

	if len(arrMessage) > 0 {
		message := strings.Join(arrMessage, "\n -")
		err = fmt.Errorf("ERROR MigrationModel.ToColumnTable: \n %s", message)
		return
	} else if len(columnSource) == 0 {
		err = errors.New("ERROR MigrationModel.ToColumnTable: column table not found, make sure model migrations is listed (beside of table relation and relation_schema)")
		return
	}

	return
}
