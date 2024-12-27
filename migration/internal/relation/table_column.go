package relation

import (
	"fmt"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
)

type tableColumn map[string][]string

type ITableColumn interface {
	Add(table string, column []string) (err error)
}

func (tc *tableColumn) Add(table string, column []string) (err error) {
	if lib.IsEmptyStr(table) {
		err = fmt.Errorf("table of column %+v is empty", column)
		return
	}

	if len(column) == 0 {
		err = fmt.Errorf("column of table %s is empty", table)
		return
	}

	if existColumn, ok := (*tc)[table]; ok {
		sameLength, sameValues := lib.CompareSliceStr(existColumn, column)
		if !sameValues || !sameLength {
			err = fmt.Errorf("tableColumn of table %s is exists with column: %+v, so we can't add new column: %+v",
				table,
				existColumn,
				column,
			)
			return
		}
		// just return if exist column == new column
		return
	}

	// try adding
	(*tc)[table] = column

	return
}
