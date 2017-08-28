package lightstep

import "fmt"

var (
	errConnectionWasClosed = fmt.Errorf("the connection was closed")
	errTracerDisabled      = fmt.Errorf("tracer is disabled; aborting Flush()")
)
