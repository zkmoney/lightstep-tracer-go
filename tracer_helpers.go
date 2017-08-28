package lightstep

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

// Flush forces a synchronous Flush.
func Flush(ctx context.Context, tracer opentracing.Tracer) {
	lsTracer, ok := tracer.(Tracer)
	if !ok {
		maybeLogError(newErrNotLisghtepTracer(tracer), true)
	}
	lsTracer.Flush(ctx)
}

// CloseTracer synchronously flushes the tracer, then terminates it.
func Close(ctx context.Context, tracer opentracing.Tracer) {
	lsTracer, ok := tracer.(Tracer)
	if !ok {
		maybeLogError(newErrNotLisghtepTracer(tracer), true)
	}
	lsTracer.Close(ctx)
}

// GetLightStepAccessToken returns the currently configured AccessToken.
func GetLightStepAccessToken(tracer opentracing.Tracer) (string, error) {
	lsTracer, ok := tracer.(Tracer)
	if !ok {
		return "", newErrNotLisghtepTracer(tracer)
	}

	return lsTracer.Options().AccessToken, nil
}

// DEPRECATED: use Flush instead.
func FlushLightStepTracer(lsTracer opentracing.Tracer) error {
	tracer, ok := lsTracer.(Tracer)
	if !ok {
		return newErrNotLisghtepTracer(lsTracer)
	}

	tracer.Flush(context.Background())
	return nil
}

// DEPRECATED: use Close instead.
func CloseTracer(tracer opentracing.Tracer) error {
	lsTracer, ok := tracer.(Tracer)
	if !ok {
		return newErrNotLisghtepTracer(tracer)
	}

	lsTracer.Close(context.Background())
	return nil
}

func newErrNotLisghtepTracer(tracer opentracing.Tracer) error {
	return fmt.Errorf("Not a LightStep Tracer type: %v", reflect.TypeOf(tracer))
}
