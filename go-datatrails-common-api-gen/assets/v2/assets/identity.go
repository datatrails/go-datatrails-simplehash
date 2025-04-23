package assets

import (
	"strings"

	"github.com/google/uuid"
)

const (
	assetIdentityPrefix  = "assets/"
	eventIdentityPrefix  = "/events/"
	publicIdentityPrefix = "public"
)

// IdentityFactory generates idintities for Events and Assets base on random UUIDv4
type IdentityFactory interface {
	// GenerateAssetIdentity produces RRN for asset
	GenerateAssetIdentity() string
	// GenerateEventIdentity produces RRN for event
	GenerateEventIdentity(string) string
}

// UUIDIdentityFactory generates UUID based identities
type UUIDIdentityFactory struct{}

// GenerateAssetIdentity produces RRN for asset
func (u *UUIDIdentityFactory) GenerateAssetIdentity() string {
	return assetIdentityPrefix + uuid.New().String()
}

// GenerateEventIdentity produces RRN for event
func (u *UUIDIdentityFactory) GenerateEventIdentity(assetIdentity string) string {
	return assetIdentity + eventIdentityPrefix + uuid.New().String()
}

// Handle UUID <-> Identity conversion.
// Ideally uuid and intentity would be different types so we don't use the wrong
// one in the wrong place, but protobuf does not seem to support "typedef" of one
// type to another

// AssetIdentityFromUuid derives RRN identity for asset from a bare uuid id
func AssetIdentityFromUuid(uuid string) string {
	return assetIdentityPrefix + uuid
}

// AssetUuidFromIdentity derives asset base uuid id from RRN identity
// When the prefix does not match this will return the original identity unchanged
func AssetUuidFromIdentity(assetIdentity string) string {
	return strings.TrimPrefix(assetIdentity, assetIdentityPrefix)
}

// EventIdentityFromUuid derives RRN identity for asset from base uuid ids
func EventIdentityFromUuid(assetUuid, eventUuid string) string {
	return assetIdentityPrefix + assetUuid + eventIdentityPrefix + eventUuid
}

// EventUuidFromIdentity gets the event UUID from an event identity
func EventUuidFromIdentity(eventIdentity string) string {
	eventSplit := strings.Split(eventIdentity, eventIdentityPrefix)
	return eventSplit[len(eventSplit)-1]
}

// AssetUuidFromEventIdentity gets the asset UUID from an event identity
func AssetUuidFromEventIdentity(eventIdentity string) string {
	eventSplit := strings.Split(eventIdentity, eventIdentityPrefix)
	assetSplit := strings.Split(eventSplit[0], assetIdentityPrefix)

	return assetSplit[len(assetSplit)-1]
}

// PublicIdentityFromPermissioned returns the public identity from the permissioned identity
func PublicIdentityFromPermissioned(permissionedIdentity string) string {
	return publicIdentityPrefix + permissionedIdentity
}

// PermissionedIdentityFromPublic returns the permissioned identity from the public identity
//
// NOTE: if the given identity is already a permissioned identity, return the permissioned identity unchanged.
func PermissionedIdentityFromPublic(publicIdentity string) string {
	return strings.TrimPrefix(publicIdentity, publicIdentityPrefix)
}
