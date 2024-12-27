package routeprotection

import (
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	"github.com/terra-discover/bbcrs-route-protection-lib/middleware"
	"github.com/terra-discover/bbcrs-route-protection-lib/migration"
)

type Environment struct {
	BaseUrl string
	AgentID string
	UserID  string
}

type RouteProtection struct {
	Error error

	migration  *migration.Migration
	middleware *middleware.Middleware
	env        Environment
	db         *gorm.DB
}

func NewRouteProtection(env Environment, db *gorm.DB) (rp *RouteProtection) {
	rp = new(RouteProtection)
	rp.setEnvironment(env)
	rp.setDB(db)

	mg := migration.NewMigration(migration.Environment(env), db)
	rp.setMigration(mg)

	md := middleware.NewMiddleware(middleware.Environment(env), db)
	rp.setMiddleware(md)

	return
}

type IRouteProtection interface {
	MigrateRelation(migrationsModel []interface{}) *RouteProtection
	MappingRoute(newRouterSource middleware.RouterSource) *RouteProtection
	ProtectRoute(c *fiber.Ctx) *RouteProtection

	newSession()
	isErrorEmpty() bool
	isMigrated() (err error)
	setEnvironment(newEnv Environment)
	setDB(newDB *gorm.DB)
	setMigration(mg *migration.Migration)
	setMiddleware(md *middleware.Middleware)
	setError(err error)
	clearError()
}

func (rp *RouteProtection) MigrateRelation(removeOldData bool, modelMigrations []interface{}) *RouteProtection {
	rp.newSession()

	err := rp.migration.MigrateRelation(modelMigrations, removeOldData).Error
	rp.setError(err)
	return rp
}

// MappingRoute - will validate and compare all model migrations with listed router
//
// @Params routerFileDir, use to validate listed endpoint by Regex.
// Example:
//
//	app.Get(`/my-endpoint/:id`, myController)
//
// @Params routerPrefix. Can be empty string if router file already describe prefix in every endpoints
// Example:
//
//	"/api/v1/master"
//
// If your router file not describe prefix *statically*, you must fill routerPrefix.
// Example:
//
//	api := app.Group("/api/v1/my-prefix")
//	app.Get(`/my-endpoint/:id`, myController)
//
// In this case, you must fill @Params routerPrefix = "/api/v1/my-prefix"
func (rp *RouteProtection) MappingRoute(newRouterSource middleware.RouterSource, modelMigrations []interface{}, routerFileDir, routerPrefix string) *RouteProtection {
	rp.newSession()

	if err := rp.isMigrated(); err != nil {
		rp.setError(err)
		return rp
	}

	err := rp.middleware.MappingRoute(newRouterSource, modelMigrations, routerFileDir, routerPrefix).Error
	rp.setError(err)
	return rp
}

func (rp *RouteProtection) ProtectRoute(c *fiber.Ctx) *RouteProtection {
	rp.newSession()

	if err := rp.isMigrated(); err != nil {
		rp.setError(err)
		return rp
	}

	err := rp.middleware.ProtectRoute(c).Error
	rp.setError(err)
	return rp
}

func (rp *RouteProtection) newSession() {
	rp.clearError()
}

func (rp *RouteProtection) isErrorEmpty() bool {
	return rp.Error == nil
}

func (rp *RouteProtection) isMigrated() (err error) {
	isMigrated := rp.migration.IsMigrated

	// Logging
	if isMigrated && !lib.IsZeroTime(rp.migration.LastUpdated) {
		log.Printf("INFO RouteProtection: Latest migrate relation at: %s", rp.migration.LastUpdated.String())
	}

	// Set result
	if !isMigrated {
		err = errors.New("WARNING RouteProtection: No relation schema found, please MigrateRelation first")
	}

	return
}

func (rp *RouteProtection) setEnvironment(newEnv Environment) {
	rp.env = newEnv
}

func (rp *RouteProtection) setDB(newDB *gorm.DB) {
	rp.db = newDB
}

func (rp *RouteProtection) setMigration(mg *migration.Migration) {
	rp.migration = mg
}

func (rp *RouteProtection) setMiddleware(md *middleware.Middleware) {
	rp.middleware = md
}

func (rp *RouteProtection) setError(newError error) {
	if rp.isErrorEmpty() {
		rp.Error = newError
	} else if newError != nil {
		rp.Error = fmt.Errorf("%v; %w", rp.Error, newError)
	}
}

func (rp *RouteProtection) clearError() {
	rp.Error = nil
}
