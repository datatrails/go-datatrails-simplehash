package query

import (
	"net/url"
	"testing"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/assets/v2/assets"
	"github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/filter/v1/filter"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFilterParsing(t *testing.T) {
	logger.New("NOOP")
	tests := []struct {
		name           string
		message        *assets.ListAssetsRequest
		values         url.Values
		parser         runtime.QueryParameterParser
		isError        bool
		expectedResult [][]string
	}{
		{
			name:           "test simple or",
			message:        &assets.ListAssetsRequest{},
			values:         map[string][]string{"filters": {"attributes.foo=2 OR attributes.bar=2"}},
			parser:         &runtime.DefaultQueryParser{},
			expectedResult: [][]string{{"attributes.foo=2", "attributes.bar=2"}},
		},
		{
			name:           "test multiple or",
			message:        &assets.ListAssetsRequest{},
			values:         map[string][]string{"filters": {"attributes.foo='2' OR attributes.bar=2", "attributes.foo='dd xx' OR attributes.bar='77'"}},
			parser:         &runtime.DefaultQueryParser{},
			expectedResult: [][]string{{"attributes.foo=2", "attributes.bar=2"}, {"attributes.foo=dd xx", "attributes.bar=77"}},
		},
		{
			name:           "test or *",
			message:        &assets.ListAssetsRequest{},
			values:         map[string][]string{"filters": {"attributes.foo=* oR attributes.bar=2"}},
			parser:         &runtime.DefaultQueryParser{},
			expectedResult: [][]string{{"attributes.foo=*", "attributes.bar=2"}},
		},
		{
			name:           "test or 5 terms",
			message:        &assets.ListAssetsRequest{},
			values:         map[string][]string{"filters": {"attributes.foo=2 or attributes.bar=2 OR attributes.lop=2 oR attributes.foo=22 Or attributes.bar=77"}},
			parser:         &runtime.DefaultQueryParser{},
			expectedResult: [][]string{{"attributes.foo=2", "attributes.bar=2", "attributes.lop=2", "attributes.foo=22", "attributes.bar=77"}},
		},
		{
			name:    "test 4 filter terms",
			message: &assets.ListAssetsRequest{},
			values: map[string][]string{"filters": {
				"attributes.foo=2 or attributes.bar=2 OR attributes.lop=2",
				"attribtues.basss=e88 oR attributes.foxo=22",
				"attribtues.ba=e oR attributes.foo=22 Or attributes.bar=77",
				"something=xx oR attributes.gg=22 Or attributes.bar=87",
			}},
			parser: &runtime.DefaultQueryParser{},
			expectedResult: [][]string{
				{"attributes.foo=2", "attributes.bar=2", "attributes.lop=2"},
				{"attribtues.basss=e88", "attributes.foxo=22"},
				{"attribtues.ba=e", "attributes.foo=22", "attributes.bar=77"},
				{"something=xx", "attributes.gg=22", "attributes.bar=87"}},
		},
		{
			name:           "test bad filter",
			message:        &assets.ListAssetsRequest{},
			values:         map[string][]string{"filters": {"attributes.foo=2 OF attributes.bar=2"}},
			parser:         &runtime.DefaultQueryParser{},
			isError:        true,
			expectedResult: [][]string{},
		},
		{
			name:    "test nil message",
			message: nil,
			values:  map[string][]string{"filters": {"attributes.foo=2 OF attributes.bar=2"}},
			parser:  &runtime.DefaultQueryParser{},
			isError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			result := handleFilters(test.message, test.parser, protoreflect.Value{}, test.message.ProtoReflect().Descriptor().Fields().ByName("filters"), test.values, nil)

			if !test.isError {
				assert.Nil(t, result)
			}
			if result == nil && test.message != nil {

				assert.Equal(t, len(test.message.Filters), len(test.expectedResult))
				for idx, expected := range test.expectedResult {
					assert.Equal(t, test.message.Filters[idx].Or, expected)
				}

			}
		})
	}

}

func TestGetFilterParts(t *testing.T) {
	logger.New("NOOP")
	tests := []struct {
		name           string
		filterSpec     string
		isError        bool
		expectedResult []string
	}{
		{
			name:           "test simple OR",
			filterSpec:     "attributes.foo=2 OR attributes.bar=2",
			expectedResult: []string{"attributes.foo=2", "attributes.bar=2"},
		},
		{
			name:           "test different cases",
			filterSpec:     `attributes.foo=2 OR attributes.bar=' stuff or 245677 some \'other\' stuff' oR foo.bar=single or bar.baz=popmo Or pop=push_1`,
			expectedResult: []string{"attributes.foo=2", "attributes.bar= stuff or 245677 some 'other' stuff", "foo.bar=single", "bar.baz=popmo", "pop=push_1"},
		},
		{
			name:           "test quoted",
			filterSpec:     "attributes.foo=2 or attributes.bar='\\'something\\' or = \\'else\\''",
			expectedResult: []string{"attributes.foo=2", "attributes.bar='something' or = 'else'"},
		},
		{
			name:           "test bad quotes",
			filterSpec:     "attributes.foo=2 or attributes.bar=something or = else",
			expectedResult: []string{},
			isError:        true,
		},
		{
			name:           "test bad filter",
			filterSpec:     "attributes.foo=2 or attributes.bar=something or = else",
			expectedResult: []string{},
			isError:        true,
		},
		{
			name:           "test single term",
			filterSpec:     "attributes.bar='something or = else'",
			expectedResult: []string{"attributes.bar=something or = else"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			result, err := getFilterParts(test.filterSpec)

			if !test.isError {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedResult, result)
			} else {
				assert.NotNil(t, err)
			}
		})
	}

}

func TestFindValues(t *testing.T) {
	logger.New("NOOP")
	tests := []struct {
		name           string
		values         url.Values
		origName       string
		ok             bool
		expectedResult map[string]string
	}{
		{
			name:           "test string attribtues",
			origName:       "attributes",
			values:         map[string][]string{"attributes.foo": {"bar", "baz"}, "attributes.bar": {"2"}},
			ok:             true,
			expectedResult: map[string]string{"foo": "bar", "bar": "2"},
		},
		{
			name:           "test dict and list",
			origName:       "attributes",
			values:         map[string][]string{"attributes.foo.bar": {"bar", "baz"}, "attributes.baz[]": {"44"}},
			ok:             true,
			expectedResult: map[string]string{"foo.bar": "bar", "baz[]": "44"},
		},
		{
			name:           "test no attributes",
			origName:       "attributes",
			values:         map[string][]string{"foo.bar": {"bar", "baz"}, "baz[]": {"44"}},
			ok:             false,
			expectedResult: map[string]string{},
		},
		{
			name:           "test only specified attrs selected",
			origName:       "boom",
			values:         map[string][]string{"boom.foo.bar": {"bap", "baz"}, "attributes.baz[]": {"44"}, "foo.bar": {"bar", "baz"}, "baz[]": {"44"}},
			ok:             true,
			expectedResult: map[string]string{"foo.bar": "bap"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			result, ok := findAttributesValues(test.values, test.origName)

			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.expectedResult, result)

		})
	}

}

func TestParseWithAssets(t *testing.T) {
	logger.New("NOOP")
	tests := []struct {
		name           string
		values         url.Values
		isError        bool
		expectedResult *assets.ListAssetsRequest
	}{
		{
			name:   "test simple attributes",
			values: map[string][]string{"attributes.foo.bar": {"bar", "baz"}, "attributes.baz[]": {"44"}},
			expectedResult: &assets.ListAssetsRequest{
				Attributes: map[string]string{"foo.bar": "bar", "baz[]": "44"},
			},
		},
		{
			name:   "test simple prop",
			values: map[string][]string{"chain_id": {"88"}, "attributes.baz[]": {"44"}},
			expectedResult: &assets.ListAssetsRequest{
				Attributes: map[string]string{"baz[]": "44"},
				ChainId:    "88",
			},
		},
		{
			name:           "test no query",
			values:         map[string][]string{},
			expectedResult: &assets.ListAssetsRequest{},
		},
		{
			name:           "test invalid props query",
			values:         map[string][]string{"foobar": {"88"}},
			isError:        true,
			expectedResult: &assets.ListAssetsRequest{},
		},
		{
			name: "test attributes and filtesrs",
			values: map[string][]string{
				"attributes.foo.bar": {"bar", "baz"},
				"attributes.baz[]":   {"44"},
				"filters":            {"attributes.foo=80 OR attribtues.foo=22"},
			},
			expectedResult: &assets.ListAssetsRequest{
				Attributes: map[string]string{"baz[]": "44", "foo.bar": "bar"},
				Filters: []*filter.Filter{
					{
						Or: []string{"attributes.foo=80", "attribtues.foo=22"},
					},
				},
			},
		},
		{
			name: "test bad filtesrs",
			values: map[string][]string{
				"attributes.foo.bar": {"bar", "baz"},
				"attributes.baz[]":   {"44"},
				"filters":            {"attributes.foo=80 AND attribtues.foo=22"},
			},
			isError:        true,
			expectedResult: nil,
		},
		{
			name: "test bad attributes",
			values: map[string][]string{
				"attributes.foo.bar.": {"bar", "baz"},
				"attributes.baz[]":    {"44"},
				"filters":             {"attributes.foo=80 OR attribtues.foo=22"},
			},
			isError:        true,
			expectedResult: nil,
		},
		{
			name: "test bad attributes double dot",
			values: map[string][]string{
				"attributes.foo..bar": {"bar", "baz"},
				"attributes.baz[]":    {"44"},
				"filters":             {"attributes.foo=80 OR attribtues.foo=22"},
			},
			isError:        true,
			expectedResult: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			message := &assets.ListAssetsRequest{}
			parser := NewQueryParserForComplexAttributes()
			err := parser.Parse(message, test.values, nil)

			if test.isError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResult.Attributes, message.Attributes)
			assert.Equal(t, test.expectedResult.ChainId, message.ChainId)
			assert.Equal(t, len(test.expectedResult.Filters), len(message.Filters))
			if len(test.expectedResult.Filters) > 0 {
				assert.Equal(t, test.expectedResult.Filters[0].Or, message.Filters[0].Or)
			}
		})
	}
}

func TestParseWithEvents(t *testing.T) {
	logger.New("NOOP")
	tests := []struct {
		name           string
		values         url.Values
		isError        bool
		expectedResult *assets.ListEventsRequest
	}{
		{
			name:   "test simple attributes",
			values: map[string][]string{"asset_attributes.foo.bar": {"bar", "baz"}, "asset_attributes.baz[]": {"44"}},
			expectedResult: &assets.ListEventsRequest{
				AssetAttributes: map[string]string{"foo.bar": "bar", "baz[]": "44"},
			},
		},
		{
			name:   "test simple prop",
			values: map[string][]string{"event_attributes.baz[]": {"44"}},
			expectedResult: &assets.ListEventsRequest{
				EventAttributes: map[string]string{"baz[]": "44"},
			},
		},
		{
			name:           "test no query",
			values:         map[string][]string{},
			expectedResult: &assets.ListEventsRequest{},
		},
		{
			name:           "test invalid props query",
			values:         map[string][]string{"foobar": {"88"}},
			isError:        true,
			expectedResult: &assets.ListEventsRequest{},
		},
		{
			name: "test attributes and filtesrs",
			values: map[string][]string{
				"asset.attributes.foo.bar": {"bar", "baz"},
				"asset.attributes.baz[]":   {"44"},
				"filters":                  {"attributes.foo=80 OR attribtues.foo=22"},
			},
			expectedResult: &assets.ListEventsRequest{
				Asset: &assets.ListEventsRequest_AssetQuery{
					Attributes: map[string]string{"baz[]": "44", "foo.bar": "bar"},
				},
				Filters: []*filter.Filter{
					{
						Or: []string{"attributes.foo=80", "attribtues.foo=22"},
					},
				},
			},
		},
		{
			name: "test bad filtesrs",
			values: map[string][]string{
				"filters": {"attributes.foo=80 AND attribtues.foo=22"},
			},
			isError:        true,
			expectedResult: nil,
		},
		{
			name: "test bad attributes",
			values: map[string][]string{
				"asset_attributes.foo.bar.": {"bar", "baz"},
				"asset_attributes.baz[]":    {"44"},
			},
			isError:        true,
			expectedResult: nil,
		},
		{
			name: "test bad attributes double dot",
			values: map[string][]string{
				"event_attributes.foo..bar": {"bar", "baz"},
				"event_attributes.baz[]":    {"44"},
			},
			isError:        true,
			expectedResult: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			message := &assets.ListEventsRequest{}
			parser := NewQueryParserForComplexAttributes()
			err := parser.Parse(message, test.values, nil)

			if test.isError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResult.AssetAttributes, message.AssetAttributes)
			assert.Equal(t, test.expectedResult.EventAttributes, message.EventAttributes)
			if test.expectedResult.Asset != nil {
				assert.Equal(t, test.expectedResult.Asset.Attributes, message.Asset.Attributes)
			}
			assert.Equal(t, len(test.expectedResult.Filters), len(message.Filters))
			if len(test.expectedResult.Filters) > 0 {
				assert.Equal(t, test.expectedResult.Filters[0].Or, message.Filters[0].Or)
			}
		})
	}
}
