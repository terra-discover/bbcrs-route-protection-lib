package migration

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Environment struct {
	BaseUrl string
	AgentID string
	UserID  string
}

// Implementing method chaining
type Migration struct {
	LastUpdated time.Time
	IsMigrated  bool
	Error       error

	env Environment
	db  *gorm.DB
}

func NewMigration(env Environment, db *gorm.DB) (m *Migration) {
	m = new(Migration)
	m.setEnvironment(env)
	m.setDB(db)
	m.checkIsMigrated()
	return
}

type IMigration interface {
	MigrateRelation(migrationsModel []interface{}, removeOldData bool) *Migration

	newSession()
	isErrorEmpty() bool
	checkIsMigrated() *Migration
	setEnvironment(newEnv Environment)
	setDB(newDB *gorm.DB)
	setError(err error)
	clearError()
}

func (m *Migration) MigrateRelation(migrationsModel []interface{}, removeOldData bool) *Migration {
	m.newSession()

	err := migrateRelation(m.db, migrationsModel, removeOldData)
	m.checkIsMigrated()
	m.setError(err)
	return m
}

func (m *Migration) newSession() {
	m.clearError()
}

func (m *Migration) isErrorEmpty() bool {
	return m.Error == nil
}

func (m *Migration) checkIsMigrated() {
	lastUpdated, isFound, err := getLatestMigrateRelation(m.db)
	m.LastUpdated = lastUpdated
	m.IsMigrated = isFound
	m.setError(err)
}

func (m *Migration) setEnvironment(newEnv Environment) {
	m.env = newEnv
}

func (m *Migration) setDB(newDB *gorm.DB) {
	m.db = newDB
}

func (m *Migration) setError(newError error) {
	if m.isErrorEmpty() {
		m.Error = newError
	} else if newError != nil {
		m.Error = fmt.Errorf("%v; %w", m.Error, newError)
	}
}

func (m *Migration) clearError() {
	m.Error = nil
}
