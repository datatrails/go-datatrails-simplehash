package assets

import (
	"fmt"
	"math"
	"math/big"
)

var (
	// EIP-2294 is the draft this limit came from. ChainID can be higher - We
	// just chose not to support larger than this for now.
	maxChainID = big.NewInt(math.MaxUint64/2 - 36)
)

// PublicBoolToPrivacy converts boolean to Privacy
func PublicBoolToPrivacy(b bool) Privacy {
	if b {
		return Privacy_PUBLIC
	}
	return Privacy_RESTRICTED
}

// PrivacyToPublicBool converts Privacy to bool
func PrivacyToPublicBool(b Privacy) bool {
	if b == Privacy_PUBLIC {
		return true
	}
	return false
}

// PrivacyStringToPublicBool converts a string repr of Privacy enum to bool.
func PrivacyStringToPublicBool(b string) (bool, error) {
	public, ok := Privacy_value[b]
	if !ok {
		return false, fmt.Errorf("illegal Privacy '%s'", b)
	}
	if Privacy(public) == Privacy_PUBLIC {
		return true, nil
	}
	return false, nil
}

// ChainIDFromString converts the string representation of a chain id to a big
// integer. Ensuring it is <= floor(MAX_UINT64/2) - 36
//
// We operate private consortia style networks and so can decide the range
// of chain ids we support for our networks. We choose to stick inside the
// uint64 representable range specified in the *stale* EIP-2294. We could in
// future relax this.
//
// History: EIP-155 which introduced replay protection,  does not specify an
// upper bound for chainId. The subsequent EIP-1344 for the ChainID opcode
// implicitly bounds it at 256bit (because that is the maximum the the evm
// allows for an atomic type).
func ChainIDFromString(chainID string) (*big.Int, error) {

	b := new(big.Int)
	b, ok := b.SetString(chainID, 10)
	if !ok {
		return nil, fmt.Errorf("could not convert '%s' to big.Int", chainID)
	}

	if !b.IsUint64() {
		return nil, fmt.Errorf("'%s' can not be represented as a uint64", chainID)
	}

	if b.Cmp(maxChainID) > 0 {
		return nil, fmt.Errorf("'%s' to excedes supported chainID range", chainID)
	}
	return b, nil
}
