package lightstep

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	cpb "github.com/lightstep/lightstep-tracer-go/collectorpb"
	"github.com/opentracing/basictracer-go"
	ot "github.com/opentracing/opentracing-go"
)

func makeSpanSlice(length int) []basictracer.RawSpan {
	return make([]basictracer.RawSpan, length)
}

func makeExpectedLogs() []*cpb.Log {
	eRes := make([]*cpb.Log, 8)
	for i := 0; i < 8; i++ {
		eRes[i] = &cpb.Log{
			Timestamp: &google_protobuf.Timestamp{1473442150, 0},
			Keyvalues: []*cpb.KeyValue{
				&cpb.KeyValue{Key: messageKey, Value: &cpb.KeyValue_StringValue{fmt.Sprintf("Event%v", i)}},
				&cpb.KeyValue{Key: payloadKey, Value: &cpb.KeyValue_IntValue{int64(i)}},
				&cpb.KeyValue{Key: payloadKey, Value: &cpb.KeyValue_IntValue{int64(i)}},
				&cpb.KeyValue{Key: payloadKey, Value: &cpb.KeyValue_BoolValue{true}},
				&cpb.KeyValue{Key: payloadKey, Value: &cpb.KeyValue_StringValue{"suhhh"}},
			},
		}
	}
	return eRes
}

func TestTranslateLogDatas(t *testing.T) {
	ts := time.Unix(1473442150, 0)
	otLogs := make([]ot.LogData, 8)
	for i := 0; i < 8; i++ {
		otLogs[i] = ot.LogData{
			Timestamp: ts,
			Event:     fmt.Sprintf("Event%v", i),
			Payload:   []interface{}{i, i, true, "suhhh"},
		}
	}
	res := translateLogDatas(otLogs)
	eRes := makeExpectedLogs()
	if !reflect.DeepEqual(res, eRes) {
		t.Errorf("%v doesn not equal %v", res, eRes)
		fmt.Println("res")
		spew.Dump(res)
		fmt.Println("eRes")
		spew.Dump(eRes)
	}
	//for i := 0; i < 8; i++ {
	//	if !reflect.DeepEqual(eRes[i].Timestamp, res[i].Timestamp) {
	//		t.Errorf("the timestamps do not match res: %v, expected res: %v", res[i].Timestamp, eRes[i].Timestamp)
	//	}
	//	if !reflect.DeepEqual(eRes[i].Keyvalues, res[i].Keyvalues) {
	//		t.Errorf("the keyvalue pair does not match. kv res: %v, expected kv: %v", res[i].Keyvalues, eRes[i].Keyvalues)
	//	}
	//}
}

func TestMaxBufferSize(t *testing.T) {
	recorder := NewRecorder(Options{
		AccessToken: "0987654321",
	}).(*Recorder)

	checkCapSize := func(spanLen, spanCap int) {
		recorder.lock.Lock()
		defer recorder.lock.Unlock()

		if recorder.buffer.cap() != spanCap {
			t.Error("Unexpected buffer size")
		}
		if recorder.buffer.len() != spanLen {
			t.Error("Unexpected buffer size")
		}
	}

	checkCapSize(0, defaultMaxSpans)

	spans := makeSpanSlice(defaultMaxSpans)
	for _, span := range spans {
		recorder.RecordSpan(span)
	}

	checkCapSize(defaultMaxSpans, defaultMaxSpans)

	spans = append(spans, makeSpanSlice(defaultMaxSpans)...)
	for _, span := range spans {
		recorder.RecordSpan(span)
	}

	checkCapSize(defaultMaxSpans, defaultMaxSpans)

	maxBuffer := 10
	recorder = NewRecorder(Options{
		AccessToken:      "0987654321",
		MaxBufferedSpans: maxBuffer,
	}).(*Recorder)

	checkCapSize(0, maxBuffer)

	spans = append(spans, makeSpanSlice(100*defaultMaxSpans)...)
	for _, span := range spans {
		recorder.RecordSpan(span)
	}

	checkCapSize(maxBuffer, maxBuffer)
}
