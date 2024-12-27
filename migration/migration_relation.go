package migration

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
	model "github.com/terra-discover/bbcrs-migration-lib/model"
	ir "github.com/terra-discover/bbcrs-route-protection-lib/migration/internal/relation"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// mapSpecialForeignColumnName - format map[used_by_column]model_source.
// Foreign column which cannot describing it's table name by removing suffix "_id".
// Included in migration relation.
var mapSpecialForeignColumnName = ir.NewColumnModel(map[string]interface{}{
	"parent_agent_message_id":                                 &model.AgentMessage{},
	"document_issue_address_id":                               &model.Address{},
	"destination_airport_id":                                  &model.Airport{},
	"origin_airport_id":                                       &model.Airport{},
	"parent_attraction_category_id":                           &model.AttractionCategory{},
	"parent_capability_category_id":                           &model.CapabilityCategory{},
	"capital_city_id":                                         &model.City{},
	"destination_city_id":                                     &model.City{},
	"origin_city_id":                                          &model.City{},
	"birth_country_id":                                        &model.Country{},
	"citizen_country_id":                                      &model.Country{},
	"timezone_id":                                             &model.Country{},
	"parent_corporate_id":                                     &model.Corporate{},
	"use_credit_limit_from_corporate_id":                      &model.Corporate{},
	"corporate_credit_limit_tolerance_id":                     &model.CorporateCreditLimit{},
	"base_currency_id":                                        &model.Currency{},
	"base_fare_currency_id":                                   &model.Currency{},
	"base_penalty_currency_id":                                &model.Currency{},
	"base_price_currency_id":                                  &model.Currency{},
	"commission_currency_id":                                  &model.Currency{},
	"currency_conversion_to_currency_id":                      &model.Currency{},
	"currency_conversion_from_currency_id":                    &model.Currency{},
	"fare_currency_id":                                        &model.Currency{},
	"foreign_exchange_base_currency_id":                       &model.Currency{},
	"from_currency_id":                                        &model.Currency{},
	"equivalent_fare_currency_id":                             &model.Currency{},
	"equivalent_price_currency_id":                            &model.Currency{},
	"equivalent_penalty_currency_id":                          &model.Currency{},
	"penalty_currency_id":                                     &model.Currency{},
	"published_currency_id":                                   &model.Currency{},
	"published_fare_currency_id":                              &model.Currency{},
	"total_fare_currency_id":                                  &model.Currency{},
	"total_fee_currency_id":                                   &model.Currency{},
	"total_penalty_currency_id":                               &model.Currency{},
	"total_penalty_tax_currency_id":                           &model.Currency{},
	"total_price_currency_id":                                 &model.Currency{},
	"total_tax_currency_id":                                   &model.Currency{},
	"to_currency_id":                                          &model.Currency{},
	"parent_division_id":                                      &model.Division{},
	"contact_employee_id":                                     &model.Employee{},
	"from_employee_id":                                        &model.Employee{},
	"leader_id":                                               &model.Employee{},
	"manager_id":                                              &model.Employee{},
	"to_employee_id":                                          &model.Employee{},
	"destination_hotel_id":                                    &model.Hotel{},
	"parent_hotel_amenity_category_id":                        &model.HotelAmenityCategory{},
	"parent_markup_rate_id":                                   &model.MarkupRate{},
	"parent_link_id":                                          &model.MenuLink{},
	"linked_module_id":                                        &model.Module{},
	"attachment_id":                                           &model.MultimediaDescription{},
	"holder_name_prefix_id":                                   &model.NamePrefix{},
	"linked_page_id":                                          &model.Page{},
	"related_page_id":                                         &model.Page{},
	"contact_person_id":                                       &model.Person{},
	"from_person_id":                                          &model.Person{},
	"on_lap_person_id":                                        &model.Person{},
	"to_person_id":                                            &model.Person{},
	"parent_preference_type_id":                               &model.PreferenceType{},
	"parent_product_category_id":                              &model.ProductCategory{},
	"parent_product_type_id":                                  &model.ProductType{},
	"parent_processing_fee_rate_id":                           &model.ProcessingFeeRate{},
	"other_processing_fee_category_id":                        &model.ProcessingFeeCategory{},
	"retail_processing_fee_category_id":                       &model.ProcessingFeeCategory{},
	"hotel_rating_type_id":                                    &model.RatingType{},
	"parent_room_amenity_category_id":                         &model.RoomAmenityCategory{},
	"after_hour_operation_schedule_id":                        &model.Schedule{},
	"corporate_contract_first_reminder_schedule_id":           &model.Schedule{},
	"corporate_contract_recurrence_reminder_schedule_id":      &model.Schedule{},
	"holding_first_reminder_schedule_id":                      &model.Schedule{},
	"holding_recurrence_reminder_schedule_id":                 &model.Schedule{},
	"invoice_first_reminder_schedule_id":                      &model.Schedule{},
	"invoice_recurrence_reminder_schedule_id":                 &model.Schedule{},
	"last_minute_holding_first_reminder_schedule_id":          &model.Schedule{},
	"last_minute_holding_recurrence_reminder_schedule_id":     &model.Schedule{},
	"last_minute_transaction_first_reminder_schedule_id":      &model.Schedule{},
	"last_minute_transaction_recurrence_reminder_schedule_id": &model.Schedule{},
	"transaction_first_reminder_schedule_id":                  &model.Schedule{},
	"transaction_recurrence_reminder_schedule_id":             &model.Schedule{},
	"travel_document_first_reminder_schedule_id":              &model.Schedule{},
	"travel_document_recurrence_reminder_schedule_id":         &model.Schedule{},
	"parent_service_fee_rate_id":                              &model.ServiceFeeRate{},
	"parent_tax_id":                                           &model.TaxRate{},
	"parent_tax_rate_id":                                      &model.TaxRate{},
	"parent_term_id":                                          &model.Term{},
	"parent_term_category_id":                                 &model.TermCategory{},
	"departure_time_zone_id":                                  &model.TimeZone{},
	"arrival_time_zone_id":                                    &model.TimeZone{},
	"altitude_unit_of_measure_id":                             &model.UnitOfMeasure{},
	"area_unit_of_measure_id":                                 &model.UnitOfMeasure{},
	"distance_unit_of_measure_id":                             &model.UnitOfMeasure{},
	"maximum_vicinity_unit_of_measure_id":                     &model.UnitOfMeasure{},
	"room_size_unit_of_measure_id":                            &model.UnitOfMeasure{},
	"weight_unit_of_measure_id":                               &model.UnitOfMeasure{},
	"approver_id":                                             &model.UserAccount{},
	"approved_by_id":                                          &model.UserAccount{},
	"creator_id":                                              &model.UserAccount{},
	"contact_user_account_id":                                 &model.UserAccount{},
	"from_user_account_id":                                    &model.UserAccount{},
	"modifier_id":                                             &model.UserAccount{},
	"registered_by_id":                                        &model.UserAccount{},
	"rejected_by_id":                                          &model.UserAccount{},
	"terminated_by_id":                                        &model.UserAccount{},
	"to_user_account_id":                                      &model.UserAccount{},
})

// mapNotDeclaredForeignColumnTable - format map[used_by_column]table_source.
// Foreign column which have valid model, but the models are not describing in modelMigrations.
// Excluded from migration relation.
var mapNotDeclaredForeignColumnTable = ir.NewColumnTable(map[string]string{
	"bank_account_id":             "bank_account",
	"business_partner_id":         "business_partner",
	"business_partner_type_id":    "business_partner_type",
	"business_partner_group_id":   "business_partner_group",
	"guarantee_payment_policy_id": "guarantee_payment_policy",
	"payment_card_id":             "payment_card",
	"product_price_type_id":       "product_price_type",
})

// listUnknownForeignColumn - format []string{used_by_column}.
// Foreign column from unknown source.
// Excluded from migration relation.
var listUnknownForeignColumn = []string{
	"agency_id",
	"apply_bundling_rate_id",
	"arrival_transport_id",
	"business_partner_id",
	"business_partner_type_id",
	"build_in_type_id",
	"credit_note_id",
	"departure_transport_id",
	"destination_system_id",
	"distribution_partner_id",
	"email_template_office_id",
	"hotel_group_id",
	"inquiry_id",
	"invoice_group_id",
	"invoice_id",
	"merchant_id",
	"news_id",
	"nominee_id",
	"preference_type",
	"processor_id",
	"profile_id",
	"promotion_id",
	"rate_plan_id",
	"reference_id",
	"relative_position_id",
	"reward_balance_transaction_id",
	"offer_discount_id",
	"offer_id",
	"supplier_id",
	"system_id",
	"terminal_id",
	"viewership_id",
	"voucher_transaction_id",
}

// listCachingPrefixTable - format []string{prefix of used_by_table}.
// List prefix of caching table.
// Excluded from migration relation.
var listCachingPrefixTable = []string{
	"flight_caching",
	"job_worker",
	"relation_schema",
}

// migrateRelation - Must called after all tables has migrated
func migrateRelation(db *gorm.DB, migrationsModel []interface{}, removeOldData bool) error {
	formatErr := func(section, message string) error {
		return fmt.Errorf("ERROR migrateRelation %s: \n%s", section, message)
	}

	// map[foreign_key]table_name
	var mapSourceForeignColumn = ir.NewColumnTable(map[string]string{})

	mustFieldSuffix := "_id"
	avoidFields := []string{}

	// Get all migration column table
	requirement := ir.SetRequirement(mustFieldSuffix, avoidFields)
	mapMigrationColumnSource, mapMigrationColumnUsedBy, err := ir.NewMigrationModel(db, migrationsModel, requirement).ToColumnTable()
	if err != nil {
		return err
	}

	// Get list main table only
	listMainTable := mapMigrationColumnSource.GetListTable()

	// Get special foreign column table
	specialForeignColumnTable, err := mapSpecialForeignColumnName.ToColumnTable(db)
	if err != nil {
		return err
	}

	arrMessage := []string{}

loopSchemaColumns:
	for columnSource, usedByTables := range mapMigrationColumnUsedBy {
		// --Excluded Column Section--
		// 1. Find field in list mapNotDeclaredForeignColumnTable
		//  2. If not found, Find field in listUnkownForeignColumn
		//  3. If not found, Find field in listCachingPrefixTable
		isExcluded := excludeColumnSection(columnSource, usedByTables)
		if isExcluded {
			continue loopSchemaColumns
		}

		// --Included Column Section--
		// 4. Find on list special foreign table name
		// 5. Cut suffix _id from column name = foreign table name
		// 6. If not found, Find foreign table name in list table name
		reqInclude := includeColumnSectionRequest{
			ColumnSource:              columnSource,
			MustFieldSuffix:           mustFieldSuffix,
			SpecialForeignColumnTable: specialForeignColumnTable,
			MapSourceForeignColumn:    mapSourceForeignColumn,
			ListMainTable:             listMainTable,
		}
		isIncluded, err := includeColumnSection(reqInclude)
		if err != nil {
			arrMessage = append(arrMessage, err.Error())
			continue loopSchemaColumns
		} else if isIncluded {
			continue loopSchemaColumns
		}

		// 7. If not found, return err
		arrMessage = append(arrMessage, fmt.Sprintf("schema column of column name: %s (used by table: %s), is not found", columnSource, usedByTables))
		continue loopSchemaColumns
	}

	// Return err
	if len(arrMessage) > 0 {
		message := strings.Join(arrMessage, " \n -")
		err = formatErr("mapMigrationColumnUsedBy", message)
		return err
	}

	// Map source and usedby and append to relation schema
	listRelationSchema := mappingRelationSchema(mapSourceForeignColumn, mapMigrationColumnUsedBy)

	// Start tx
	tx := db.Begin()

	// Generate table relation schema
	err = genTableRelationSchema(tx)
	if err != nil {
		tx.Rollback()
		return formatErr("genTableRelationSchema", err.Error())
	}

	// Remove old relation schema
	if removeOldData {
		err = mustDeleteOldData(tx)
		if err != nil {
			tx.Rollback()
			return formatErr("mustDeleteOldData", err.Error())
		}
	}

	// Publish relation schema to DB
	// Must disable nested tx on CreateInBatches
	if err := tx.Session(&gorm.Session{
		DisableNestedTransaction: true,
	}).Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(&listRelationSchema, 100).Error; err != nil {
		tx.Rollback()
		return formatErr("CreateInBatches relation schema", err.Error())
	}

	// Commit tx
	tx.Commit()

	return nil
}

/*
excludeColumnSection --Excluded Column Section--
 1. Find field in list mapNotDeclaredForeignColumnTable
 2. If not found, Find field in listUnkownForeignColumn
 3. If not found, Find field in listCachingPrefixTable
*/
func excludeColumnSection(columnSource string, usedByTables []string) (isExcluded bool) {
	strUsedByTables := strings.Join(usedByTables, " | ")

	// --Excluded Column Section--
	// 1. Find field in list mapNotDeclaredForeignColumnTable
	if mapNotDeclaredForeignColumnTable != nil {
	loopNotDeclared:
		for notDeclaredColumn, notDeclaredTable := range mapNotDeclaredForeignColumnTable.GetData() {
			if notDeclaredColumn == columnSource {
				log.Printf("INFO Not Declared: \n used_by_column %s,\n used_by_table %s,\n table_source %s", columnSource, strUsedByTables, notDeclaredTable)
				isExcluded = true
				break loopNotDeclared
			}
		}
	}

	if isExcluded {
		return
	}

	// 2. If not found, Find field in listUnkownForeignColumn
loopUnknown:
	for _, unknownForeignColumn := range listUnknownForeignColumn {
		if unknownForeignColumn == columnSource {
			log.Printf("INFO Unknown:\n used_by_column %s,\n used_by_table %s", columnSource, strUsedByTables)
			isExcluded = true
			break loopUnknown
		}
	}

	if isExcluded {
		return
	}

	// 3. If not found, Find field in listCachingPrefixTable
loopUsedByTables:
	for _, usedByTable := range usedByTables {
		for _, cPrefixTable := range listCachingPrefixTable {
			if strings.HasPrefix(usedByTable, cPrefixTable) {
				log.Printf("INFO Caching Prefix Table:\n used_by_column %s,\n used_by_table %s", columnSource, usedByTable)
				isExcluded = true
				break loopUsedByTables
			}
		}
	}

	if isExcluded {
		return
	}

	return
}

type includeColumnSectionRequest struct {
	ColumnSource              string
	MustFieldSuffix           string
	SpecialForeignColumnTable ir.ColumnTable
	MapSourceForeignColumn    ir.IColumnTable
	ListMainTable             []string
}

/*
includeColumnSection --Included Column Section--
 4. Find on list special foreign table name
 5. Cut suffix _id from column name = foreign table name
 6. If not found, Find foreign table name in list table name
*/
func includeColumnSection(req includeColumnSectionRequest) (isIncluded bool, err error) {
	formatErr := func(section, message string) error {
		return fmt.Errorf("ERROR includeColumnSection %s: \n%s", section, message)
	}

	var (
		columnSource              string          = req.ColumnSource
		mustFieldSuffix           string          = req.MustFieldSuffix
		specialForeignColumnTable ir.ColumnTable  = req.SpecialForeignColumnTable
		mapSourceForeignColumn    ir.IColumnTable = req.MapSourceForeignColumn
		listMainTable             []string        = req.ListMainTable
	)

	// 4. Find on list special foreign table name
loopSpecial:
	for specialColumn, specialTable := range specialForeignColumnTable {
		if columnSource == specialColumn {
			errAdd := mapSourceForeignColumn.Add(specialColumn, specialTable)
			if errAdd != nil {
				err = formatErr("loop specialForeignColumnTable", errAdd.Error())
				break loopSpecial
			}
			isIncluded = true
			break loopSpecial
		}
	}

	if err != nil || isIncluded {
		return
	}

	// 5. Cut suffix _id from column name = foreign table name
	foreignTable := strings.TrimSuffix(columnSource, mustFieldSuffix)

	// 6. If not found, Find foreign table name in list table name
loopMain:
	for _, mainTable := range listMainTable {
		if mainTable == foreignTable {
			errAdd := mapSourceForeignColumn.Add(columnSource, foreignTable)
			if errAdd != nil {
				err = formatErr("loop listMainTable", errAdd.Error())
				break loopMain
			}
			isIncluded = true
			break loopMain
		}
	}

	if err != nil || isIncluded {
		return
	}

	return
}

// mappingRelationSchema -
// Map source and usedby and append to relation schema
func mappingRelationSchema(mapSourceForeignColumn ir.IColumnTable, mapMigrationColumnUsedBy ir.ColumnTables) (listRelationSchema []model.RelationSchema) {
	for columnUsedBy1, tableSource := range mapSourceForeignColumn.GetData() {
		for columnUsedBy2, usedByTables := range mapMigrationColumnUsedBy {
			if columnUsedBy1 == columnUsedBy2 {
				for _, usedByTable := range usedByTables {
					newRelationSchema := model.RelationSchema{}
					newRelationSchema.ColumnSource = lib.Strptr("id")
					newRelationSchema.TableSource = lib.Strptr(tableSource)
					newRelationSchema.UsedByColumn = lib.Strptr(columnUsedBy2)
					newRelationSchema.UsedByTable = lib.Strptr(usedByTable)
					// Append to list
					listRelationSchema = append(listRelationSchema, newRelationSchema)
				}
			}
		}
	}

	return
}

// genTableRelationSchema - auto create table relation_schema if not exists
func genTableRelationSchema(db *gorm.DB) (err error) {
	isHasTable := db.Migrator().HasTable(&model.RelationSchema{})
	if !isHasTable {
		err = db.Migrator().CreateTable(&model.RelationSchema{})
	}
	return
}

func mustDeleteOldData(db *gorm.DB) (err error) {
	err = db.Unscoped().Where(`deleted_at IS NULL`).Delete(&model.RelationSchema{}).Error
	return
}

func getLatestMigrateRelation(db *gorm.DB) (lastUpdated time.Time, isFound bool, err error) {
	latestRelationSchema := model.RelationSchema{}
	res := db.Order(`updated_at DESC`).Take(&latestRelationSchema)
	if latestRelationSchema.UpdatedAt != nil {
		lastUpdated = time.Time(*latestRelationSchema.UpdatedAt)
	}
	isFound = res.RowsAffected == 1
	// Don't return error, if record not found
	if res.Error != gorm.ErrRecordNotFound {
		err = res.Error
	}
	return
}
