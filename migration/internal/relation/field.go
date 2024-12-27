package relation

import (
	"strings"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	"gorm.io/gorm/schema"
)

type field schema.Field

type IField interface {
	GetColumnTable(requirement Requirement) (c ColumnTable)
}

func (f field) GetColumnTable(requirement Requirement) (column, table string, isValid bool) {
	c := f.DBName
	isMatchSuffix := strings.HasSuffix(c, requirement.fieldHasSuffix)
	_, isMatchNotIn := lib.FindSlice(requirement.fieldNotIn, c)
	if isMatchSuffix && !isMatchNotIn {
		column = c
		table = strings.TrimSuffix(c, requirement.fieldHasSuffix)
		isValid = true
	}
	return
}
