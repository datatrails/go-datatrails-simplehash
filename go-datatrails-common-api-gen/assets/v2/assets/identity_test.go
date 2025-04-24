package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEventUUIDFromIdentity tests:
//
// 1. we get the correct event uuid from the event identity
// 2. an incorrectly formatted identity returns the identity as given
func TestEventUUIDFromIdentity(t *testing.T) {
	type args struct {
		eventIdentity string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "positive",
			args: args{
				eventIdentity: assetIdentityPrefix + "9937fec4-95e9-4e90-9b80-5584d4ab7739" + eventIdentityPrefix + "1afa9f3b-cd1b-477d-b9b5-1448a8760adb",
			},
			expected: "1afa9f3b-cd1b-477d-b9b5-1448a8760adb",
		},
		{
			name: "bogus identity",
			args: args{
				eventIdentity: "foobarbaz",
			},
			expected: "foobarbaz",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := EventUuidFromIdentity(test.args.eventIdentity)

			assert.Equal(t, test.expected, actual)
		})
	}
}

// TestAssetUUIDFromEventIdentity tests:
//
// 1. we correctly get the asset uuid from the event identity
// 2. an incorrectly formatted identity returns the identity as given
func TestAssetUUIDFromEventIdentity(t *testing.T) {
	type args struct {
		eventIdentity string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "positive",
			args: args{
				eventIdentity: assetIdentityPrefix + "9937fec4-95e9-4e90-9b80-5584d4ab7739" + eventIdentityPrefix + "1afa9f3b-cd1b-477d-b9b5-1448a8760adb",
			},
			expected: "9937fec4-95e9-4e90-9b80-5584d4ab7739",
		},
		{
			name: "bogus identity",
			args: args{
				eventIdentity: "foobarbaz",
			},
			expected: "foobarbaz",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := AssetUuidFromEventIdentity(test.args.eventIdentity)

			assert.Equal(t, test.expected, actual)
		})
	}
}
