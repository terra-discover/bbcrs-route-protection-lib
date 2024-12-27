package relation

import "github.com/terra-discover/bbcrs-helper-lib/pkg/lib"

type Requirement struct {
	fieldHasSuffix string
	fieldNotIn     []string
}

func SetRequirement(fieldHasSuffix string, fieldNotIn []string) (r Requirement) {
	r = Requirement{}
	r.setFieldHasSuffix(fieldHasSuffix)
	r.setFieldNotIn(fieldNotIn)
	return
}

func (r *Requirement) setFieldHasSuffix(fieldHasSuffix string) {
	r.fieldHasSuffix = fieldHasSuffix
}

func (r *Requirement) setFieldNotIn(fieldNotIn []string) {
	r.fieldNotIn = fieldNotIn
}

func (r *Requirement) isFulfilled() bool {
	return lib.IsEmptyStr(r.fieldHasSuffix)
}
