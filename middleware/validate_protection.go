package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	"gorm.io/gorm"
)

// validateRouterSource - check duplicate route, validate route, and validate table listed on deleteRouterSource
// Note: Only can compare with components inside this service
func validateRouterSource(db *gorm.DB, modelMigrations []interface{}, routerFileDir, routerPrefix string) (err error) {
	// Set method to check
	methodCheck := DeleteMethod

	// Get router source
	routerSource := getDeleteRouterSource()

	// 1. check valid route
	err = checkValidRoute(methodCheck, routerSource, routerFileDir, routerPrefix)
	if err != nil {
		return
	}

	// 2. check valid table
	err = checkValidTable(db, routerSource, modelMigrations)
	if err != nil {
		return
	}

	return
}

type MappingRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func checkValidRoute(methodCheck method, routerSource RouterSource, routerFileDir, routerPrefix string) (err error) {

	listRoute, errMapping := mappingRoute(methodCheck, routerFileDir, routerPrefix)
	if errMapping != nil {
		err = errMapping
		return
	}

	errValidate := validateRoute(routerSource, listRoute)
	if errValidate != nil {
		err = errValidate
		return
	}

	return
}

type method string

const (
	_            method = ""
	PostMethod   method = "POST"
	PutMethod    method = "PUT"
	DeleteMethod method = "DELETE"
	GetMethod    method = "GET"
)

func (m method) String() string {
	return string(m)
}

type fileExtension string

const (
	_      fileExtension = ""
	Golang fileExtension = ".go"
	Text   fileExtension = ".txt"
	Json   fileExtension = ".json"
)

func (f fileExtension) String() string {
	return string(f)
}

func listFileExtension() []string {
	return []string{
		Golang.String(),
		Text.String(),
		Json.String(),
	}
}

func mappingRoute(methodCheck method, routerFileDir, routerPrefix string) (listRoute []MappingRoute, err error) {
	// Find path from router file
	bte, errRead := os.ReadFile(routerFileDir)
	if errRead != nil {
		err = fmt.Errorf("mappingRoute open router file: %s", errRead.Error())
		return
	}

	// Switch extension file
switchExt:
	switch fileExt := filepath.Ext(routerFileDir); fileExt {
	case Golang.String(), Text.String():
		{
			listRoute = mappingRouteRawCodeFile(bte, methodCheck)
			break switchExt
		}
	case Json.String():
		{
			listRoute, err = mappingRouteJsonFile(bte, methodCheck)
			if err != nil {
				return
			}
			break switchExt
		}
	default:
		{
			strSupportedExt := strings.Join(listFileExtension(), " | ")
			suffixMessage := fmt.Sprintf("Please use one of this extensions: %s", strSupportedExt)
			if lib.IsEmptyStr(fileExt) {
				err = fmt.Errorf("mappingRoute switch file extension: file extension %s of path %s is not supported. %s", fileExt, routerFileDir, suffixMessage)
			} else {
				err = fmt.Errorf("mappingRoute switch file extension: please declare your file extension of path %s. %s", routerFileDir, suffixMessage)
			}
			break switchExt
		}
	}

	// Append routerPrefix
	if !lib.IsEmptyStr(routerPrefix) {
		for i := 0; i < len(listRoute); i++ {
			existPath := listRoute[i].Path
			listRoute[i].Path = routerPrefix + existPath
		}
	}

	return
}

/*
mappingRouteRawCodeFile
Format:

	app.Get("/my-endpoint", controller.GetList.go)
	app.Delete("/my-endpoint/:id", controller.DeleteByID.go)
*/
func mappingRouteRawCodeFile(bte []byte, methodCheck method) (listRoute []MappingRoute) {
	re := regexp.MustCompile("[^\\. \\t\\s]+\\.(Get|Put|Post|Delete|Patch)\\([`\"]([^`\"]+)[`\"],[ ]?([^\\)]+)\\)")
	result := re.FindAllSubmatch(bte, -1)
	for i := range result {
		item := MappingRoute{}
		item.Method = strings.ToUpper(string(result[i][1]))
		if item.Method != methodCheck.String() {
			continue
		}
		item.Path = string(result[i][2])
		listRoute = append(listRoute, item)
	}

	return
}

/*
mappingRouteJsonFile
Format:

	[
		{
			"method": "GET",
			"path": "/my-endpoint-1"
		},
		{
			"method": "GET",
			"path": "/my-endpoint-2"
		}
	]
*/
func mappingRouteJsonFile(bte []byte, methodCheck method) (listRoute []MappingRoute, err error) {
	tempListRoute := []MappingRoute{}
	errUnmarshal := lib.JSONUnmarshal(bte, &tempListRoute)
	if errUnmarshal != nil {
		err = fmt.Errorf("mappingRouteJsonFile cannot unmarshal json: %s", errUnmarshal)
		return
	}

	for _, item := range tempListRoute {
		item.Method = strings.ToUpper(string(item.Method))
		if item.Method != methodCheck.String() {
			continue
		}

		listRoute = append(listRoute, item)
	}

	return
}

type mapValidateRoute map[string][]string

func newMapValidateRoute() *mapValidateRoute {
	newM := make(mapValidateRoute)
	return &newM
}

func (m *mapValidateRoute) init(k string) {
	if _, ok := (*m)[k]; !ok {
		(*m)[k] = []string{}
	}
}

func (m *mapValidateRoute) add(k, v string) {
	if oldValue, ok := (*m)[k]; !ok {
		(*m)[k] = []string{v}
	} else {
		newValue := append(oldValue, v)
		(*m)[k] = newValue
	}
}

func validateRoute(routerSource RouterSource, listRoute []MappingRoute) (err error) {
	arrMessage := []string{}
	mapPatternMatchRoute := newMapValidateRoute()
	mapRouteMatchPattern := newMapValidateRoute()

	// Append to map
loopFirstRouterSource:
	for routePattern := range routerSource {
		// Try compile pattern
		pattern, errCompile := regexp.Compile(routePattern)
		if errCompile != nil {
			arrMessage = append(arrMessage, fmt.Sprintf("failed compile route pattern: %s, message: %s", routePattern, errCompile.Error()))
			continue loopFirstRouterSource
		}

	loopListRoute:
		for _, route := range listRoute {
			// Init
			mapPatternMatchRoute.init(routePattern)
			mapRouteMatchPattern.init(route.Path)

			// Is pattern match route path?
			if !pattern.MatchString(route.Path) {
				continue loopListRoute
			}

			// Add
			mapPatternMatchRoute.add(routePattern, route.Path)
			mapRouteMatchPattern.add(route.Path, routePattern)
		}
	}

	// Start validate
loopMapPatternMatchRoute:
	for k, v := range *mapPatternMatchRoute {
		// Indicate pattern not match any route
		if len(v) == 0 {
			arrMessage = append(arrMessage, fmt.Sprintf("pattern %s is not match any route", k))
			continue loopMapPatternMatchRoute
		}

		// Indicates the pattern match more than one route
		if len(v) > 1 {
			strRoutes := strings.Join(v, " ,\n ")
			arrMessage = append(arrMessage, fmt.Sprintf("the pattern match more than one route [%s]: \n%s", k, strRoutes))
			continue loopMapPatternMatchRoute
		}
	}

	// Log result
	bytePatternMatchRoute, _ := json.MarshalIndent(mapPatternMatchRoute, "", " ")
	log.Println("validateRoute mapPatternMatchRoute: \n", string(bytePatternMatchRoute))

loopMapRouteMatchPattern:
	for k, v := range *mapRouteMatchPattern {
		if v == nil || len(v) == 1 {
			continue loopMapRouteMatchPattern
		}

		// Indicates the route match more than one pattern
		if len(v) > 1 {
			strPatterns := strings.Join(v, " ,\n ")
			arrMessage = append(arrMessage, fmt.Sprintf("the route match more than one pattern [%s]: \n%s", k, strPatterns))
			continue loopMapRouteMatchPattern
		}
	}

	// Log result
	byteRouteMatchPattern, _ := json.MarshalIndent(mapRouteMatchPattern, "", " ")
	log.Println("validateRoute mapRouteMatchPattern: \n", string(byteRouteMatchPattern))

	// Return err
	if len(arrMessage) > 0 {
		message := strings.Join(arrMessage, ";\n")
		err = fmt.Errorf("ERROR validateRoute: \n %s", message)
		return
	}

	return
}

func checkValidTable(db *gorm.DB, routerSource RouterSource, modelMigrations []interface{}) (err error) {
	arrMessage := []string{}

	formatErr := func(section string, listMessage ...string) error {
		message := strings.Join(listMessage, ";\n")
		return fmt.Errorf("ERROR checkValidTable %s: \n %s", section, message)
	}

	// Get list migration table from list migration model
	listMigrationTable, arrMessage1 := getListMigrationTable(db, modelMigrations)
	if len(arrMessage1) > 0 {
		arrMessage = append(arrMessage, arrMessage1...)
	}

	// Log err
	if len(arrMessage) > 0 {
		err = formatErr("getListMigrationTable", arrMessage...)
		return
	}

	// Compare model migrations with router table
	arrMessage2 := matchingModelMigrationsWithRouterSource(routerSource, listMigrationTable)
	if len(arrMessage2) > 0 {
		arrMessage = append(arrMessage, arrMessage2...)
	}

	// Log err
	if len(arrMessage) > 0 {
		err = formatErr("matchingModelMigrationsWithRouterSource", arrMessage...)
		return
	}

	return
}

func getListMigrationTable(db *gorm.DB, modelMigrations []interface{}) (listMigrationTable, arrMessage []string) {
	if len(modelMigrations) == 0 {
		arrMessage = append(arrMessage, "error getListMigrationTable: model migrations is empty")
		return
	}

loopModelMigrations:
	for idx, model := range modelMigrations {
		assignModel := model
		mod := db.Model(assignModel).Take(assignModel)
		if mod.Error != nil && mod.Error != gorm.ErrRecordNotFound {
			arrMessage = append(arrMessage, fmt.Sprintf("error getListMigrationTable when find table on index: %d, message: %s", idx, mod.Error.Error()))
			continue loopModelMigrations
		}

		table := mod.Statement.Table
		if lib.IsEmptyStr(table) {
			arrMessage = append(arrMessage, fmt.Sprintf("error getListMigrationTable: table is empty on index: %d", idx))
			continue loopModelMigrations
		}

		listMigrationTable = append(listMigrationTable, table)
	}

	return
}

func matchingModelMigrationsWithRouterSource(routerSource RouterSource, listMigrationTable []string) (arrMessage []string) {
	// Compare model migrations with router table
loopRouterSource:
	for pattern, source := range routerSource {
		// Validate source
		isSourceMatch := false
	loopMTable1:
		for _, mTable := range listMigrationTable {
			if mTable == source.Source {
				isSourceMatch = true
				break loopMTable1
			}
		}

		// If source not match any model
		if !isSourceMatch {
			arrMessage = append(arrMessage, fmt.Sprintf("source %s on pattern %s not match any model. Make sure using model pointer on declare model migrations, ex: []interface{\n&model1{}, \n&model2{} \n}", source.Source, pattern))
		}

		// Validate required relation
		if len(source.RequiredRelation) > 0 {
		loopRequiredRelation:
			for rIdx, rTable := range source.RequiredRelation {
				isRequiredMatch := false

			loopMTable2:
				for _, mTable := range listMigrationTable {
					if mTable == rTable {
						isRequiredMatch = true
						break loopMTable2
					}
				}

				// If required relation not match any model
				if !isRequiredMatch {
					arrMessage = append(arrMessage, fmt.Sprintf("required relation %s index %d , on pattern %s not match any model", rTable, rIdx, pattern))
					continue loopRequiredRelation
				}
			}

			continue loopRouterSource
		}

		// Validate ignore relation
	loopIgnoreRelation:
		for iIdx, iTable := range source.IgnoreRelation {
			isIgnoreMatch := false

		loopMTable3:
			for _, mTable := range listMigrationTable {
				if mTable == iTable {
					isIgnoreMatch = true
					break loopMTable3
				}
			}

			// If ignore relation not match any model
			if !isIgnoreMatch {
				arrMessage = append(arrMessage, fmt.Sprintf("ignore relation %s index %d , on pattern %s not match any model", iTable, iIdx, pattern))
				continue loopIgnoreRelation
			}
		}
	}

	return
}
