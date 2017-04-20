package basictracer

import (
	"encoding/base64"

	"github.com/gogo/protobuf/proto"
	lightstep "github.com/lightstep/lightstep-tracer-go/lightsteppb"
	opentracing "github.com/opentracing/opentracing-go"
)

type lightstepBinaryPropagator struct {
}

func (_ *lightstepBinaryPropagator) Inject(
	spanContext opentracing.SpanContext,
	opaqueCarrier interface{},
) error {
	sc, ok := spanContext.(SpanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	var scarrier *string
	var bcarrier *[]byte
	switch t := opaqueCarrier.(type) {
	case *string:
		scarrier = t
	case *[]byte:
		bcarrier = t
	default:
		return opentracing.ErrInvalidCarrier
	}
	pb := &lightstep.BinaryCarrier{}
	pb.BasicCtx = &lightstep.BasicTracerCarrier{
		TraceId:      sc.TraceID,
		SpanId:       sc.SpanID,
		Sampled:      true,
		BaggageItems: sc.Baggage,
	}
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	if bcarrier == nil {
		*scarrier = base64.StdEncoding.EncodeToString(data)
	} else {
		*bcarrier = make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(*bcarrier, data)
	}
	return nil
}

func (_ *lightstepBinaryPropagator) Extract(
	opaqueCarrier interface{},
) (opentracing.SpanContext, error) {
	var scarrier string
	var bcarrier []byte
	switch t := opaqueCarrier.(type) {
	case string:
		scarrier = t
	case []byte:
		bcarrier = t
	default:
		return nil, opentracing.ErrInvalidCarrier
	}
	var data []byte
	var err error
	if bcarrier == nil {
		data, err = base64.StdEncoding.DecodeString(scarrier)
	} else {
		var n int
		data = make([]byte, base64.StdEncoding.DecodedLen(len(bcarrier)))
		n, err = base64.StdEncoding.Decode(data, bcarrier)
		if err == nil {
			data = data[:n]
		}
	}
	if err != nil {
		return nil, err
	}
	pb := &lightstep.BinaryCarrier{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		return nil, err
	}
	if pb.BasicCtx == nil {
		return nil, opentracing.ErrInvalidCarrier
	}

	return SpanContext{
		TraceID: pb.BasicCtx.TraceId,
		SpanID:  pb.BasicCtx.SpanId,
		Baggage: pb.BasicCtx.BaggageItems,
	}, nil
}
