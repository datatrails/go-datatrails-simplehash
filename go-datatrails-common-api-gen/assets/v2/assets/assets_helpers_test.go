package assets

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublicBoolToPrivacy(t *testing.T) {
	tests := []struct {
		name            string
		publicBool      bool
		expectedPrivacy Privacy
	}{
		{
			name:            "true should translate to Privacy_PUBLIC",
			publicBool:      true,
			expectedPrivacy: Privacy_PUBLIC,
		},
		{
			name:            "false should translate to Privacy_RESTRICTED",
			publicBool:      false,
			expectedPrivacy: Privacy_RESTRICTED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PublicBoolToPrivacy(test.publicBool)
			assert.Equal(t, test.expectedPrivacy, result)
		})
	}
}

func TestPrivacyToPublicBool(t *testing.T) {
	tests := []struct {
		name               string
		privacy            Privacy
		expectedPublicBool bool
	}{
		{
			name:               "Privacy_PUBLIC should translate to true",
			privacy:            Privacy_PUBLIC,
			expectedPublicBool: true,
		},
		{
			name:               "Privacy_RESTRICTED should translate to false",
			privacy:            Privacy_RESTRICTED,
			expectedPublicBool: false,
		},
		{
			name:               "Privacy_UNSPECIFIED should translate to false",
			privacy:            Privacy_PRIVACY_UNSPECIFIED,
			expectedPublicBool: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PrivacyToPublicBool(test.privacy)
			assert.Equal(t, test.expectedPublicBool, result)
		})
	}
}

func TestPrivacyStringToPublicBool(t *testing.T) {
	tests := []struct {
		name               string
		privacyString      string
		expectedPublicBool bool
		expectedError      error
	}{
		{
			name:               "PUBLIC should translate to true",
			privacyString:      "PUBLIC",
			expectedPublicBool: true,
			expectedError:      nil,
		},
		{
			name:               "RESTRICTED should translate to false",
			privacyString:      "RESTRICTED",
			expectedPublicBool: false,
			expectedError:      nil,
		},
		{
			name:               "private should return an error (case sensitive)",
			privacyString:      "private",
			expectedPublicBool: false,
			expectedError:      fmt.Errorf("illegal Privacy 'private'"),
		},
		{
			name:               "public should return an error (case sensitive)",
			privacyString:      "public",
			expectedPublicBool: false,
			expectedError:      fmt.Errorf("illegal Privacy 'public'"),
		},
		{
			name:               "nonsense string should return an error",
			privacyString:      "foobar",
			expectedPublicBool: false,
			expectedError:      fmt.Errorf("illegal Privacy 'foobar'"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := PrivacyStringToPublicBool(test.privacyString)

			assert.Equal(t, test.expectedPublicBool, result)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestChainIDFromString(t *testing.T) {
	type args struct {
		chainID string
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name:    "one",
			args:    args{chainID: "1"},
			want:    big.NewInt(1),
			wantErr: false,
		},
		{
			name:    "negative one",
			args:    args{chainID: "-1"},
			want:    nil,
			wantErr: true,
		},
		// Note: academically the value is unlimited. technically it is bounded
		// by the evm at 256 bit. There was an aspirational EIP to limit it
		// within 64 bits. We get to chose so we take the lower limit.
		{
			name:    "on the limit",
			args:    args{chainID: "9223372036854775771"},
			want:    maxChainID,
			wantErr: false,
		},
		{
			name:    "one over the limit",
			args:    args{chainID: "9223372036854775772"},
			want:    nil,
			wantErr: true,
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChainIDFromString(tt.args.chainID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChainIDFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChainIDFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
