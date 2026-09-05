package dragout

import "errors"

type Result string

const (
	Dropped   Result = "dropped"
	Cancelled Result = "cancelled"
)

var ErrUnsupported = errors.New("native file dragging is not available on this platform")
