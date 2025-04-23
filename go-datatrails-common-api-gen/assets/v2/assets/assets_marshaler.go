package assets

import (
	"reflect"

	"github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/marshalers/simpleoneof"
)

// NewFlatMarshalerForEvents creates marshaler configured to handle assets/events
func NewFlatMarshalerForEvents() *simpleoneof.Marshaler {
	return simpleoneof.NewFlatMarshaler(
		[]reflect.Type{
			reflect.TypeOf(AssetResponse{}),
			reflect.TypeOf(EventResponse{}),
		},
		[]reflect.Type{
			reflect.TypeOf(ListAssetsResponse{}),
			reflect.TypeOf(ListEventsResponse{}),
		},
	)
}

// NewFlatMarshalerForAssets creates marshaler configured to handle assets/events
func NewFlatMarshalerForAssets(ignoreFields []string) *simpleoneof.Marshaler {
	return simpleoneof.NewFlatMarshalerForAssetsV2(
		[]reflect.Type{
			reflect.TypeOf(AssetResponse{}),
			reflect.TypeOf(EventResponse{}),
		},
		[]reflect.Type{
			reflect.TypeOf(ListAssetsResponse{}),
			reflect.TypeOf(ListEventsResponse{}),
		},
		ignoreFields,
	)
}

// NewFlatMarshalerForNotifications create marshaler configured to handle event message - we send it to UI
func NewFlatMarshalerForNotifications() *simpleoneof.Marshaler {
	return simpleoneof.NewFlatMarshaler(
		[]reflect.Type{
			reflect.TypeOf(EventMessage{}),
		},
		[]reflect.Type{},
	)
}
