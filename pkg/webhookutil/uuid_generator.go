package webhookutil

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// UUIDGenerator generates new UUID
type UUIDGenerator func() types.UID

// New returns new UUID.
// If UUIDGenerator is not initialised, then implementation from "k8s.io/apimachinery/pkg/util/UUID" is used
// otherwise UUIDGenerator is called and result returned.
func (generator UUIDGenerator) New() types.UID {
	if generator == nil {
		return uuid.NewUUID()
	}
	return generator()
}
