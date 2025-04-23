package assets

import (
	"fmt"

	"github.com/datatrails/go-datatrails-common/logger"
)

// The policy format is documented in src/accesspolicy/doc.go

const (
	AssetAttributesRead   = "asset_attributes_read"
	AssetAttributesWrite  = "asset_attributes_write"
	EventDisplayTypeRead  = "event_arc_display_type_read"
	EventDisplayTypeWrite = "event_arc_display_type_write"
	WalletValueMarker     = "wallet"
	TesseraPubValueMarker = "tessera"
	AttributeItem         = "attribute"
	ValueItem             = "value"
	AttributeWildcard     = "*"
)

// AccessPolicyStringValue retreives the names value form an access policy
// attribute. if the attribute is missing the defaultValue is returned.
// otherwise the value is returned if it is a string. If the value exists and is
// not a string an error is returned.
func (a *AssetResponse) AccessPolicyStringValue(name, defaultValue string) (string, error) {
	var zeroValue string

	attr, ok := a.AccessPolicy[name]
	if !ok {
		return defaultValue, nil
	}
	value, ok := attr.GetString()
	if !ok {
		return zeroValue, fmt.Errorf("value of atribute '%s' is not a string", name)
	}
	return value, nil
}

// IsPubliclyAttested determines if an asset is publicly attested
// if public wallet address is present *anywhere* in the policy
// we assume the asset is public as there is no other way of getting
// public wallet to policy other than setting public: true on an asset
// previous method was to specific and did not work with partial attestation
func (a *AssetResponse) IsPubliclyAttested(publicWallet string) bool {

	for attrs := range a.AccessPolicy {
		data, ok := a.AccessPolicy[attrs].GetList()
		if !ok {
			continue
		}

		for _, v := range data {
			if vv, ok := v[publicWallet]; ok {
				if vv == "wallet" {
					return true
				}
			}
		}
	}

	return false
}

// IsSharedForWallet determines if the value identified by `attribute` is shared
// with the organisation identified by `wallet`. `policyKey` determines the kind
// of share: asset attribute or event attribute and whether it is a read share
// or a write share. The same attribute may be independently read or write
// shared, and may be independently shared when present on event attributes or
// asset attributes. Please use the appropriate constants defined in this
// package.
func (a *AssetResponse) IsSharedForWallet(policyKey, wallet, attribute string) (bool, error) {

	var attributeOrValue string

	switch policyKey {
	case AssetAttributesRead, AssetAttributesWrite:
		attributeOrValue = AttributeItem

	case EventDisplayTypeRead, EventDisplayTypeWrite:
		attributeOrValue = ValueItem
	default:
		return false, fmt.Errorf("unknown policy key: %s", policyKey)
	}

	// Note: The policy format on the asset response is documented in src/accesspolicy/doc.go
	attr, ok := a.AccessPolicy[policyKey]
	if !ok {
		logger.Sugar.DebugR("ap", a.AccessPolicy)
		return false, nil
	}

	readers, ok := attr.GetList()
	if !ok {
		return false, fmt.Errorf(
			"access policy bad format (not a list): %s", policyKey)
	}

	for _, share := range readers {

		// * If the policy share entry is a WILDCARD we check the supplied wallet.
		// * OR If the policy share entry matches the supplied attribute we check if
		//   the supplied wallet matches the share.
		//
		// Note that for event_arc_display_type_[read|write] we are testing share['value'] is wild or == attribute
		// And if it is asset_attribute_[read|write] we are testing share['attribute'] is wild or == attribute
		if attr, ok := share[attributeOrValue]; !ok || attr != attribute && attr != AttributeWildcard {
			continue
		}

		// Ok, the policy value or attribute entry is either a match to the
		// requested attribute or is wild. Check for wallet match.
		for maybeWallet, whatIsIt := range share {
			if whatIsIt == WalletValueMarker && maybeWallet == wallet {
				return true, nil
			}
		}
	}

	return false, nil
}
