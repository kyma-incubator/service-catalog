package tester

import "github.com/appscode/jsonpatch"

func FilterOutStatusPatch(operations []jsonpatch.JsonPatchOperation) []jsonpatch.JsonPatchOperation {
	var filtered []jsonpatch.JsonPatchOperation
	for _, op := range operations {
		if op.Path != "/status" {
			filtered = append(filtered, op)
		}
	}

	return filtered
}
