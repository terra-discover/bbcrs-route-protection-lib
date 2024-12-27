package middleware

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
)

/*
deleteRouterSource
Format:

	"regex_pattern": sourceRelation{
	    Source: "table_name",
		IgnoreRelation: []string{"table_name_1", "table_name_2"],
		RequiredRelation: []string{"table_name_3", "table_name_4"]
	}

Notes: Regex pattern must capture a group as an id.
To reduce the number of lines, please separate the format for each model.

example:

	var deleteRouterSource = routerSource{
		useMasterPattern("attractions"): sourceRelation{
			Source: "attraction",
			IgnoreRelation: []string{
				"attraction_translation",
				"attraction_category_attraction",
				"attraction_asset",
			},
		},
		useMasterPattern("cities"): sourceRelation{
			Source:         "city",
			RequiredRelation: []string{
				"country"
			},
		},
	}

routerSource = the main table will be deleted by path :id, if delete action on route executed

Note: If RequiredRelation and IgnoreRelation are declared, data protection ONLY validate RequiredRelation
*/
var deleteRouterSource = RouterSource{}

func getDeleteRouterSource() (r RouterSource) {
	r = deleteRouterSource
	return
}

func setDeleteRouterSource(newRouterSource RouterSource) {
	for k, v := range newRouterSource {
		if lib.IsEmptyStr(k) {
			continue
		}
		if lib.IsEmptyStr(v.Source) {
			continue
		}

		if existValue, ok := deleteRouterSource[k]; !ok {
			deleteRouterSource[k] = v
		} else {
			byteVal, _ := json.MarshalIndent(existValue, "", " ")
			log.Printf("exist deleteRouterSource[%s] = %s", k, string(byteVal))
		}
	}
}

// Service prefix endpoint
const (
	masterServiceEndpoint      string = "/api/v1/master"
	integrationServiceEndpoint string = "/api/v1/integration"
	userServiceEndpoint        string = "/api/v1/user"
	paymentServiceEndpoint     string = "/api/v1/payment"
	emailServiceEndpoint       string = "/api/v1/email"
	multimediaServiceEndpoint  string = "/api/v1/multimedia"
)

// Notes: Change this pattern will effects to batch-actions route
// var masterPattern
var (
	masterPattern = ".*" + masterServiceEndpoint + "/%s?/([^/]+)$"
)

// const routerSourcePattern = ".*/%s?/([^/]+)$"

func UseMasterPattern(input string) string {
	return fmt.Sprintf(masterPattern, input)
}
