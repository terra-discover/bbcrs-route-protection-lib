package relation

import (
	"fmt"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
)

type ColumnTable map[string]string

type IColumnTable interface {
	Add(column, table string) (err error)
	Table(column string) (table string, isFound bool)
	GetListTable() []string
	GetData() (data map[string]string)
}

func NewColumnTable(initial map[string]string) (ct IColumnTable) {
	assignCt := make(ColumnTable)
	if initial != nil {
		assignCt = initial
	}
	ct = &assignCt
	return
}

func (ct *ColumnTable) Add(column, table string) (err error) {
	if lib.IsEmptyStr(column) {
		err = fmt.Errorf("column of table %s is empty", table)
	}
	if lib.IsEmptyStr(table) {
		err = fmt.Errorf("table of column %s is empty", column)
	}

	if existTable, ok := (*ct)[column]; ok {
		if existTable != table {
			err = fmt.Errorf("columnModel of column %s is exists with table: %+v, so we can't add new table: %+v",
				column,
				existTable,
				table,
			)
			return
		}
		// just return if exist model == new model
		return
	}

	// try adding
	(*ct)[column] = table

	return
}

func (ct ColumnTable) Table(column string) (table string, isFound bool) {
	if existTable, ok := ct[column]; ok {
		table = existTable
		isFound = true
	}
	return
}

// get list table only and make it unique
func (ct ColumnTable) GetListTable() []string {
	listTable := []string{}

	for _, dupTable := range ct {
		isMatch := false

		for _, uqTable := range listTable {
			if uqTable == dupTable {
				isMatch = true
			}
		}

		if !isMatch {
			listTable = append(listTable, dupTable)
		}
	}

	return listTable
}

func (ct ColumnTable) GetData() (data map[string]string) {
	data = ct
	return
}

type ColumnTables map[string][]string
