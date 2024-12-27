package middleware

import (
	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"

	"github.com/terra-discover/bbcrs-route-protection-lib/model"

	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
)

type routerMap map[string]string
type routerMaps map[string]routerMap

func (r *routerMaps) addMap(k string, v routerMap) {
	// assign value
	if existValue, ok := (*r)[k]; !ok || existValue == nil {
		(*r)[k] = v
		return
	} else {
		for rk, rv := range v {
			if rOldValue, ok := existValue[rk]; !ok || lib.IsEmptyStr(rOldValue) {
				existValue[rk] = rv
			}
		}
		(*r)[k] = existValue
		return
	}
}

// SourceRelation
// Note: If RequiredRelation and IgnoreRelation are declared, data protection ONLY validate RequiredRelation
type SourceRelation struct {
	Source           string   // ex: country
	RequiredRelation []string // ex: city
	IgnoreRelation   []string // ex: city_translation
}
type RouterSource map[string]SourceRelation

func (rs *RouterSource) toRouterMaps(listRelationSchema []model.RelationSchema) (deleteRouteMaps routerMaps, err error) {
	deleteRouteMaps = make(routerMaps)

	for routePattern, sourceRelation := range *rs {
	loopRelationSchema:
		for _, relationSchema := range listRelationSchema {
			// declare relation schema
			schemaSource := relationSchema.TableSource
			schemaUsedByTable := relationSchema.UsedByTable
			schemaUsedByColumn := relationSchema.UsedByColumn

			// validate relation schema
			if schemaSource == nil {
				log.Println("ERROR generateDeleteRouteMaps: one of relation schema.table source is nil")
				err = errors.New("one of relation schema.table source is nil")
				return
			} else if schemaUsedByTable == nil {
				log.Println("ERROR generateDeleteRouteMaps: one of relation schema.used_by_table is nil")
				err = errors.New("one of relation schema.used_by_table is nil")
				return
			} else if schemaUsedByColumn == nil {
				log.Println("ERROR generateDeleteRouteMaps: one of relation schema.used_by_column is nil")
				err = errors.New("one of relation schema.used_by_column is nil")
				return
			}

			// validate source
			if sourceRelation.Source != *schemaSource {
				continue loopRelationSchema
			}

			// validate required relation
			if len(sourceRelation.RequiredRelation) > 0 {
			loopRequired:
				for _, requiredRelation := range sourceRelation.RequiredRelation {
					if requiredRelation == *schemaUsedByTable {
						// add to routerMaps
						newRouterMap := make(routerMap)
						newRouterMap[*relationSchema.UsedByTable] = *schemaUsedByColumn
						deleteRouteMaps.addMap(routePattern, newRouterMap)
						break loopRequired
					}
				}

				continue loopRelationSchema
			}

			// validate ignore relation
			mustIgnore := false
		loopIgnore:
			for _, ignoreRelation := range sourceRelation.IgnoreRelation {
				if ignoreRelation == *schemaUsedByTable {
					mustIgnore = true
					break loopIgnore
				}
			}
			if mustIgnore {
				continue loopRelationSchema
			}

			// add to routerMaps
			newRouterMap := make(routerMap)
			newRouterMap[*relationSchema.UsedByTable] = *schemaUsedByColumn
			deleteRouteMaps.addMap(routePattern, newRouterMap)
		}
	}

	return
}

// deleteRouteMaps
// Format:
//
//	"regex_pattern": {
//	    "table_name": "id_field_name",
//	    "table_name": "id_field_name",
//	}
//
// Notes: Regex pattern must capture a group as an id
// To reduce the number of lines, please separate the format for each model
// example:
// var deleteRouteMaps = routeMaps{model.CorporateDeleteRouterMap, model.AgentDeleteRouterMap}
// var deleteRouteMaps = routerMaps{
// 	".*/corporates?/([^/]+)$": routerMap{
// 		"agent_corporate":    "corporate_id",
// 		"corporate_employee": "corporate_id",
// 	},
// }

// getListRelationSchema - find list relation schema from database
func getListRelationSchema(db *gorm.DB) (listRelationSchema []model.RelationSchema, err error) {
	if errFind := db.Find(&listRelationSchema).Error; errFind != nil {
		log.Println("ERROR getListRelationSchema: ", errFind)
		err = errors.New("failed to find relation: " + errFind.Error())
		return
	}
	return
}

// generateDeleteRouteMaps - map list relation schema to router maps
func generateDeleteRouteMaps(db *gorm.DB) (deleteRouteMaps routerMaps, err error) {
	listRelationSchema, errGet := getListRelationSchema(db)
	if errGet != nil {
		err = errGet
		return
	}

	listRouterSource := getDeleteRouterSource()

	deleteRouteMaps, err = listRouterSource.toRouterMaps(listRelationSchema)
	if err != nil {
		return
	}

	return
}

// updateRouteMaps
// Format:
//
//	"regex_pattern": {
//	    "table_name": "id_field_name",
//	    "table_name": "id_field_name",
//	}
var updateRouteMaps = routerMaps{}

// generateDataProtectionQuery - count data based on table criteria
func generateDataProtectionQuery(tables routerMap, ids []uuid.UUID) string {
	queries := []string{}
	// Argument indexes (simplify repeatable arguments)
	// Source: https://faun.pub/golangs-fmt-sprintf-and-printf-demystified-4adf6f9722a2
	queryTemplate := `
	SELECT COUNT(*) total FROM "%[1]s" 
	WHERE "%[1]s"."%[2]s" IN(%[3]s) AND "%[1]s"."deleted_at" IS NULL`
	for tableName := range tables {
		fieldName := tables[tableName]
		strIds := lib.ConvertSliceUUIDToStr(ids, ",", `'%s'`)
		queries = append(queries,
			fmt.Sprintf(queryTemplate,
				tableName, fieldName, strIds))
	}

	output := ""
	if len(queries) > 0 {
		output = fmt.Sprintf(
			`SELECT SUM("s"."total") "total" FROM (
				%s
			) "s"`, strings.Join(queries, "\nUNION\n"))
	}

	return output
}

func isDeleteMethod(method string) bool {
	return method == "DELETE"
}

func isDeleteBatchAction(c fiber.Ctx) (moduleName string, ids []uuid.UUID, isDeleteBatchAction bool, err error) {
	// Validate method
	method := c.Method()
	if method != "POST" {
		return
	}

	// Validate path
	var batchActionRoutePattern = ".*/batch-actions?/([^/]+\\S)/([^/]+\\S)$"

	pattern, err := regexp.Compile(batchActionRoutePattern)
	if err != nil {
		err = fmt.Errorf("isDeleteBatchAction: %s", err.Error())
		return
	}

	route := c.Path()

	if isMatchPattern := pattern.MatchString(route); !isMatchPattern {
		return
	}

	matches := pattern.FindStringSubmatch(route)
	if len(matches) < 2 {
		return
	}

	if matches[1] != "delete" {
		return
	}

	ids = *new([]uuid.UUID)
	if errParse := c.BodyParser(&ids); err != nil {
		err = fmt.Errorf("isDeleteBatchAction: %s", errParse.Error())
		return
	}

	isDeleteBatchAction = true
	moduleName = matches[2]

	return
}

// matchingRouteToTables - map route with table sources
func matchingRouteToTables(db *gorm.DB, route, method string) (*routerMap, *uuid.UUID, error) {
	isDeleteMethod := isDeleteMethod(method)

	var r *routerMap
	var id *uuid.UUID
	rMaps := updateRouteMaps
	if isDeleteMethod {
		maps, errMaps := generateDeleteRouteMaps(db)
		if errMaps != nil {
			return nil, nil, errMaps
		}

		rMaps = maps
	}
	for i := range rMaps {
		pattern, err := regexp.Compile(i)
		if nil == err && pattern.MatchString(route) {
			result := pattern.FindStringSubmatch(route)
			if len(result) > 0 {
				idParam, err := uuid.Parse(result[1])
				if nil == err && idParam != uuid.Nil {
					rMap := rMaps[i]
					id = &idParam
					r = &rMap
				}
			}
		}
	}

	return r, id, nil
}

func getBatchActionModuleName(db *gorm.DB, name string) (string, error) {
	regex := regexp.MustCompile(`[^A-z0-9\_]+`)
	moduleName, _ := url.PathUnescape(name)
	moduleName = regex.ReplaceAllString(moduleName, "_")
	moduleName = strcase.ToSnake(moduleName)
	moduleNameLength := len(moduleName)
	if moduleNameLength > 0 && strings.ReplaceAll(moduleName, "_", "") != "" {
		aiueo := regexp.MustCompile(`[^aiueo](ie|e)?s$`)
		originalName := moduleName
		if aiueo.MatchString(moduleName) {
			if moduleNameLength > 4 && moduleName[moduleNameLength-3:] == "ies" {
				moduleName = moduleName[:moduleNameLength-3] + "y"
			} else {
				moduleName = moduleName[:moduleNameLength-1]
			}
		}

		var total int64
		result := db.Table(moduleName).Count(&total)
		if nil != result.Error && total == 0 {
			result2 := db.Table(originalName).Count(&total)
			if nil != result2.Error && total == 0 {
				return "", fmt.Errorf("module %s is not found", moduleName)
			}
			moduleName = originalName
		}

		return moduleName, nil
	}

	return "", errors.New("module not found")

}

func matchBatchActionRouteTable(db *gorm.DB, moduleName string) (r *routerMap, err error) {
	var deleteRouterSourcePattern = ".*/([a-z-]+)\\S+"

	formatModuleName, errFormat := getBatchActionModuleName(db, moduleName)
	if errFormat != nil {
		err = fmt.Errorf("matchBatchActionRouteTable: route module is invalid, %s", err.Error())
		return
	}

	rMaps, errGen := generateDeleteRouteMaps(db)
	if errGen != nil {
		err = errGen
		return
	}

	for k, v := range rMaps {
		pattern, err := regexp.Compile(deleteRouterSourcePattern)
		if nil == err && pattern.MatchString(k) {
			result := pattern.FindStringSubmatch(k)
			if len(result) > 0 {
				routerModule := (result[1])
				formatRouterModule, err := getBatchActionModuleName(db, routerModule)

				if nil == err && formatRouterModule == formatModuleName {
					// assign value
					rMap := v
					r = &rMap
					break
				}
			}
		}
	}

	return
}

func validateProtectionQuery(db *gorm.DB, rmap routerMap, ids []uuid.UUID) (isAllowed bool) {
	query := generateDataProtectionQuery(rmap, ids)
	result := struct {
		Total int64
	}{}
	if query != "" {
		raw := db.Raw(query).Scan(&result)
		if raw.RowsAffected > 0 && result.Total > 0 {
			isAllowed = false
			return
		}
	}

	isAllowed = true
	return
}

// runDataProtection - middleware for protect specific data, map by router
func runDataProtection(c *fiber.Ctx, db *gorm.DB) error {
	errResp := initValidation()
	if !errResp.IsEmpty() {
		return errResp.SendToContext(c)
	}

	var (
		dataMap *routerMap
		dataIds *[]uuid.UUID
	)

	moduleName, ids, isDeleteBatchAction, err := isDeleteBatchAction(*c)
	if err != nil {
		log.Println("ERROR failed validate request batch 1:", err.Error())
		return lib.ErrorInternal(c, "failed validate request batch 1")
	}

	if isDeleteBatchAction {
		rmap, err := matchBatchActionRouteTable(db, moduleName)
		if err != nil {
			log.Println("ERROR failed validate request batch 2:", err.Error())
			return lib.ErrorBadRequest(c, "failed validate request batch 2, invalid path module name")
		}

		if nil != rmap && len(ids) > 0 {
			dataMap = rmap
			dataIds = &ids
		}

	} else {

		isDeleteMethod := isDeleteMethod(c.Method())
		if isDeleteMethod { // || c.Method() == "UPDATE" {
			if rmap, id, err := matchingRouteToTables(db, c.Path(), c.Method()); err != nil {
				log.Println("ERROR failed validate request delete:", err.Error())
				return lib.ErrorInternal(c, "failed validate request delete")

			} else if nil != rmap && !lib.IsEmptyUUIDPtr(id) {
				dataMap = rmap
				dataIds = &[]uuid.UUID{*id}
			}
		}
	}

	if dataMap != nil && dataIds != nil {
		if isAllowed := validateProtectionQuery(db, *dataMap, *dataIds); !isAllowed {
			return lib.ErrorNotAllowed(c,
				"Sorry, you are not allowed to delete this data. It is already used in transactions.")
		}
	}

	return c.Next()
}

var isInit bool = true

func initValidation() (errResp lib.ErrorResponse) {
	// Only validate one time
	if !isInit {
		return
	}

	log.Println("START initValidation Data Protection")
	defer log.Println("END initValidation Data Protection")

	// check endpoint
	if endpoint := masterServiceEndpoint; lib.IsEmptyStr(endpoint) {
		errResp = lib.SetErrorInternal("all routes cannot be access on master data. Please contact our support")
		return
	}

	// Set isInit to false
	isInit = false
	return
}
