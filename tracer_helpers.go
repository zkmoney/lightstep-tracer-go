package lightstep

import (
	"context"
	"fmt"
	"reflect"

	opentracing "github.com/opentracing/opentracing-go"
)

// FlushLightStepTracer forces a synchronous Flush.
func FlushLightStepTracer(lsTracer opentracing.Tracer) error {
	tracer, ok := lsTracer.(Tracer)
	if !ok {
		return fmt.Errorf("Not a LightStep Tracer type: %v", reflect.TypeOf(lsTracer))
	}

	tracer.Flush(context.Background())
	return nil
}

// GetLightStepAccessToken returns the currently configured AccessToken.
func GetLightStepAccessToken(lsTracer opentracing.Tracer) (string, error) {
	tracer, ok := lsTracer.(Tracer)
	if !ok {
		return "", fmt.Errorf("Not a LightStep Tracer type: %v", reflect.TypeOf(lsTracer))
	}

	return tracer.Options().AccessToken, nil
}

// CloseTracer synchronously flushes the tracer, then terminates it.
func CloseTracer(tracer opentracing.Tracer) error {
	lsTracer, ok := tracer.(Tracer)
	if !ok {
		return fmt.Errorf("Not a LightStep Tracer type: %v", reflect.TypeOf(tracer))
	}

	lsTracer.Close(context.Background())
	return nil
}
