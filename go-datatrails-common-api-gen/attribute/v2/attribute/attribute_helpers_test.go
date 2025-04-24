package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeIsEqual(t *testing.T) {

	table := []struct {
		name     string
		a        *Attribute
		b        *Attribute
		expected bool
	}{
		{
			"string equal",
			NewStringAttribute("a"),
			NewStringAttribute("a"),
			true,
		},
		{
			"string not equal",
			NewStringAttribute("a"),
			NewStringAttribute("b"),
			false,
		},
		{
			"nil equal", nil, nil, true,
		},
		{
			"dict empty equal",
			NewDictAttribute(map[string]string{}),
			NewDictAttribute(map[string]string{}),
			true,
		},
		{
			"dict equal",
			NewDictAttribute(map[string]string{"a": "1", "b": "2"}),
			NewDictAttribute(map[string]string{"a": "1", "b": "2"}),
			true,
		},
		{
			"dict not equal single value",
			NewDictAttribute(map[string]string{"a": "1", "b": "2"}),
			NewDictAttribute(map[string]string{"a": "1", "b": "3"}),
			false,
		},
		{
			"dict not equal missing key",
			NewDictAttribute(map[string]string{"a": "1", "b": "2"}),
			NewDictAttribute(map[string]string{"a": "1"}),
			false,
		},
		{
			"dict not equal extra key",
			NewDictAttribute(map[string]string{"a": "1", "b": "2"}),
			NewDictAttribute(map[string]string{"a": "1", "b": "2", "c": "3"}),
			false,
		},

		{
			"list equal",
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "4"},
			}),
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "4"},
			}),
			true,
		},
		{
			"list equal empty",
			NewListAttribute([]map[string]string{}),
			NewListAttribute([]map[string]string{}),
			true,
		},
		{
			"list equal empty items",
			NewListAttribute([]map[string]string{
				{},
				{},
			}),
			NewListAttribute([]map[string]string{
				{},
				{},
			}),
			true,
		},

		{
			"list not equal single value",
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "4"},
			}),
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "5"},
			}),
			false,
		},
		{
			"list not equal extra value",
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "4"},
			}),
			NewListAttribute([]map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "5", "e": "6"},
			}),
			false,
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {

			actual := test.a.IsEqual(test.b)
			assert.Equal(
				t, actual, test.expected,
				"%s IsEqual result not as expected. ", test.name)
		})
	}
}
