package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Environment struct {
	BaseUrl string
	AgentID string
	UserID  string
}

// Implementing method chaining
type Middleware struct {
	Error error

	env Environment
	db  *gorm.DB
}

func NewMiddleware(env Environment, db *gorm.DB) (m *Middleware) {
	m = new(Middleware)
	m.setEnvironment(env)
	m.setDB(db)
	return
}

type IMiddleware interface {
	MappingRoute(newRouterSource RouterSource) *Middleware
	ProtectRoute(c *fiber.Ctx) *Middleware

	newSession()
	isRouterSourceEmpty() bool
	isErrorEmpty() bool
	setEnvironment(newEnv Environment)
	setDB(newDB *gorm.DB)
	setError(err error)
	clearError()
}

func (m *Middleware) MappingRoute(newRouterSource RouterSource, modelMigrations []interface{}, routerFileDir, routerPrefix string) *Middleware {
	m.newSession()

	setDeleteRouterSource(newRouterSource)
	err := validateRouterSource(m.db, modelMigrations, routerFileDir, routerPrefix)
	m.setError(err)
	return m
}

func (m *Middleware) ProtectRoute(c *fiber.Ctx) *Middleware {
	m.newSession()

	if m.isRouterSourceEmpty() {
		m.setError(errors.New("router source is not found. Please Mapping Route first"))
		return m
	}

	err := runDataProtection(c, m.db)
	m.setError(err)
	return m
}

func (m *Middleware) newSession() {
	m.clearError()
}

func (m *Middleware) isRouterSourceEmpty() bool {
	mapRouterSource := getDeleteRouterSource()
	return len(mapRouterSource) == 0
}

func (m *Middleware) isErrorEmpty() bool {
	return m.Error == nil
}

func (m *Middleware) setEnvironment(newEnv Environment) {
	m.env = newEnv
}

func (m *Middleware) setDB(newDB *gorm.DB) {
	m.db = newDB
}

func (m *Middleware) setError(newError error) {
	if m.isErrorEmpty() {
		m.Error = newError
	} else if newError != nil {
		m.Error = fmt.Errorf("%v; %w", m.Error, newError)
	}
}

func (m *Middleware) clearError() {
	m.Error = nil
}
