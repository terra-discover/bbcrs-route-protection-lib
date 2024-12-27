package middleware

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/google/uuid"
	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	standardModel "github.com/terra-discover/bbcrs-migration-lib/model"
	"github.com/terra-discover/bbcrs-route-protection-lib/model"
	"gorm.io/gorm"
)

func Test_routerMaps_addMap(t *testing.T) {
	routerMaps1 := make(routerMaps)
	key1 := "abc"
	value1 := make(routerMap)

	routerMaps2 := make(routerMaps)
	key2 := "def"
	value2 := make(routerMap)
	value2["test"] = "test"
	routerMaps2[key2] = value2

	type args struct {
		k string
		v routerMap
	}
	tests := []struct {
		name        string
		r           *routerMaps
		args        args
		wantIsExist bool
	}{
		{
			name: "not exist",
			r:    &routerMaps1,
			args: args{
				k: key1,
				v: value1,
			},
			wantIsExist: false,
		},
		{
			name: "exist",
			r:    &routerMaps2,
			args: args{
				k: key2,
				v: value2,
			},
			wantIsExist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.addMap(tt.args.k, tt.args.v)
		})
	}
}

func Test_getListRelationSchema(t *testing.T) {
	db := testSupport__DBConnectAndSeedTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Case 1: empty relation schema, not err
	gotListRelationSchema, err := getListRelationSchema(db)
	utils.AssertEqual(t, true, err == nil, "validate err")
	utils.AssertEqual(t, false, len(gotListRelationSchema) > 0, "validate value")

	// Case 2: found relation schema, not err
	newRelationSchema := standardModel.RelationSchema{}
	newRelationSchema.TableSource = lib.Strptr("country")
	newRelationSchema.ColumnSource = lib.Strptr("id")
	newRelationSchema.UsedByTable = lib.Strptr("city")
	newRelationSchema.UsedByColumn = lib.Strptr("country_id")
	err = db.Create(&newRelationSchema).Error
	utils.AssertEqual(t, true, err == nil, "mock data")

	gotListRelationSchema, err = getListRelationSchema(db)
	utils.AssertEqual(t, true, err == nil, "validate err")
	utils.AssertEqual(t, true, len(gotListRelationSchema) > 0, "validate value")

	// Case 3: db closed, err
	sqlDB, err := db.DB()
	utils.AssertEqual(t, true, err == nil, "mock db")
	sqlDB.Close()

	gotListRelationSchema, err = getListRelationSchema(db)
	utils.AssertEqual(t, false, err == nil, "validate err")
	utils.AssertEqual(t, false, len(gotListRelationSchema) > 0, "validate value")
}

func Test_generateDeleteRouteMaps(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name                string
		args                args
		wantDeleteRouteMaps routerMaps
		wantErr             bool
	}{
		{
			name: "",
			args: args{
				db: testSupport__DBConnectAndSeedTest(),
			},
			wantDeleteRouteMaps: routerMaps{},
			wantErr:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDeleteRouteMaps, err := generateDeleteRouteMaps(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDeleteRouteMaps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDeleteRouteMaps, tt.wantDeleteRouteMaps) {
				t.Errorf("generateDeleteRouteMaps() = %v, want %v", gotDeleteRouteMaps, tt.wantDeleteRouteMaps)
			}
		})
	}
}

func TestDataProtection(t *testing.T) {
	db := testSupport__DBConnectAndSeedTest()
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		return runDataProtection(c, db)
	})

	// Case 1: Only declare Source
	// Result: Validate all relation (Not Allowed)
	app.Delete("/api/v1/master/agent-corporates/:id", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": 200})
	})

	// make relation schema
	relationSchema := standardModel.RelationSchema{
		ColumnSource: lib.Strptr("id"),
		TableSource:  lib.Strptr("corporate"),
		UsedByColumn: lib.Strptr("corporate_id"),
		UsedByTable:  lib.Strptr("agent_corporate"),
	}
	mod := db.Create(&relationSchema)
	utils.AssertEqual(t, nil, mod.Error)

	// Set dummy config
	deleteRouterSource = RouterSource{
		UseMasterPattern("agent-corporates"): SourceRelation{
			Source: "corporate",
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	id := uuid.New()
	agentCorporate := standardModel.AgentCorporate{}
	agentCorporate.CorporateID = &id
	agentCorporate.AgentID = &id
	mod = db.Create(&agentCorporate)
	utils.AssertEqual(t, nil, mod.Error)

	res, body, err := lib.DeleteTest(app, "/api/v1/master/agent-corporates/"+id.String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 405, res.StatusCode, "Must be not allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	res, body, err = lib.DeleteTest(app, "/api/v1/master/agent-corporates/"+uuid.New().String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 200, res.StatusCode, "Must be ok")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	// Case 2: More test, only declare Source
	// Result: Validate all relation (Not Allowed)
	app.Delete("/api/v1/master/cities/:id", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": 200})
	})

	// make relation schema
	listRelationSchema := []standardModel.RelationSchema{
		{
			ColumnSource: lib.Strptr("id"),
			TableSource:  lib.Strptr("city"),
			UsedByColumn: lib.Strptr("city_id"),
			UsedByTable:  lib.Strptr("city_translation"),
		},
		{
			ColumnSource: lib.Strptr("id"),
			TableSource:  lib.Strptr("city"),
			UsedByColumn: lib.Strptr("city_id"),
			UsedByTable:  lib.Strptr("airport"),
		},
	}
	mod = db.CreateInBatches(&listRelationSchema, 10)
	utils.AssertEqual(t, nil, mod.Error)

	// Set dummy config
	deleteRouterSource = RouterSource{
		UseMasterPattern("cities"): SourceRelation{
			Source: "city",
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	id = uuid.New()
	city := standardModel.City{}
	city.ID = &id
	city.CityCode = lib.Strptr(lib.RandomChars(6))
	city.CityName = lib.Strptr(lib.RandomChars(6))
	mod = db.Create(&city)
	utils.AssertEqual(t, nil, mod.Error)

	cityTranslation := standardModel.CityTranslation{
		CityID: city.ID,
		CityTranslationAPI: standardModel.CityTranslationAPI{
			LanguageCode: lib.Strptr("id"),
			CityName:     lib.Strptr(lib.RandomChars(10)),
		},
	}
	mod = db.Create(&cityTranslation)
	utils.AssertEqual(t, nil, mod.Error)

	airport := standardModel.Airport{
		AirportAPI: standardModel.AirportAPI{
			AirportCode: lib.Strptr(lib.RandomChars(6)),
			AirportName: lib.Strptr(lib.RandomChars(6)),
			CityID:      city.ID,
		},
	}
	mod = db.Create(&airport)
	utils.AssertEqual(t, nil, mod.Error)

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+id.String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 405, res.StatusCode, "Must be not allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+uuid.New().String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 200, res.StatusCode, "Must be ok")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	// Case 3: Using IgnoreRelation
	// Result: Validate IgnoreRelation (Allowed)
	// Set dummy config with ignore relation
	deleteRouterSource = RouterSource{
		UseMasterPattern("cities"): SourceRelation{
			Source:         "city",
			IgnoreRelation: []string{"airport", "city_translation"},
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+id.String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 200, res.StatusCode, "Must be allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	// Case 4: Using RequiredRelation
	// Result: Validate RequiredRelation (Not Allowed)
	// Set dummy config
	deleteRouterSource = RouterSource{
		UseMasterPattern("cities"): SourceRelation{
			Source: "city",
			RequiredRelation: []string{
				"airport",
			},
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	id = uuid.New()
	city = standardModel.City{}
	city.ID = &id
	city.CityCode = lib.Strptr(lib.RandomChars(6))
	city.CityName = lib.Strptr(lib.RandomChars(6))
	mod = db.Create(&city)
	utils.AssertEqual(t, nil, mod.Error)

	cityTranslation = standardModel.CityTranslation{
		CityID: city.ID,
		CityTranslationAPI: standardModel.CityTranslationAPI{
			LanguageCode: lib.Strptr("id"),
			CityName:     lib.Strptr(lib.RandomChars(10)),
		},
	}
	mod = db.Create(&cityTranslation)
	utils.AssertEqual(t, nil, mod.Error)

	airport = standardModel.Airport{
		AirportAPI: standardModel.AirportAPI{
			AirportCode: lib.Strptr(lib.RandomChars(6)),
			AirportName: lib.Strptr(lib.RandomChars(6)),
			CityID:      city.ID,
		},
	}
	mod = db.Create(&airport)
	utils.AssertEqual(t, nil, mod.Error)

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+id.String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 405, res.StatusCode, "Must be not allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+uuid.New().String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 200, res.StatusCode, "Must be ok")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	// Case 5: Using RequiredRelation and IgnoreRelation at the same time.
	// Result: Will only validate RequiredRelation (Not Allowed)
	// Set dummy config with ignore relation
	deleteRouterSource = RouterSource{
		UseMasterPattern("cities"): SourceRelation{
			Source: "city",
			RequiredRelation: []string{
				"airport",
			},
			IgnoreRelation: []string{
				"airport",
			},
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	res, body, err = lib.DeleteTest(app, "/api/v1/master/cities/"+id.String(), nil)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 405, res.StatusCode, "Must be not allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	// Case 6: Batch Action endpoint (Special case)
	// Result: Validate all relation (Not Allowed)
	app.Post("/api/v1/master/batch-actions/:action/:module", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": 200})
	})

	// make relation schema
	relationSchema = standardModel.RelationSchema{
		ColumnSource: lib.Strptr("id"),
		TableSource:  lib.Strptr("country"),
		UsedByColumn: lib.Strptr("country_id"),
		UsedByTable:  lib.Strptr("city"),
	}
	mod = db.Create(&relationSchema)
	utils.AssertEqual(t, nil, mod.Error)

	// Set dummy config
	deleteRouterSource = RouterSource{
		UseMasterPattern("countries"): SourceRelation{
			Source:         "country",
			IgnoreRelation: []string{"country_translation"},
		},
	}

	// GenerateDeleteRouteMaps first
	generateDeleteRouteMaps(db)

	id = uuid.New()
	country := standardModel.Country{}
	country.ID = &id
	country.CountryCode = lib.Strptr(lib.RandomChars(6))
	country.CountryName = lib.Strptr(lib.RandomChars(6))
	mod = db.Create(&country)
	utils.AssertEqual(t, nil, mod.Error)

	city = standardModel.City{}
	city.CityCode = lib.Strptr(lib.RandomChars(6))
	city.CityName = lib.Strptr(lib.RandomChars(6))
	city.CountryID = country.ID
	mod = db.Create(&city)
	utils.AssertEqual(t, nil, mod.Error)

	payload := []uuid.UUID{
		*country.ID,
	}

	payloadByte, err := lib.JSONMarshal(payload)
	utils.AssertEqual(t, nil, err, "marshal payload")

	payloadJson := string(payloadByte)

	res, body, err = lib.PostTest(app, "/api/v1/master/batch-actions/delete/country", nil, payloadJson)
	utils.AssertEqual(t, nil, err, "Must success")
	// buf := new(bytes.Buffer)
	// buf.ReadFrom(res.Body)
	// bodyJson := buf.String()
	// log.Println("bodyJson", bodyJson)
	utils.AssertEqual(t, 405, res.StatusCode, "Must be not allowed")
	utils.AssertEqual(t, false, nil == body, "Body is required")

	payload = []uuid.UUID{
		uuid.New(),
	}

	payloadByte, err = lib.JSONMarshal(payload)
	utils.AssertEqual(t, nil, err, "marshal payload")

	payloadJson = string(payloadByte)

	res, body, err = lib.PostTest(app, "/api/v1/master/batch-actions/delete/country", nil, payloadJson)
	utils.AssertEqual(t, nil, err, "Must success")
	utils.AssertEqual(t, 200, res.StatusCode, "Must be ok")
	utils.AssertEqual(t, false, nil == body, "Body is required")
}

func TestRouterSource_toRouterMaps(t *testing.T) {
	setDummyPattern := func(prefix, path string) string {
		return fmt.Sprintf(".*%s/%s?/([^/]+)$", prefix, path)
	}

	type args struct {
		listRelationSchema []model.RelationSchema
	}
	tests := []struct {
		name                string
		rs                  *RouterSource
		args                args
		wantDeleteRouteMaps routerMaps
		wantErr             bool
	}{
		{
			name: "generated 1 router maps, using required relation, not error",
			rs: func() *RouterSource {
				deleteRouterSource = RouterSource{
					setDummyPattern("/my-endpoint", "cities"): SourceRelation{
						Source:           "city",
						RequiredRelation: []string{"state_province"},
					},
				}
				newRs := deleteRouterSource
				return &newRs
			}(),
			args: args{
				listRelationSchema: []model.RelationSchema{
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  lib.Strptr("city"),
						UsedByColumn: lib.Strptr("city_id"),
						UsedByTable:  lib.Strptr("state_province"),
					},
				},
			},
			wantDeleteRouteMaps: routerMaps{
				setDummyPattern("/my-endpoint", "cities"): routerMap{
					"state_province": "city_id",
				},
			},
			wantErr: false,
		},
		{
			name: "generated 1 router maps, using ignore relation, not error",
			rs: func() *RouterSource {
				deleteRouterSource = RouterSource{
					setDummyPattern("/my-endpoint", "cities"): SourceRelation{
						Source:         "city",
						IgnoreRelation: []string{"city_translation"},
					},
				}
				newRs := deleteRouterSource
				return &newRs
			}(),
			args: args{
				listRelationSchema: []model.RelationSchema{
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  lib.Strptr("city"),
						UsedByColumn: lib.Strptr("city_id"),
						UsedByTable:  lib.Strptr("state_province"),
					},
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  lib.Strptr("city"),
						UsedByColumn: lib.Strptr("city_id"),
						UsedByTable:  lib.Strptr("city_translation"),
					},
				},
			},
			wantDeleteRouteMaps: routerMaps{
				setDummyPattern("/my-endpoint", "cities"): routerMap{
					"state_province": "city_id",
				},
			},
			wantErr: false,
		},
		{
			name: "relation schema.table source is nil, error",
			rs: func() *RouterSource {
				deleteRouterSource = RouterSource{
					setDummyPattern("/my-endpoint", "cities"): SourceRelation{
						Source:         "city",
						IgnoreRelation: []string{"city_translation"},
					},
				}
				newRs := deleteRouterSource
				return &newRs
			}(),
			args: args{
				listRelationSchema: []model.RelationSchema{
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  nil,
						UsedByColumn: lib.Strptr("city_id"),
						UsedByTable:  lib.Strptr("state_province"),
					},
				},
			},
			wantDeleteRouteMaps: routerMaps{},
			wantErr:             true,
		},
		{
			name: "relation schema.used by table is nil, error",
			rs: func() *RouterSource {
				deleteRouterSource = RouterSource{
					setDummyPattern("/my-endpoint", "cities"): SourceRelation{
						Source:         "city",
						IgnoreRelation: []string{"city_translation"},
					},
				}
				newRs := deleteRouterSource
				return &newRs
			}(),
			args: args{
				listRelationSchema: []model.RelationSchema{
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  lib.Strptr("city"),
						UsedByColumn: lib.Strptr("city_id"),
						UsedByTable:  nil,
					},
				},
			},
			wantDeleteRouteMaps: routerMaps{},
			wantErr:             true,
		},
		{
			name: "relation schema.used by table is nil, error",
			rs: func() *RouterSource {
				deleteRouterSource = RouterSource{
					setDummyPattern("/my-endpoint", "cities"): SourceRelation{
						Source:         "city",
						IgnoreRelation: []string{"city_translation"},
					},
				}
				newRs := deleteRouterSource
				return &newRs
			}(),
			args: args{
				listRelationSchema: []model.RelationSchema{
					{
						ColumnSource: lib.Strptr("id"),
						TableSource:  lib.Strptr("city"),
						UsedByColumn: nil,
						UsedByTable:  lib.Strptr("state_province"),
					},
				},
			},
			wantDeleteRouteMaps: routerMaps{},
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDeleteRouteMaps, err := tt.rs.toRouterMaps(tt.args.listRelationSchema)
			if (err != nil) != tt.wantErr {
				t.Errorf("RouterSource.toRouterMaps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDeleteRouteMaps, tt.wantDeleteRouteMaps) {
				t.Errorf("RouterSource.toRouterMaps() = %v, want %v", gotDeleteRouteMaps, tt.wantDeleteRouteMaps)
			}
		})
	}
}

func Test_generateDataProtectionQuery(t *testing.T) {
	uuid1 := *lib.GenUUID()
	uuid2 := *lib.GenUUID()

	type args struct {
		tables routerMap
		ids    []uuid.UUID
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "2 queries generated, not error",
			args: args{
				tables: routerMap{
					"city":           "country_id",
					"state_province": "country_id",
				},
				ids: []uuid.UUID{
					uuid1,
					uuid2,
				},
			},
			want: fmt.Sprintf(`
				SELECT SUM("s"."total") "total" FROM (
					SELECT COUNT(*) total FROM "city" 
					WHERE "city"."country_id" IN(%[1]s) AND "city"."deleted_at" IS NULL
					UNION
					SELECT COUNT(*) total FROM "state_province" 
					WHERE "state_province"."country_id" IN(%[1]s) AND "state_province"."deleted_at" IS NULL
				) "s"
			`,
				lib.ConvertSliceUUIDToStr([]uuid.UUID{
					uuid1,
					uuid2,
				}, ",", `'%s'`),
			),
		},
		{
			name: "no query generated, not error",
			args: args{
				ids: []uuid.UUID{
					uuid1,
					uuid2,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDataProtectionQuery(tt.args.tables, tt.args.ids)
			gotNoSpace := strings.Join(strings.Fields(got), " ")
			wantNoSpace := strings.Join(strings.Fields(tt.want), " ")
			if !strings.EqualFold(gotNoSpace, wantNoSpace) {
				t.Errorf("generateDataProtectionQuery() = %v, want %v", gotNoSpace, wantNoSpace)
			}
		})
	}
}
