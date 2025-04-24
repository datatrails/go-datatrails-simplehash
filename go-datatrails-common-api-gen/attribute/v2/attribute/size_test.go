package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAttributeSize tests we can accurately get the size in bytes
// of a number of complex attributes.
func TestAttributesSize(t *testing.T) {
	tables := []struct {
		subtest string
		attrs   map[string]*Attribute

		expected int64
	}{
		{
			`size works for a map of different typed attributes`,

			map[string]*Attribute{
				"string attribute": {
					Value: &Attribute_StrVal{
						StrVal: "1",
					},
				},

				"list attribute": {
					Value: &Attribute_ListVal{
						ListVal: &ListAttr{
							Value: []*DictAttr{
								{
									Value: map[string]string{
										"2": "2",
										"3": "3",
									},
								},
							},
						},
					},
				},

				"dict attribute": {
					Value: &Attribute_DictVal{
						DictVal: &DictAttr{
							Value: map[string]string{
								"4":  "4",
								"5":  "5",
								"6":  "6",
								"Яя": "Яя", // use go runes that is 4 bytes long each
							},
						},
					},
				},
			},

			63,
		},
		{
			"count works for a nil map",

			nil,

			0,
		},
		{
			"count works for a nil attributes in map",

			map[string]*Attribute{
				"foo": nil,
				"bar": nil,
			},

			6, // for the keys
		},
	}

	for _, table := range tables {
		t.Run(table.subtest, func(t *testing.T) {
			actual := AttributesSize(table.attrs)
			assert.Equal(t, table.expected, actual)
		})
	}
}

// TestMakeUnion
func TestMakeUnion(t *testing.T) {
	tables := []struct {
		subtest  string
		attrsMap []map[string]*Attribute

		expected map[string]*Attribute
	}{
		{
			`one value in attributes maps`,

			[]map[string]*Attribute{
				{
					"string attribute": {
						Value: &Attribute_StrVal{
							StrVal: "1",
						},
					},

					"list attribute": {
						Value: &Attribute_ListVal{
							ListVal: &ListAttr{
								Value: []*DictAttr{
									{
										Value: map[string]string{
											"2": "2",
											"3": "3",
										},
									},
								},
							},
						},
					},

					"dict attribute": {
						Value: &Attribute_DictVal{
							DictVal: &DictAttr{
								Value: map[string]string{
									"4": "4",
									"5": "5",
									"6": "6",
								},
							},
						},
					},
				},
			},

			// expected
			map[string]*Attribute{
				"string attribute": {
					Value: &Attribute_StrVal{
						StrVal: "1",
					},
				},

				"list attribute": {
					Value: &Attribute_ListVal{
						ListVal: &ListAttr{
							Value: []*DictAttr{
								{
									Value: map[string]string{
										"2": "2",
										"3": "3",
									},
								},
							},
						},
					},
				},

				"dict attribute": {
					Value: &Attribute_DictVal{
						DictVal: &DictAttr{
							Value: map[string]string{
								"4": "4",
								"5": "5",
								"6": "6",
							},
						},
					},
				},
			},
		},
		{
			`multiple maps in attributes maps with at least one key the same,
			 but with different values.`,

			[]map[string]*Attribute{
				{
					"String1": {
						Value: &Attribute_StrVal{
							StrVal: "1",
						},
					},
					"String2": {
						Value: &Attribute_StrVal{
							StrVal: "2",
						},
					},
				},
				{
					// override String2 value
					"String2": {
						Value: &Attribute_StrVal{
							StrVal: "22",
						},
					},
					"String3": {
						Value: &Attribute_StrVal{
							StrVal: "3",
						},
					},
				},
			},

			// expected
			map[string]*Attribute{
				"String1": {
					Value: &Attribute_StrVal{
						StrVal: "1",
					},
				},
				"String2": {
					Value: &Attribute_StrVal{
						StrVal: "22",
					},
				},
				"String3": {
					Value: &Attribute_StrVal{
						StrVal: "3",
					},
				},
			},
		},
	}

	for _, table := range tables {
		t.Run(table.subtest, func(t *testing.T) {
			actual := MakeUnion(table.attrsMap...)
			assert.Equal(t, table.expected, actual)
		})
	}
}
