package relation

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func getSchema(db *gorm.DB, model interface{}) (s *schema.Schema, err error) {
	if model == nil {
		err = errors.New("getSchema, model must not nil")
		return
	}

	// Set new db session
	tempDB := db.Session(&gorm.Session{
		NewDB: true,
	})

	// Get fields of model
	makeStmt := tempDB.Model(model).Take(model)
	if errStmt := makeStmt.Error; errStmt != nil && errStmt != gorm.ErrRecordNotFound {
		if errStmt == gorm.ErrUnsupportedRelation {
			err = fmt.Errorf("getSchema, error on finding model %+v, message: %s. %s", model, errStmt.Error(), "Please make sure you have migrate this model.")
			return
		}
		err = fmt.Errorf("getSchema, error on finding model %+v, message: %s", model, errStmt.Error())
		return
	}
	if makeStmt.Statement == nil || makeStmt.Statement.Schema == nil {
		err = fmt.Errorf("getSchema, statement or schema result of model %+v is nil", model)
		return
	}
	s = makeStmt.Statement.Schema
	return
}
