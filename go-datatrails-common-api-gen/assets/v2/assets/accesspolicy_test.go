package assets

import (
	"fmt"
	"testing"

	v2attribute "github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/attribute/v2/attribute"
)

// TestAssetResponse_IsAssetSharedForWallet
// covers the case where the only policy is a wild share and so the whole asset is shared
func TestAssetResponse_IsAssetSharedForWallet(t *testing.T) {

	type fields struct {
		AccessPolicy map[string]*v2attribute.Attribute
	}
	type args struct {
		policyKey string
		wallet    string
		attribute string
	}

	Wallet1 := "0xWALLET1"
	Wallet2 := "0xWALLET2"
	WalletWild := "0xWILD-WALLET"

	TesseraPubWild := "b64-TESSERAPUB-WILD"
	tractor_colour := "tractor_colour"
	engine_size := "engine_size"

	twoAllRead := map[string]*v2attribute.Attribute{}

	policyAddAssetAttributeReaderOrFailNow(
		t, twoAllRead, "*", WalletWild, TesseraPubWild)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should match wildwallet in policy sharing all asset attributes to that wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    WalletWild,
				attribute: tractor_colour,
			},
			want: true,
		},
		{
			name: "Should match wildwallet in policy sharing all asset attributes to that wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    WalletWild,
				attribute: engine_size,
			},
			want: true,
		},

		{
			name: "Should NOT match wallet1 in policy sharing all asset attributes to single different wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet1,
				attribute: tractor_colour,
			},
			want: false,
		},

		{
			name: "Should NOT match wallet2 in policy sharing all asset attributes to single wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet2,
				attribute: tractor_colour,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AssetResponse{
				AccessPolicy: tt.fields.AccessPolicy,
			}
			got, err := a.IsSharedForWallet(tt.args.policyKey, tt.args.wallet, tt.args.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetResponse.IsSharedForWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AssetResponse.IsSharedForWallet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAssetResponse_AreEventsSharedForWallet
// covers the case where the only policy is a wild share on the event type and so the whole asset event history is shared
func TestAssetResponse_AreEventsSharedForWallet(t *testing.T) {

	type fields struct {
		AccessPolicy map[string]*v2attribute.Attribute
	}
	type args struct {
		policyKey string
		wallet    string
		attribute string
	}

	Wallet1 := "0xWALLET1"
	Wallet2 := "0xWALLET2"
	WalletWild := "0xWILD-WALLET"

	TesseraPubWild := "b64-TESSERAPUB-WILD"
	tractor_colour := "tractor_colour"
	engine_size := "engine_size"

	twoAllRead := map[string]*v2attribute.Attribute{}

	policyAddEventDisplayTypeReaderOrFailNow(
		t, twoAllRead, "*", WalletWild, TesseraPubWild)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should match wildwallet in policy sharing all asset attributes to that wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: EventDisplayTypeRead,
				wallet:    WalletWild,
				attribute: tractor_colour,
			},
			want: true,
		},
		{
			name: "Should match wildwallet in policy sharing all asset attributes to that wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: EventDisplayTypeRead,
				wallet:    WalletWild,
				attribute: engine_size,
			},
			want: true,
		},

		{
			name: "Should match wallet1 in policy sharing all asset attributes to single wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: EventDisplayTypeRead,
				wallet:    Wallet1,
				attribute: tractor_colour,
			},
			want: false,
		},

		{
			name: "Should NOT match wallet2 in policy sharing all asset attributes to single wallet",
			fields: fields{
				AccessPolicy: twoAllRead,
			},
			args: args{
				policyKey: EventDisplayTypeRead,
				wallet:    Wallet2,
				attribute: tractor_colour,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AssetResponse{
				AccessPolicy: tt.fields.AccessPolicy,
			}
			got, err := a.IsSharedForWallet(tt.args.policyKey, tt.args.wallet, tt.args.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetResponse.IsSharedForWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AssetResponse.IsSharedForWallet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAssetResponse_IsSharedForWallet_AssetAttributesReader
// Covers various plain read cases (no wild cards in policy, all shares are read shares)
func TestAssetResponse_IsSharedForWallet_AssetAttributesReader(t *testing.T) {

	type fields struct {
		AccessPolicy map[string]*v2attribute.Attribute
	}
	type args struct {
		policyKey string
		wallet    string
		attribute string
	}

	Wallet1 := "0xWALLET1"
	Wallet2 := "0xWALLET2"

	TesseraPub1 := "b64-TESSERAPUB1"
	TesseraPub2 := "b64-TESSERAPUB2"
	tractor_colour := "tractor_colour"
	engine_size := "engine_size"

	twoReadSharedAssetAttributes := map[string]*v2attribute.Attribute{}

	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributes, tractor_colour, Wallet1, TesseraPub1)
	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributes, engine_size, Wallet2, TesseraPub2)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should match wallet1 in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributes,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet1,
				attribute: tractor_colour,
			},
			want: true,
		},

		{
			name: "Should NOT match wallet2 in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributes,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet2,
				attribute: tractor_colour,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AssetResponse{
				AccessPolicy: tt.fields.AccessPolicy,
			}
			got, err := a.IsSharedForWallet(tt.args.policyKey, tt.args.wallet, tt.args.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetResponse.IsSharedForWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AssetResponse.IsSharedForWallet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAssetResponse_IsSharedForWallet_AssetAttributesMixedWildeReader
// Covers mixed read cases (wild card present in policy, all shares are read shares)
func TestAssetResponse_IsSharedForWallet_AssetAttributesMixedWildeReader(t *testing.T) {

	type fields struct {
		AccessPolicy map[string]*v2attribute.Attribute
	}
	type args struct {
		policyKey string
		wallet    string
		attribute string
	}

	Wallet1 := "0xWALLET1"
	Wallet2 := "0xWALLET2"
	WalletWild := "0xWILD-WALLET"

	TesseraPub1 := "b64-TESSERAPUB1"
	TesseraPub2 := "b64-TESSERAPUB2"
	TesseraPubWild := "b64-TESSERAPUB-WILD"
	tractor_colour := "tractor_colour"
	engine_size := "engine_size"

	twoReadSharedAssetAttributesAndOneWildWallet := map[string]*v2attribute.Attribute{}

	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, tractor_colour, Wallet1, TesseraPub1)
	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, engine_size, Wallet2, TesseraPub2)
	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, "*", WalletWild, TesseraPubWild)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should match wildwallet in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    WalletWild,
				attribute: tractor_colour,
			},
			want: true,
		},
		{
			name: "Should match wildwallet in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    WalletWild,
				attribute: engine_size,
			},
			want: true,
		},

		{
			name: "Should match wallet1 in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet1,
				attribute: tractor_colour,
			},
			want: true,
		},

		{
			name: "Should NOT match wallet2 in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				policyKey: AssetAttributesRead,
				wallet:    Wallet2,
				attribute: tractor_colour,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AssetResponse{
				AccessPolicy: tt.fields.AccessPolicy,
			}
			got, err := a.IsSharedForWallet(tt.args.policyKey, tt.args.wallet, tt.args.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetResponse.IsSharedForWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AssetResponse.IsSharedForWallet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAssetResponse_IsPubliclyAttested_AssetAttributesMixedWildeReader
// Covers mixed read cases (wild card present in policy, all shares are read shares)
func TestAssetResponse_IsPubliclyAttested_AssetAttributesMixedWildeReader(t *testing.T) {

	type fields struct {
		AccessPolicy map[string]*v2attribute.Attribute
	}
	type args struct {
		policyKey string
		wallet    string
		attribute string
	}

	Wallet1 := "0xWALLET1"
	Wallet2 := "0xWALLET2"
	WalletWild := "0xWILD-WALLET"

	TesseraPub1 := "b64-TESSERAPUB1"
	TesseraPub2 := "b64-TESSERAPUB2"
	TesseraPubWild := "b64-TESSERAPUB-WILD"
	tractor_colour := "tractor_colour"
	engine_size := "engine_size"

	twoReadSharedAssetAttributesAndOneWildWallet := map[string]*v2attribute.Attribute{}

	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, tractor_colour, Wallet1, TesseraPub1)
	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, engine_size, Wallet2, TesseraPub2)
	policyAddAssetAttributeReaderOrFailNow(
		t, twoReadSharedAssetAttributesAndOneWildWallet, "*", WalletWild, TesseraPubWild)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should match wildwallet in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				wallet: WalletWild,
			},
			want: true,
		},
		{
			name: "Should match wildwallet in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				wallet: WalletWild,
			},
			want: true,
		},

		{
			name: "Should match wallet1 in policy sharing two asset attributes",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				wallet: Wallet1,
			},
			want: true,
		},

		{
			name: "Should match wallet2 in policy even partial share",
			fields: fields{
				AccessPolicy: twoReadSharedAssetAttributesAndOneWildWallet,
			},
			args: args{
				wallet: Wallet2,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AssetResponse{
				AccessPolicy: tt.fields.AccessPolicy,
			}
			got := a.IsPubliclyAttested(tt.args.wallet)
			if got != tt.want {
				t.Errorf("AssetResponse.IsPubliclyAttested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func policyAddAssetAttributeReaderOrFailNow(
	t *testing.T,
	policy map[string]*v2attribute.Attribute,
	attribute, wallet, tesseraPub string) {
	if err := policyAddAssetAttributeReader(policy, attribute, wallet, tesseraPub); err != nil {
		t.Fatalf("setting up policy failed: %v", err)
	}
}

func policyAddAssetAttributeWriterOrFailNow(
	t *testing.T,
	policy map[string]*v2attribute.Attribute,
	attribute, wallet, tesseraPub string) {
	if err := policyAddAssetAttributeWriter(policy, attribute, wallet, tesseraPub); err != nil {
		t.Fatalf("setting up policy failed: %v", err)
	}
}

func policyAddEventDisplayTypeReaderOrFailNow(
	t *testing.T,
	policy map[string]*v2attribute.Attribute,
	value, wallet, tesseraPub string) {
	if err := policyAddEventDisplayType(
		policy, value, wallet, tesseraPub, false); err != nil {
		t.Fatalf("setting up policy failed: %v", err)
	}
}

func policyAddEventDisplayTypeWriterOrFailNow(
	t *testing.T,
	policy map[string]*v2attribute.Attribute,
	value, wallet, tesseraPub string) {
	if err := policyAddEventDisplayType(
		policy, value, wallet, tesseraPub, true); err != nil {
		t.Fatalf("setting up policy failed: %v", err)
	}
}

func policyAddAssetAttributeReader(
	policy map[string]*v2attribute.Attribute,
	attribute, wallet, tesseraPub string) error {
	return policyAddAssetAttribute(policy, attribute, wallet, tesseraPub, false)
}

func policyAddAssetAttributeWriter(
	policy map[string]*v2attribute.Attribute,
	attribute, wallet, tesseraPub string) error {
	return policyAddAssetAttribute(policy, attribute, wallet, tesseraPub, false)
}

// listAttrAppend appends all the complex list entries from b to a. We *shallow*
// copy all items from b into a.
func listAttrAppend(a *v2attribute.Attribute, b *v2attribute.Attribute) error {
	la, ok := a.Value.(*v2attribute.Attribute_ListVal)
	if !ok {
		return fmt.Errorf("a must be a complex list")
	}
	lb, ok := b.Value.(*v2attribute.Attribute_ListVal)
	if !ok {
		return fmt.Errorf("b must be a complex list")
	}

	// Note: we could deal with appending dicts to complex lists here if we
	// wanted.  And we could do a deep copy of b into a. But neither of those
	// things are required for these tests.

	la.ListVal.Value = append(la.ListVal.Value, lb.ListVal.Value...)
	return nil
}

func policyAddAssetAttribute(
	policy map[string]*v2attribute.Attribute,
	attribute, wallet, tesseraPub string, writer bool) error {

	var policyKey string
	if writer {
		policyKey = AssetAttributesWrite
	} else {
		policyKey = AssetAttributesRead
	}

	attr := v2attribute.NewListAttribute([]map[string]string{
		{
			AttributeItem: attribute,
			wallet:        WalletValueMarker,
			tesseraPub:    TesseraPubValueMarker,
		},
	})

	var ok bool
	var policyAttr *v2attribute.Attribute
	// If the policy key doesn't exist, we just insert the new one
	if policyAttr, ok = policy[policyKey]; !ok {
		policy[policyKey] = attr
		return nil
	}
	if err := listAttrAppend(policyAttr, attr); err != nil {
		return err
	}

	return nil
}

func policyAddEventDisplayType(
	policy map[string]*v2attribute.Attribute,
	value, wallet, tesseraPub string, writer bool) error {

	var policyKey string
	if writer {
		policyKey = EventDisplayTypeWrite
	} else {
		policyKey = EventDisplayTypeRead
	}

	attrs, err := v2attribute.MakeAttributes(policyKey, []map[string]string{
		{
			ValueItem:  value,
			wallet:     WalletValueMarker,
			tesseraPub: TesseraPubValueMarker,
		},
	})
	if err != nil {
		return err
	}
	for a, attr := range attrs {
		policy[a] = attr
	}
	return nil

}
