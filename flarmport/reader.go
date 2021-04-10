package flarmport

import "context"

// Common interface for Conn and Port objects.
type Reader interface {
	// Range iterates over the values received from the flarm.
	Range(context.Context, func(Data)) error
	// Close stops reading flarm data.
	Close() error
}
