package httpx

import (
	"encoding/json"
	"errors"
)

// MergePatch applies a JSON merge patch to the target struct
// It merges non-zero/non-nil fields from patch into target
// Follows RFC 7396 JSON Merge Patch semantics
func MergePatch(target interface{}, patchJSON []byte) error {
	if len(patchJSON) == 0 {
		return errors.New("patch: empty patch data")
	}

	// Parse patch into a map
	var patchMap map[string]interface{}
	if err := json.Unmarshal(patchJSON, &patchMap); err != nil {
		return err
	}

	// Convert target struct to map
	targetJSON, err := json.Marshal(target)
	if err != nil {
		return err
	}

	var targetMap map[string]interface{}
	if err := json.Unmarshal(targetJSON, &targetMap); err != nil {
		return err
	}

	// Merge patch into target map
	mergeMaps(targetMap, patchMap)

	// Convert back to struct
	mergedJSON, err := json.Marshal(targetMap)
	if err != nil {
		return err
	}

	return json.Unmarshal(mergedJSON, target)
}

// mergeMaps recursively merges patch into target following JSON Merge Patch semantics
func mergeMaps(target, patch map[string]interface{}) {
	for key, patchValue := range patch {
		if patchValue == nil {
			// null in patch means remove the field (RFC 7396)
			delete(target, key)
			continue
		}

		// Check if patch value is a map and target value is also a map
		if patchMap, ok := patchValue.(map[string]interface{}); ok {
			if targetValue, exists := target[key]; exists {
				if targetMap, ok := targetValue.(map[string]interface{}); ok {
					// Recursively merge nested maps
					mergeMaps(targetMap, patchMap)
					continue
				}
			} else {
				// Target doesn't have this key, create new map
				target[key] = make(map[string]interface{})
				mergeMaps(target[key].(map[string]interface{}), patchMap)
				continue
			}
		}

		// For all other values (including empty strings, zero numbers, false booleans),
		// update the target. This allows clearing fields by setting them to zero values.
		target[key] = patchValue
	}
}
