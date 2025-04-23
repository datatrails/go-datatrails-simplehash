package attribute

import (
	fmt "fmt"
)

// Helpers to make dealing with Attribute less horrific (and error prone)
//
// For behaviours which deal with more than one type, we provide a uniform
// 'Get' method which returns the appropriate native golang type. The behaviour
// should switch on a type assertion to pick. Eg.,
//
//  attr is a Attribute
//
//	switch av := attr.Value.(type) {
//  case *v2assets.Attribute_ListVal:
//		v := av.Get()
//		// v is a []map[string]string
//
// Behaviours which require a single type should use the "bool getters".
// The bool is false if requested type isn't held in the Attribute.
//
// v, ok := attr.GetDict(); ok {
// 		// v is a map[string]string
// }
//
// Note: we don't deep copy the maps. The presumption is that we never want
// to modify the Attribute in place. We always consume it and
// transform it into something else

func NewStringAttribute(v string) *Attribute {
	return &Attribute{Value: &Attribute_StrVal{StrVal: v}}
}

func NewDictAttribute(v map[string]string) *Attribute {
	m := map[string]string{}
	for k, mv := range v {
		m[k] = mv
	}
	return &Attribute{Value: &Attribute_DictVal{DictVal: &DictAttr{Value: m}}}
}

func NewListAttribute(v []map[string]string) *Attribute {

	l := make([]*DictAttr, len(v))
	for i, m := range v {

		l[i] = &DictAttr{Value: map[string]string{}}
		// copy, rather than borrow
		for k, mv := range m {
			l[i].Value[k] = mv
		}
	}
	return &Attribute{Value: &Attribute_ListVal{ListVal: &ListAttr{Value: l}}}
}

// NewAttribute creates an attribute whose type is infered from a native go-lang
// type. string -> StrVal, map[string]string -> DictVal, []map[string]string -> ListVal
func NewAttribute(value any) (*Attribute, error) {
	switch v := value.(type) {
	case string:
		return NewStringAttribute(v), nil
	case map[string]string:
		return NewDictAttribute(v), nil
	case []map[string]string:
		return NewListAttribute(v), nil
	default:
		return nil, fmt.Errorf("value not string, map or list")
	}
}

// MakeAttributes accepts pairs of name, value and returns a
// map[string]*Attribute. value must be a string, a map[string]string or a
// []map[string]string
func MakeAttributes(namedAttributes ...any) (map[string]*Attribute, error) {

	if (len(namedAttributes) % 2) != 0 {
		return nil, fmt.Errorf("pairs of name, value required")
	}

	// empty named slice (no args) creates an empty but not nil map
	attributes := map[string]*Attribute{}

	for i := 0; i < len(namedAttributes); i += 2 {

		name, ok := namedAttributes[i].(string)
		if !ok {
			return nil, fmt.Errorf("name is not a string")
		}
		attr, err := NewAttribute(namedAttributes[i+1])
		if err != nil {
			return nil, err
		}
		attributes[name] = attr
	}
	return attributes, nil
}

// stringMapEqual tests the equivelance of a pair of map[string]strings'. We
// could use the reflection package deep equal but that seems needlessly
// wasteful of CPU.
func stringMapEqual(a map[string]string, b map[string]string) bool {

	if a == nil && b == nil {
		return true
	}
	if len(a) != len(b) {
		return false
	}

	var ok bool

	for ak := range a {

		var bv string
		if bv, ok = b[ak]; !ok {
			return false
		}
		if a[ak] != bv {
			return false
		}
	}

	return true
}

// IsEqual returns true if both values are equal. For dicts and lists we do a
// deep compare. For lists, the the order of items in both list must be the
// same.
func (m *Attribute) IsEqual(other *Attribute) bool {

	if m == nil && other == nil {
		return true
	}

	if m == nil || other == nil {
		return false
	}

	if v, ok := m.GetString(); ok {

		var ov string

		if ov, ok = other.GetString(); !ok {
			return false
		}
		return v == ov
	}

	if v, ok := m.GetDict(); ok {

		var ov map[string]string

		if ov, ok = other.GetDict(); !ok {
			return false
		}
		return stringMapEqual(v, ov)
	}

	if v, ok := m.GetList(); ok {

		var ov []map[string]string

		if ov, ok = other.GetList(); !ok {
			return false
		}

		if len(v) != len(ov) {
			return false
		}

		for i, vv := range v {

			if !stringMapEqual(vv, ov[i]) {
				return false
			}
		}

		return true
	}

	return false
}

// GetString returns the string if the value is the correct type.
// Otherwise the bool return value will be false.
func (m *Attribute) GetString() (string, bool) {

	// Note: GetStrVal is a particular anoyance and one motiviation for these
	// helpers, it returns "" if the type assertion fails. So callers have to
	// do type assertions anyway.
	if v, ok := m.Value.(*Attribute_StrVal); ok {
		return v.StrVal, true
	}
	return "", false
}

// Sets an attribute to a string value and returns true IF its a string type
// Otherwise returns false
func (m *Attribute) SetString(value string) bool {
	v, ok := m.Value.(*Attribute_StrVal)
	if !ok {
		return false
	}
	v.StrVal = value
	return true
}

// GetDict returns the map if the value is the correct type.  Otherwise the
// bool return value will be false.
func (m *Attribute) GetDict() (map[string]string, bool) {
	da, ok := m.Value.(*Attribute_DictVal)
	if !ok || da == nil {
		return nil, false
	}
	return da.DictVal.Value, true
}

// GetList returns the list of maps if the value is the correct type.
// Otherwise the bool return value will be false.
func (m *Attribute) GetList() ([]map[string]string, bool) {
	la, ok := m.Value.(*Attribute_ListVal)
	if !ok || la == nil {
		return nil, false
	}
	return la.Get(), true
}

func (sv *Attribute_StrVal) Get() string {
	return sv.StrVal
}

// GetDict returns the map if the value is the correct type.  Otherwise the
// bool return value will be false.
func (dv *Attribute_DictVal) Get() map[string]string {
	return dv.DictVal.Value
}

// GetList returns the list of maps if the value is the correct type.
// Otherwise the bool return value will be false.
func (lv *Attribute_ListVal) Get() []map[string]string {

	list := make([]map[string]string, len(lv.ListVal.Value))
	for i := range lv.ListVal.Value {
		// We expect that the protobuf serialisation means this case never
		// happens. nil => empty map is a perfectly reasonable transformation
		// if it ever does.
		iv := lv.ListVal.Value[i]
		m := map[string]string{}
		if iv != nil {
			m = iv.Value
		}
		list[i] = m
	}
	return list
}
