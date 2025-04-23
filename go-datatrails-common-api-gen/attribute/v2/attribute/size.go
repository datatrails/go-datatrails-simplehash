package attribute

// Provides functions to get the size of Attribute instances

const (
	// MaxAttributesSize is the maximum total size in bytes, for the event+asset attributes.
	//
	// Currently the limiting factor is the eventsourcepoller topic with a max bytes of a message at
	//   262144 bytes. (256kb due to the pricing tier we have for azure service bus)
	//
	// The message being added to the topic is the evidence format for the entire event
	//   (including asset creation events). So it is not the same size of the requested asset/event
	//   via the api.
	//
	// We need some wiggle room for possible size increase due to conversion from api -> evidence format.
	// We need to take into account sizes principal_declared, principal_accepted, and timestamps as well.
	//
	// For now make the max 51,200 bytes. This allows for 10 attributes
	//   with key size 1024 and value size 4096 to be created.
	//
	// Note. the input validation for attribute key/validation is measured in golang rune count,
	//   not number of bytes.
	//
	// Note. it is easy to increase the limit, but harder to decrease.
	// Also Note. assets requests of size 100,000 bytes have been added successfully onto the block chain.
	MaxAttributesSize = 51200
)

// AttributeSize gets the size of the given attribute in bytes.
func AttributeSize(attr *Attribute) int64 {

	if attr == nil {
		return int64(0)
	}

	if strAttr, ok := attr.GetString(); ok {
		return sizeStrAttribute(strAttr)
	}

	if listAttr, ok := attr.GetList(); ok {
		return sizeListAttribute(listAttr)
	}

	if dictAttr, ok := attr.GetDict(); ok {
		return sizeDictAttribute(dictAttr)
	}

	// if we get to here we don't know what the atttribute type is.
	return 0
}

// AttributesSize gets the size of the given attributes in bytes.
//
// including the key values in the map.
func AttributesSize(attrs map[string]*Attribute) int64 {

	var size int64 = 0
	for key, attr := range attrs {
		size += sizeStrAttribute(key)
		size += AttributeSize(attr)
	}
	return size
}

// sizeStrAttribute gets the size of a string attribute in bytes
func sizeStrAttribute(attr string) int64 {
	return int64(len(attr))
}

func sizeKeyValuePair(key, value string) int64 {
	return int64(len(key) + len(value))
}

// sizeListAttribute gets the size of a list attribute in bytes
func sizeListAttribute(attr []map[string]string) int64 {

	var size int64 = 0
	for _, value := range attr {
		size += sizeDictAttribute(value)
	}

	return size
}

// sizeDictAttribute gets the size of a dict attribute in bytes
func sizeDictAttribute(attr map[string]string) int64 {

	var size int64 = 0
	for key, value := range attr {
		size += sizeKeyValuePair(key, value)
	}

	return size
}

// MakeUnion makes a union of map attributes from a list of attribute maps.
//
// the value priority for same keys goes for maps later in the argument list.
// e.g:
//
//	if `map1["foo"] = "1"`` and `map2["foo"] = "2"`,
//	then the resulting map would be `mapResult["foo"] = "2"``
func MakeUnion(attributesMaps ...map[string]*Attribute) map[string]*Attribute {

	attrs := make(map[string]*Attribute)

	for _, attributes := range attributesMaps {
		for k, v := range attributes {
			attrs[k] = v
		}
	}

	return attrs
}
