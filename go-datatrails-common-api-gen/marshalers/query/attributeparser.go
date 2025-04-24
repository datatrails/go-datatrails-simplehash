package query

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	dot = "."
)

type propertyHandler func(protoreflect.ProtoMessage, runtime.QueryParameterParser, protoreflect.Value, protoreflect.FieldDescriptor, url.Values, *utilities.DoubleArray) error

// NewQueryParserForComplexAttributes returns query parser set up to
// use custom parsing for attributes.*
func NewQueryParserForComplexAttributes() runtime.QueryParameterParser {
	return &paramQueryParser{
		customHandlingFor: map[string]propertyHandler{
			"event_attributes": handleComplexAttributes,
			"asset_attributes": handleComplexAttributes,
			"attributes":       handleComplexAttributes,
			"asset":            newRecursiveHandler("asset"),
			"filters":          handleFilters,
		},
	}
}

type paramQueryParser struct {
	customHandlingFor map[string]propertyHandler
}

func (p *paramQueryParser) Parse(msg proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	// get the value of msg
	if msg == nil {
		return errors.New("nil msg")
	}

	msgVal := reflect.ValueOf(msg)
	if !msgVal.IsNil() && msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	if msgVal.Kind() == reflect.Struct {

		fields := msg.ProtoReflect().Descriptor().Fields()

		for i := 0; i < fields.Len(); i++ {
			desc := fields.Get(i)
			val := msg.ProtoReflect().Get(desc)
			// check if we have custom handler defined for a property - if so use it
			if handler, ok := p.customHandlingFor[string(desc.Name())]; ok {
				err := handler(msg, p, val, desc, values, filter)
				if err != nil {
					logger.Sugar.Infow("err setting from custom handler", "name", desc.FullName(), "err", err)
					return err
				}
				continue
			}

			// no custom handling - use runtime.PopulateFieldFromPath
			for path, v := range allWithPrefix(values, string(desc.Name())) {
				for _, vv := range v {
					err := runtime.PopulateFieldFromPath(msg, path, vv)
					if err != nil {
						logger.Sugar.Infow("err setting", "path", path, "vv", vv, "err", err)
						return err
					}
				}
			}
		}

		// do validation now
		// get all valid param names
		vals := p.validValues(msg)

		// re matching words separated with dots
		// no consecutive dots and no dot at the end
		re, err := regexp.Compile(`^([^\.]+\.)*[^\.]+$`)
		if err != nil {
			logger.Sugar.Infof("parsing query failed: %v", err)
			return fmt.Errorf("parsing query failed")
		}

		for val := range values {
			logger.Sugar.Debugf("val: %v", val)
			// check for '.'s
			if !re.MatchString(val) {
				logger.Sugar.Debugf("invalid query parameter does not match: %s", val)
				return fmt.Errorf("invalid query parameter does not match: %s", val)
			}

			// if custom handled skip it
			if p.isCustomProp(val) {
				continue
			}

			// check if we have a slot on proto
			if ok := vals[val]; !ok {
				logger.Sugar.Debugf("invalid query parameter not in proto: %s", val)
				return fmt.Errorf("invalid query parameter not in proto: %s", val)
			}
		}
	}
	return nil
}

// check if the field is the one we handle in special way
// aka: ComplexAttributes
func (p *paramQueryParser) isCustomProp(name string) bool {
	path := strings.Split(name, dot)
	if len(path) < 1 {
		return false
	}

	_, ok := p.customHandlingFor[path[0]]
	return ok
}

// walk the proto message
// and get all possible valid attributes we can ask for
func (p *paramQueryParser) validValues(msg proto.Message) map[string]bool {
	values := map[string]bool{}
	if msg == nil {
		return values
	}
	// staticcheck shows this is never used
	//msgVal := reflect.ValueOf(msg)
	//if !msgVal.IsNil() && msgVal.Kind() == reflect.Ptr {
	//msgVal = msgVal.Elem()
	//}

	fields := msg.ProtoReflect().Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		desc := fields.Get(i)
		prefix := string(desc.FullName().Name())
		// if it's the custom field, skip it
		if p.isCustomProp(prefix) {
			continue
		}

		val := msg.ProtoReflect().Get(desc)

		// !IsList excludes repeated fields
		if desc.Kind() == protoreflect.MessageKind && !desc.IsList() {
			v := p.validValues(val.Message().Interface())
			for k := range v {
				values[prefix+dot+k] = true
				// add this so that a query can specify the base prefix as well as any
				// subfields. This is done so that timestamps can work sensibly without needing
				// to explicitly specify '.seconds'
				if k == "seconds" {
					values[prefix] = true
				}
			}

			continue
		}

		// got something else - just accept as is
		// as we can't infer much more in this code
		values[prefix] = true
	}

	return values
}

func newRecursiveHandler(attribute string) propertyHandler {
	f := func(msg protoreflect.ProtoMessage, p runtime.QueryParameterParser, customField protoreflect.Value, prop protoreflect.FieldDescriptor, values url.Values, filter *utilities.DoubleArray) error {

		// we get the kind here - we expect to get a pointer to protomsg here
		kind := prop.Kind()
		if prop.Kind() == protoreflect.MessageKind {

			// create the new instance of our object
			newVal := customField.Message().New()
			msg.ProtoReflect().Set(prop, protoreflect.ValueOf(newVal))

			// collect all values with our prefix
			nv, found := findAttributesValues(values, attribute)
			if !found {
				return nil
			}

			// rewrite into new url.Values
			newValues := map[string][]string{}
			for k, v := range nv {
				newValues[k] = []string{v}
			}

			return p.Parse(newVal.Interface(), newValues, filter)
		}

		return fmt.Errorf("unhandled attribute type when filtering %v", kind)
	}
	return f
}

func handleComplexAttributes(msg protoreflect.ProtoMessage, p runtime.QueryParameterParser, customField protoreflect.Value, prop protoreflect.FieldDescriptor, values url.Values, filter *utilities.DoubleArray) error {
	if v, ok := findAttributesValues(values, string(prop.Name())); ok {
		mv := msg.ProtoReflect().Mutable(prop)
		for vv, kk := range v {
			mv.Map().Set(protoreflect.MapKey(protoreflect.ValueOf(vv)), protoreflect.ValueOf(kk))
		}
	}

	return nil
}

func handleFilters(msg protoreflect.ProtoMessage, p runtime.QueryParameterParser, customField protoreflect.Value, prop protoreflect.FieldDescriptor, values url.Values, filter *utilities.DoubleArray) error {
	// have we even got a valid msg?
	if !msg.ProtoReflect().IsValid() {
		return nil
	}

	// did user specify any filters in the query string
	if values.Has(string(prop.Name())) {

		// create mutable 'filters' property
		filtersMutable := msg.ProtoReflect().Mutable(prop)

		// get the filters from query
		//filterSpec := values.Get(string(prop.Name()))
		filterSpecs := values[string(prop.Name())]
		for _, filterSpec := range filterSpecs {

			// split the filter terms
			filters, err := getFilterParts(filterSpec)
			if err != nil {
				return err
			}

			// append new filter to the list
			newFilter := filtersMutable.List().NewElement()
			// get the descriptor for the 'or' element of a filter
			orDesc := newFilter.Message().Descriptor().Fields().ByJSONName("or")
			// create mutable 'or' proprty of a filter
			or := newFilter.Message().Mutable(orDesc)

			// append our terms eg. attributes.foo=5 to the 'or'
			for _, term := range filters {
				v := protoreflect.ValueOf(term)
				or.List().Append(v)
			}
			// append our new filter to 'filters'
			filtersMutable.List().Append(newFilter)
		}
	}

	return nil
}

func getFilterParts(spec string) ([]string, error) {
	filters := []string{}

	// greedy match for stuff surrounded by single quotes
	matchQuote := `'((?:[^'\\]|\\.)*)'`
	q, err := regexp.Compile(matchQuote)
	if err != nil {
		return filters, err
	}

	// match fiter-ish thing - some attribute equals to a value in quotes with spaces
	// or continous string
	matchTerm := `(\S+?)=(('((?:[^'\\]|\\.)*)')|(\S+?))(\s|$)`
	r, err := regexp.Compile(matchTerm)
	if err != nil {
		return filters, err
	}
	filterSpec := spec
	//repeast until we consume the whole filter
	for len(filterSpec) > 0 {
		// if we matched filter like expression process it
		if loc := r.FindStringSubmatchIndex(filterSpec); len(loc) > 3 && loc[0] == 0 {
			term := r.FindStringSubmatch(filterSpec)
			// trim excessive unquoted spaces and trim the part we're processing
			filterSpec = strings.TrimSpace(strings.TrimPrefix(filterSpec, term[0]))
			// trim outer quotes if they are present and replace the escape sequences
			t := q.ReplaceAllString(term[2], "$1")
			t = strings.ReplaceAll(t, "\\'", "'")
			// append our new term to filters list
			filters = append(filters, fmt.Sprintf("%s=%s", strings.TrimSpace(term[1]), t))
			// check if we have anything left
			if len(filterSpec) > 0 {
				// if we have an or we remove it, otherwise we fail
				// as we have some additonal unexpected text
				if !strings.HasPrefix(strings.ToLower(filterSpec), "or ") {
					return filters, fmt.Errorf("malformed filter")
				}
				filterSpec = filterSpec[3:]
			}
		} else {
			// if we didn't match filter like thing we error out
			// as the filter is not valid
			return filters, fmt.Errorf("malformed filter")
		}
	}

	return filters, nil
}

// find all value that have prefix of origName. (eg. principal_declared.issuer)
func allWithPrefix(values url.Values, origName string) url.Values {
	ret := url.Values{}
	for k, v := range values {
		if strings.HasPrefix(k, origName+".") || k == origName {
			ret[k] = v
		}
	}
	return ret
}

// this is specific for complex attributes - it sprinkles listval/value/dictval and strval
// in between the . separated path user provided in request
// . are assumed to reference deeper structure and [] signifies that
// following . separated key has a value in any element in the list
// examples:
//
//	attributes.info_list[].b_key=b value  -> info_list.value.listval.value.[].value.b_key: b value
//	attributes.info_dict.id=1234          -> info_dict.value.dictval.value.id: 1234
//	attributes.prop=val                   -> prop.value.strval: val
//	we strip attributes as we fill in the attributes map on the message
func findAttributesValues(values url.Values, origName string) (map[string]string, bool) {
	data := map[string]string{}
	ok := false

	dottedPrefix := origName + dot
	for k := range values {
		if strings.HasPrefix(k, dottedPrefix) {
			ok = true
			// remove the "attributes." prefix
			attrKeyTrimmed := strings.TrimPrefix(k, dottedPrefix)
			data[attrKeyTrimmed] = values.Get(k)

		}
	}
	return data, ok
}
