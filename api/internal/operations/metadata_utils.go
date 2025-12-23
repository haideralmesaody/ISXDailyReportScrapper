package operations

// cloneMetadata creates a shallow copy of a metadata map.
func cloneMetadata(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
