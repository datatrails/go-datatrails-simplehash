package simpleoneof

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/unittestapi"
	"github.com/stretchr/testify/assert"
)

// message Int32Test {
//   int32 integer_value = 1;
// }

func TestDecodeInt32(t *testing.T) {
	logger.New("NOOP")
	defer logger.OnExit()

	tests := []struct {
		name        string
		input       []byte
		shouldError bool
		shouldBe    int32
	}{
		{
			name:        "int32 value decodes okay",
			input:       []byte(`{"integer_value":123}`),
			shouldError: false,
			shouldBe:    123,
		},
		{
			name:        "zero int32 value decodes okay",
			input:       []byte(`{"integer_value":0}`),
			shouldError: false,
			shouldBe:    0,
		},
		{
			name:        "integer larger than int32 errors",
			input:       []byte(`{"integer_value":3000000000}`),
			shouldError: true,
		},
		{
			name:        "number with fractional component should error",
			input:       []byte(`{"integer_value":123.45}`),
			shouldError: true,
		},
		{
			name:        "number with zero fractional component should error",
			input:       []byte(`{"integer_value":123.0}`),
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputReader := bytes.NewReader(test.input)
			protoMsg := unittestapi.Int32Test{}

			marshaller := NewFlatMarshaler([]reflect.Type{}, []reflect.Type{})
			decoder := marshaller.NewDecoder(inputReader)

			err := decoder.Decode(&protoMsg)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.shouldBe, protoMsg.IntegerValue)
			}
		})
	}
}
