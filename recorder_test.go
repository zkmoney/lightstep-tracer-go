package lightstep

import (
	"fmt"
	"testing"
	"time"

	"github.com/opentracing/basictracer-go"
	ot "github.com/opentracing/opentracing-go"
)

func makeSpanSlice(length int) []basictracer.RawSpan {
	return make([]basictracer.RawSpan, length)
}

func TestTranslateLogDatas(t *testing.T) {
	ts := time.Unix(1473442150, 0)
	otLogs := make([]ot.LogData, 8)
	for i := 0; i < 8; i++ {
		otLogs[i] = ot.LogData{
			Timestamp: ts,
			Event:     fmt.Sprintf("Event%v", i),
			Payload:   map[string]interface{}{"hi": i, "bye": i, "bool": true, "string": "suhhh"},
		}
	}
	res := translateLogDatas(otLogs)
	fmt.Println(res)
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
