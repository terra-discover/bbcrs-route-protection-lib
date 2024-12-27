package middleware

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
	model "github.com/terra-discover/bbcrs-migration-lib/model"
	"gorm.io/gorm"
)

func Test_checkValidRoute(t *testing.T) {
	fullPermissionFile := fs.FileMode(0777)

	dummyRoutePrefix := "/my-endpoint"

	setDummyPattern := func(prefix, path string) string {
		return fmt.Sprintf(".*%s/%s?/([^/]+)$", prefix, path)
	}

	type args struct {
		methodCheck   method
		routerSource  RouterSource
		routerFileDir string
		routerPrefix  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "mapping route and validate route are passed, not error",
			args: args{
				methodCheck: DeleteMethod,
				routerSource: RouterSource{
					setDummyPattern(dummyRoutePrefix, "cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
					setDummyPattern(dummyRoutePrefix, "countries"): SourceRelation{
						Source: "country",
						IgnoreRelation: []string{
							"country_translation",
						},
					},
				},
				routerFileDir: func() string {
					fileName := "router.txt"
					file, err := os.Create(fileName)
					utils.AssertEqual(t, nil, err, "mock file")
					defer file.Close()

					err = file.Chmod(fullPermissionFile)
					utils.AssertEqual(t, nil, err, "chmod file")

					_, err = file.WriteString(`
					func Handle(app *fiber.App) {
						app.Delete("/cities/:id", controller.DeleteCity)
						
						app.Delete("/countries/:id", controller.DeleteCountry)

						app.Get("/countries/:id", controller.GetCountry)
					}
					`)
					utils.AssertEqual(t, nil, err, "write file")

					return fileName
				}(),
				routerPrefix: dummyRoutePrefix,
			},
			wantErr: false,
		},
		{
			name: "mapping route passed, validate route failed, error",
			args: args{
				methodCheck: DeleteMethod,
				routerSource: RouterSource{
					setDummyPattern(dummyRoutePrefix, "invalid-endpoint"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
					setDummyPattern(dummyRoutePrefix, "countries"): SourceRelation{
						Source: "country",
						IgnoreRelation: []string{
							"country_translation",
						},
					},
				},
				routerFileDir: func() string {
					fileName := "router-2.txt"
					file, err := os.Create(fileName)
					utils.AssertEqual(t, nil, err, "mock file")
					defer file.Close()

					err = file.Chmod(fullPermissionFile)
					utils.AssertEqual(t, nil, err, "chmod file")

					_, err = file.WriteString(`
					func Handle(app *fiber.App) {
						app.Delete("/cities/:id", controller.DeleteCity)
						
						app.Delete("/countries/:id", controller.DeleteCountry)

						app.Get("/countries/:id", controller.GetCountry)
					}
					`)
					utils.AssertEqual(t, nil, err, "write file")

					return fileName
				}(),
				routerPrefix: dummyRoutePrefix,
			},
			wantErr: true,
		},
		{
			name: "mapping route failed, error",
			args: args{
				methodCheck:  DeleteMethod,
				routerSource: RouterSource{},
				routerFileDir: func() string {
					fileName := "router-3.txt"
					return fileName
				}(),
				routerPrefix: dummyRoutePrefix,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				_, errRead := os.ReadFile(tt.args.routerFileDir)
				if errRead == nil {
					errRemove := os.Remove(tt.args.routerFileDir)
					utils.AssertEqual(t, nil, errRemove, "clean up dir")
				}
			})

			if err := checkValidRoute(tt.args.methodCheck, tt.args.routerSource, tt.args.routerFileDir, tt.args.routerPrefix); (err != nil) != tt.wantErr {
				t.Errorf("checkValidRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mappingRoute(t *testing.T) {
	fullPermissionFile := fs.FileMode(0777)

	type args struct {
		methodCheck   method
		routerFileDir string
		routerPrefix  string
	}
	tests := []struct {
		name          string
		args          args
		wantListRoute []MappingRoute
		wantErr       bool
	}{
		{
			name: "using .txt file, some routes found, not error",
			args: args{
				methodCheck: DeleteMethod,
				routerFileDir: func() string {
					fileName := "router.txt"
					file, err := os.Create(fileName)
					utils.AssertEqual(t, nil, err, "mock file")
					defer file.Close()

					err = file.Chmod(fullPermissionFile)
					utils.AssertEqual(t, nil, err, "chmod file")

					_, err = file.WriteString(`
					func Handle(app *fiber.App) {
						app.Delete("/cities/:id", controller.DeleteCity)
						
						app.Delete("/countries/:id", controller.DeleteCountry)

						app.Get("/countries/:id", controller.GetCountry)
					}
					`)
					utils.AssertEqual(t, nil, err, "write file")

					return fileName
				}(),
				routerPrefix: "/my-endpoint",
			},
			wantListRoute: []MappingRoute{
				{
					Method: DeleteMethod.String(),
					Path:   "/my-endpoint" + "/cities/:id",
				},
				{
					Method: DeleteMethod.String(),
					Path:   "/my-endpoint" + "/countries/:id",
				},
			},
			wantErr: false,
		},
		{
			name: "using .json file, some routes found, not error",
			args: args{
				methodCheck: DeleteMethod,
				routerFileDir: func() string {
					fileName := "router.json"
					file, err := os.Create(fileName)
					utils.AssertEqual(t, nil, err, "mock file")
					defer file.Close()

					err = file.Chmod(fullPermissionFile)
					utils.AssertEqual(t, nil, err, "chmod file")

					_, err = file.WriteString(`
						[
							{
								"method":"Delete",
								"path": "/cities/:id"
							},
							{
								"method":"Delete",
								"path": "/countries/:id"
							},
							{
								"method":"Get",
								"path": "/countries/:id"
							}
						]
					`)
					utils.AssertEqual(t, nil, err, "write file")

					return fileName
				}(),
				routerPrefix: "/my-endpoint",
			},
			wantListRoute: []MappingRoute{
				{
					Method: DeleteMethod.String(),
					Path:   "/my-endpoint" + "/cities/:id",
				},
				{
					Method: DeleteMethod.String(),
					Path:   "/my-endpoint" + "/countries/:id",
				},
			},
			wantErr: false,
		},
		{
			name: "router file not found, error",
			args: args{
				methodCheck: DeleteMethod,
				routerFileDir: func() string {
					fileName := "router.txt"
					return fileName
				}(),
				routerPrefix: "/my-endpoint",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				_, errRead := os.ReadFile(tt.args.routerFileDir)
				if errRead == nil {
					errRemove := os.Remove(tt.args.routerFileDir)
					utils.AssertEqual(t, nil, errRemove, "clean up dir")
				}
			})

			gotListRoute, err := mappingRoute(tt.args.methodCheck, tt.args.routerFileDir, tt.args.routerPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("mappingRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotListRoute, tt.wantListRoute) {
				t.Errorf("mappingRoute() = %v, want %v", gotListRoute, tt.wantListRoute)
			}
		})
	}
}

func Test_newMapValidateRoute(t *testing.T) {
	tests := []struct {
		name string
		want *mapValidateRoute
	}{
		{
			name: "create new map validate route, not error",
			want: func() *mapValidateRoute {
				newM := make(mapValidateRoute)
				return &newM
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMapValidateRoute()
			if !reflect.DeepEqual(got != nil, tt.want != nil) {
				t.Errorf("newMapValidateRoute() not nil = %v, want not nil %v", got != nil, tt.want != nil)

			} else if !reflect.DeepEqual(len(*got), len(*tt.want)) {
				t.Errorf("newMapValidateRoute() = length %v, want length %v", len(*got), len(*tt.want))
			}
		})
	}
}

func Test_mapValidateRoute_init(t *testing.T) {
	setDummyPattern := func(path string) string {
		return fmt.Sprintf(".*/my-endpoint/%s?/([^/]+)$", path)
	}

	dummyPattern := setDummyPattern("cities")

	type args struct {
		k string
	}
	tests := []struct {
		name      string
		m         *mapValidateRoute
		args      args
		wantFound bool
	}{
		{
			name: "key not exist, not error",
			m:    newMapValidateRoute(),
			args: args{
				k: dummyPattern,
			},
			wantFound: true,
		},
		{
			name: "key exist, not error",
			m:    newMapValidateRoute(),
			args: args{
				k: dummyPattern,
			},
			wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.init(tt.args.k)

			_, ok := (*tt.m)[tt.args.k]
			utils.AssertEqual(t, tt.wantFound, ok, "validate found key")
		})
	}
}

func Test_mapValidateRoute_add(t *testing.T) {
	setDummyPattern := func(path string) string {
		return fmt.Sprintf(".*/my-endpoint/%s?/([^/]+)$", path)
	}

	dummyPattern := setDummyPattern("cities")

	dummyPath1 := "/my-endpoint/cities/:id"
	dummyPath2 := "/my-endpoint/cities/:id/translations"

	type args struct {
		k string
		v string
	}
	tests := []struct {
		name      string
		m         *mapValidateRoute
		args      args
		wantValue []string
	}{
		{
			name: "key not exists, create new value, not error",
			m:    newMapValidateRoute(),
			args: args{
				k: dummyPattern,
				v: dummyPath1,
			},
			wantValue: []string{
				dummyPath1,
			},
		},
		{
			name: "key exists, value will append, not error",
			m: func() *mapValidateRoute {
				newM := newMapValidateRoute()
				(*newM)[dummyPattern] = []string{
					dummyPath1,
				}
				return newM
			}(),
			args: args{
				k: dummyPattern,
				v: dummyPath2,
			},
			wantValue: []string{
				dummyPath1,
				dummyPath2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.add(tt.args.k, tt.args.v)

			gotValue, ok := (*tt.m)[tt.args.k]
			utils.AssertEqual(t, len(tt.wantValue) > 0, ok, "validate key")
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("getListMigrationTable() gotValue = %v, wantValue %v", gotValue, tt.wantValue)
			}
		})
	}
}

func Test_validateRoute(t *testing.T) {
	patternReferSomeRoute := ".*/%s?/([^/]+)$"

	routeReferToSomePattern1 := ".*/%s?/([^/]+)$"
	routeReferToSomePattern2 := "/%s?/([^/]+)$"

	type args struct {
		routerSource RouterSource
		listRoute    []MappingRoute
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "one pattern is match one route, success",
			args: args{
				routerSource: RouterSource{
					UseMasterPattern("cities"): SourceRelation{
						Source: "city",
					},
				},
				listRoute: []MappingRoute{
					{
						Method: "DELETE",
						Path:   "/api/v1/master/cities/:id",
					},
				},
			},
			wantErr:         false,
			wantErrContains: "",
		},
		{
			name: "the pattern match more than one route",
			args: args{
				routerSource: RouterSource{
					fmt.Sprintf(patternReferSomeRoute, "cabin-types"): SourceRelation{
						Source: "cabin_type",
					},
				},
				listRoute: []MappingRoute{
					{
						Method: "DELETE",
						Path:   "/api/v1/master/cabin-types/:id",
					},
					{
						Method: "DELETE",
						Path:   "/api/v1/master/integrations-partners/:id/cabin-types/:cabin_type_id",
					},
				},
			},
			wantErr:         true,
			wantErrContains: "the pattern match more than one route",
		},
		{
			name: "the route match more than one pattern",
			args: args{
				routerSource: RouterSource{
					fmt.Sprintf(routeReferToSomePattern1, "cabin-types"): SourceRelation{
						Source: "cabin_type",
					},
					fmt.Sprintf(routeReferToSomePattern2, "cabin-types"): SourceRelation{
						Source: "cabin_type_translation",
					},
				},
				listRoute: []MappingRoute{
					{
						Method: "DELETE",
						Path:   "/api/v1/master/cabin-types/:id",
					},
				},
			},
			wantErr:         true,
			wantErrContains: "the route match more than one pattern",
		},
		{
			name: "pattern is not match any route, error",
			args: args{
				routerSource: RouterSource{
					UseMasterPattern("cities"): SourceRelation{
						Source: "city",
					},
				},
				listRoute: []MappingRoute{
					{
						Method: "DELETE",
						Path:   "/api/v1/master/countries/:id",
					},
				},
			},
			wantErr:         true,
			wantErrContains: "is not match any route",
		},
		{
			name: "pattern cannot be compiled",
			args: args{
				routerSource: RouterSource{
					"*/#": SourceRelation{
						Source: "city",
					},
				},
				listRoute: []MappingRoute{
					{
						Method: "DELETE",
						Path:   "/api/v1/master/countries/:id",
					},
				},
			},
			wantErr:         true,
			wantErrContains: "failed compile route pattern",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoute(tt.args.routerSource, tt.args.listRoute)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("validateRoute() error message = %s, wantErrContains %s", err.Error(), tt.wantErrContains)
			}
		})
	}
}

func Test_checkValidTable(t *testing.T) {
	setDummyPattern := func(path string) string {
		return fmt.Sprintf(".*/my-endpoint/%s?/([^/]+)$", path)
	}

	type invalidStruct struct{}

	type args struct {
		db              *gorm.DB
		routerSource    RouterSource
		modelMigrations []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "all table is valid, no error",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				modelMigrations: []interface{}{
					&model.City{},
					&model.CityTranslation{},
				},
			},
			wantErr: false,
		},
		{
			name: "model not match router source, error",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				modelMigrations: []interface{}{
					&model.City{},
				},
			},
			wantErr: true,
		},
		{
			name: "model migrations is invalid, error",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				modelMigrations: []interface{}{
					&model.City{},
					&model.CityTranslation{},
					&invalidStruct{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				sqlDB, _ := tt.args.db.DB()
				sqlDB.Close()
			})

			if err := checkValidTable(tt.args.db, tt.args.routerSource, tt.args.modelMigrations); (err != nil) != tt.wantErr {
				t.Errorf("checkValidTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getListMigrationTable(t *testing.T) {
	type invalidStruct struct{}

	type args struct {
		db              *gorm.DB
		modelMigrations []interface{}
	}
	tests := []struct {
		name                   string
		args                   args
		wantListMigrationTable []string
		wantArrMessageContains []string
	}{
		{
			name: "all migration table found, not error",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
				modelMigrations: []interface{}{
					&model.Attraction{},
					&model.Country{},
					&model.Zone{},
				},
			},
			wantListMigrationTable: []string{
				"attraction",
				"country",
				"zone",
			},
			wantArrMessageContains: []string{},
		},
		{
			name: "one model migration is invalid, error",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
				modelMigrations: []interface{}{
					&model.Attraction{},
					&model.Country{},
					&model.Zone{},
					&invalidStruct{}, // make error
				},
			},
			wantListMigrationTable: []string{
				"attraction",
				"country",
				"zone",
			},
			wantArrMessageContains: []string{
				"find table on index",
			},
		},
		{
			name: "params model migrations is empty, error",
			args: args{
				db:              testSupport__DBConnectAndSeedTest(),
				modelMigrations: []interface{}{},
			},
			// if empty, no need to declare because we using reflect for checking field on unit test
			// wantListMigrationTable: []string{},
			wantArrMessageContains: []string{
				"model migrations is empty",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				sqlDB, _ := tt.args.db.DB()
				sqlDB.Close()
			})

			gotListMigrationTable, gotArrMessage := getListMigrationTable(tt.args.db, tt.args.modelMigrations)
			if !reflect.DeepEqual(gotListMigrationTable, tt.wantListMigrationTable) {
				t.Errorf("getListMigrationTable() gotListMigrationTable = %v, want %v", gotListMigrationTable, tt.wantListMigrationTable)
			}
			testSupport__validateArrMessage(t, gotArrMessage, tt.wantArrMessageContains)
		})
	}
}

func Test_matchingModelMigrationsWithRouterSource(t *testing.T) {
	setDummyPattern := func(path string) string {
		return fmt.Sprintf(".*/my-endpoint/%s?/([^/]+)$", path)
	}

	type args struct {
		routerSource       RouterSource
		listMigrationTable []string
	}
	tests := []struct {
		name                   string
		args                   args
		wantArrMessageContains []string
	}{
		{
			name: "some relation ignore only, not error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				listMigrationTable: []string{
					"city",
					"city_translation",
				},
			},
			wantArrMessageContains: []string{},
		},
		{
			name: "some relation required only, not error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						RequiredRelation: []string{
							"country",
						},
					},
				},
				listMigrationTable: []string{
					"city",
					"country",
				},
			},
			wantArrMessageContains: []string{},
		},
		{
			name: "some relation required and ignore, must only validate required relation, not error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						RequiredRelation: []string{
							"country",
						},
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				listMigrationTable: []string{
					"city",
					"country",
				},
			},
			wantArrMessageContains: []string{},
		},
		{
			name: "relation ignore, model not found, error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				listMigrationTable: []string{
					"city",
				},
			},
			wantArrMessageContains: []string{
				"ignore relation",
			},
		},
		{
			name: "relation required, model not found, error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						RequiredRelation: []string{
							"country",
						},
					},
				},
				listMigrationTable: []string{
					"city",
				},
			},
			wantArrMessageContains: []string{
				"required relation",
			},
		},
		{
			name: "source declared, model not found, error",
			args: args{
				routerSource: RouterSource{
					setDummyPattern("cities"): SourceRelation{
						Source: "city",
						IgnoreRelation: []string{
							"city_translation",
						},
					},
				},
				listMigrationTable: []string{
					"country",
				},
			},
			wantArrMessageContains: []string{
				"source",
				"ignore relation",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArrMessage := matchingModelMigrationsWithRouterSource(tt.args.routerSource, tt.args.listMigrationTable)
			testSupport__validateArrMessage(t, gotArrMessage, tt.wantArrMessageContains)
		})
	}
}

func testSupport__validateArrMessage(t *testing.T, gotArrMessage, wantArrMessageContains []string) {
	utils.AssertEqual(t, len(wantArrMessageContains), len(gotArrMessage), fmt.Sprintf("arr message length got:\n%s", strings.Join(gotArrMessage, ";\n")))

	notFoundMessage := []string{}
	for _, messageContains := range wantArrMessageContains {
		isFound := false

	loopGotArrMessage:
		for _, gotMess := range gotArrMessage {
			if strings.Contains(gotMess, messageContains) {
				isFound = true
				break loopGotArrMessage
			}
		}

		if !isFound {
			notFoundMessage = append(notFoundMessage, fmt.Sprintf("not found message contains: %s", messageContains))
		}
	}

	printMessage := fmt.Sprintf("%s \n %s",
		strings.Join(notFoundMessage, " ;\n"),
		fmt.Sprintf("only got message: %s",
			strings.Join(gotArrMessage, " ;\n"),
		))
	utils.AssertEqual(t, true, len(notFoundMessage) == 0, printMessage)
}
