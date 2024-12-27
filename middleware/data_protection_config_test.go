package middleware

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/terra-discover/bbcrs-helper-lib/pkg/lib"
)

func Test_patternConstant(t *testing.T) {
	input := "cabin-types"
	pattern, err := regexp.Compile(fmt.Sprintf(masterPattern, input))
	utils.AssertEqual(t, nil, err, "compile regex")

	url1 := "/api/v1/master/integration-partners/id1/cabin-types/id2"
	matched := pattern.FindStringSubmatch(url1)
	utils.AssertEqual(t, false, len(matched) > 0, "match regex")

	url2 := "/api/v1/master/cabin-types/id1"
	matched = pattern.FindStringSubmatch(url2)
	utils.AssertEqual(t, true, len(matched) > 0, "match regex")

	url3 := "/api/v1/master/cabin-types"
	matched = pattern.FindStringSubmatch(url3)
	utils.AssertEqual(t, false, len(matched) > 0, "match regex")
}

func Test_usePattern(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				input: "agent-corporates",
			},
			want: fmt.Sprintf(masterPattern, "agent-corporates"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UseMasterPattern(tt.args.input); got != tt.want {
				t.Errorf("usePattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setDeleteRouterSource(t *testing.T) {
	existKey := UseMasterPattern("countries")
	existVal := SourceRelation{
		Source:         "country",
		IgnoreRelation: []string{"country_translation"},
	}

	deleteRouterSource[existKey] = existVal

	type args struct {
		newRouterSource RouterSource
	}
	tests := []struct {
		name    string
		args    args
		wantKey string
		wantVal SourceRelation
	}{
		{
			name: "router source value and key filled, not exists, save success",
			args: args{
				newRouterSource: RouterSource{
					UseMasterPattern("cities"): SourceRelation{
						Source: "city",
					},
				},
			},
			wantKey: UseMasterPattern("cities"),
			wantVal: SourceRelation{
				Source: "city",
			},
		},
		{
			name: "router source value and key filled, already exists, not saved",
			args: args{
				newRouterSource: RouterSource{
					existKey: SourceRelation{
						Source:         "states",
						IgnoreRelation: []string{"states_translation", "city"},
					},
				},
			},
			wantKey: existKey,
			wantVal: existVal,
		},
		{
			name: "router source key filled, value empty, not saved",
			args: args{
				newRouterSource: RouterSource{
					UseMasterPattern("attractions"): SourceRelation{},
				},
			},
		},
		{
			name: "router source key empty, value filled, not saved",
			args: args{
				newRouterSource: RouterSource{
					"": SourceRelation{
						Source: "attraction_translation",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDeleteRouterSource(tt.args.newRouterSource)

			// get only 1 args key, because filling maximum 1 key in this test
			argsKey := ""
			for k := range tt.args.newRouterSource {
				argsKey = k
			}

			val, ok := deleteRouterSource[argsKey]
			utils.AssertEqual(t, !lib.IsEmptyStr(tt.wantKey), ok, "validate key")
			utils.AssertEqual(t, tt.wantVal.Source, val.Source, "validate source")
			utils.AssertEqual(t, len(tt.wantVal.IgnoreRelation), len(val.IgnoreRelation), "valdate ignore relation")
			utils.AssertEqual(t, len(tt.wantVal.RequiredRelation), len(val.RequiredRelation), "valdate required relation")
		})
	}
}
