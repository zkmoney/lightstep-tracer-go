package truncator

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	. "testing"
	"unsafe"
)

//----------------------------------------------------------------------------//
// Coverage tests
//----------------------------------------------------------------------------//
//
// Tested designed to get as closed to 100% from `go test -cover` as possible.
//

func TestCoverage(t *T) {
	//
	// Run through all the sample payloads to get good "base" coverage
	//
	truncators := []*Truncator{
		NewTruncator(32, 256),
		NewTruncator(1024, 16*1024),
	}
	for _, str := range samplePayloadJSON {
		source := mustUnmarshalJSON(str)
		for _, s := range truncators {
			s.TruncateToJSON(source)
		}
	}

	//
	// Edge cases (unsafe.Pointer, chan)
	//
	someInt := 0
	truncators[1].TruncateToJSON(map[interface{}]interface{}{
		0: unsafe.Pointer(&someInt),
		1: make(chan string),
	})
}

//----------------------------------------------------------------------------//
// Benchmarks
//----------------------------------------------------------------------------//

func BenchmarkTruncateNil(b *B) {
	s := newTestTruncator()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.TruncateToJSON(nil)
	}
}

func BenchmarkTruncateString(b *B) {
	s := newTestTruncator()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.TruncateToJSON("This is a string")
	}
}

func BenchmarkTruncateStringSlice256(b *B) {
	doBench(b, newTestTruncator(), makeStringSlice(256))
}
func BenchmarkTruncateStringSlice1k(b *B) {
	doBench(b, newTestTruncator(), makeStringSlice(1024))
}
func BenchmarkTruncateStringSlice4k(b *B) {
	doBench(b, newTestTruncator(), makeStringSlice(4096))
}
func BenchmarkTruncateStringSlice16k(b *B) {
	doBench(b, newTestTruncator(), makeStringSlice(4*4096))
}

func BenchmarkTruncateBrowserPayload(b *B) {
	doBench(b, newTestTruncator(), mustUnmarshalJSON(samplePayloadJSON["browser_payload"]))
}
func BenchmarkTruncatePerfSnapshotPayload(b *B) {
	doBench(b, newTestTruncator(), mustUnmarshalJSON(samplePayloadJSON["perf_snapshot"]))
}
func BenchmarkTruncateAllPayloads(b *B) {
	payloads := make([]interface{}, 0, len(samplePayloadJSON))
	for _, str := range samplePayloadJSON {
		payloads = append(payloads, mustUnmarshalJSON(str))
	}

	tr := newTestTruncator()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, p := range payloads {
			tr.TruncateToJSON(p)
		}
	}
}

//----------------------------------------------------------------------------//
// Unit tests
//----------------------------------------------------------------------------//

func testJsonEqual(t *T, expectedJson string, obj interface{}) {
	cleanActual, _ := json.Marshal(obj)
	tmp := map[string]interface{}{}
	json.Unmarshal([]byte(expectedJson), &tmp)
	cleanExpected, _ := json.Marshal(tmp)
	if string(cleanActual) != string(cleanExpected) {
		t.Errorf("expected != actual.\n\nexpected: %v\n\nactual: %v\n", string(cleanExpected), string(cleanActual))
	}
}

func TestValueToJSON(t *T) {

	checkJson := func(inObj interface{}, expected interface{}, fieldMaxBytes, totalMaxBytes int) {
		// Convert to thrift, then convert that to a json-friendly interface{}.
		truncator := &Truncator{fieldMaxBytes, totalMaxBytes}

		rval, err := truncator.TruncateToJSON(inObj)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		var obj interface{}
		err = json.Unmarshal([]byte(rval), &obj)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if !reflect.DeepEqual(obj, expected) {
			fmt.Printf("\n\nActual:   %v\n\nExpected: %v\n\n", obj, expected)
			t.Error("Structures differ")
		}
	}

	checkJsonLength := func(inObj interface{}, fieldMaxBytes, totalMaxBytes, expectedMinBytes, expectedMaxBytes int) {
		// Convert to thrift, then convert that to a json-friendly interface{}.
		truncator := &Truncator{fieldMaxBytes, totalMaxBytes}
		rval, err := truncator.TruncateToJSON(inObj)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if len(rval) < expectedMinBytes || len(rval) > expectedMaxBytes {
			t.Errorf("Length constraint not satisfied: %v < %v < %v\nActual JSON: %v",
				expectedMinBytes, len(rval), expectedMaxBytes, rval)
		}
	}

	type wackyStruct struct {
		Exported     float32
		unexported   int8
		AGenericMap  map[string]interface{}
		AFloatKeyMap map[float64]interface{}
		SomeFunc     func()
		Interface    interface{}
	}

	emptyMap := map[string]string{}
	var nilMap map[string]string = nil

	ptrToEmptyMap := &emptyMap
	ptrToNilMap := &nilMap

	wacky := &wackyStruct{
		Exported:   3.14,
		unexported: 42,
		AGenericMap: map[string]interface{}{
			"key1": "one",
			"key2": 2,
			"key3": &wackyStruct{},
		},
		AFloatKeyMap: map[float64]interface{}{
			3.14: "pi",
			2.71: struct {
				Name string
				Type string
			}{"e", "transcendental"},
			0.0:          emptyMap,
			0.1:          &emptyMap,
			0.2:          nilMap,
			0.3:          &nilMap,
			0.4:          ptrToEmptyMap,
			0.5:          &ptrToEmptyMap,
			0.6:          ptrToNilMap,
			0.7:          &ptrToNilMap,
			math.Inf(-1): "-infinity",
			math.Inf(+1): "+infinity",
		},
		SomeFunc: func() {
			fmt.Printf("don't panic!")
		},
	}

	expectedResult := map[string]interface{}{
		"AFloatKeyMap": map[string]interface{}{
			"+Inf": "+infinity",
			"-Inf": "-infinity",
			"0":    make(map[string]interface{}),
			"0.1":  make(map[string]interface{}),
			"0.2":  nil,
			"0.3":  nil,
			"0.4":  make(map[string]interface{}),
			"0.5":  make(map[string]interface{}),
			"0.6":  nil,
			"0.7":  nil,
			"2.71": map[string]interface{}{
				"Name": "e",
				"Type": "transcendental",
			},
			"3.14": "pi",
		},
		"AGenericMap": map[string]interface{}{
			"key1": "one",
			"key2": "2",
			"key3": map[string]interface{}{
				"AFloatKeyMap": nil,
				"AGenericMap":  nil,
				"Exported":     "0",
				"Interface":    nil,
				"SomeFunc":     "\u003cfunc\u003e",
			},
		},
		"Exported":  "3.14",
		"Interface": nil,
		"SomeFunc":  "\u003cfunc\u003e",
	}

	checkJson(wacky, expectedResult, 100, 1000) // plenty of maxBytes headroom

	fieldTruncatedResult := map[string]interface{}{
		// Note how all dynamic-length strings longer than 8 chars are truncated.
		"AFloatKe…": map[string]interface{}{
			"+Inf": "+infinit…",
			"-Inf": "-infinit…",
			"0":    make(map[string]interface{}),
			"0.1":  make(map[string]interface{}),
			"0.2":  nil,
			"0.3":  nil,
			"0.4":  make(map[string]interface{}),
			"0.5":  make(map[string]interface{}),
			"0.6":  nil,
			"0.7":  nil,
			"2.71": map[string]interface{}{
				"Name": "e",
				"Type": "transcen…",
			},
			"3.14": "pi",
		},
		"AGeneric…": map[string]interface{}{
			"key1": "one",
			"key2": "2",
			"key3": map[string]interface{}{
				"AFloatKe…": nil,
				"AGeneric…": nil,
				"Exported":  "0",
				"Interfac…": nil,
				"SomeFunc":  "\u003cfunc\u003e",
			},
		},
		"Exported":  "3.14",
		"Interfac…": nil,
		"SomeFunc":  "\u003cfunc\u003e",
	}
	checkJson(wacky, fieldTruncatedResult, 8, 1000) // not enough per-field headroom

	// These strings have a lot of filler, so even a maxBytes val of 25 can
	// result in a considerably longer JSON-encoded final product. E.g., for
	// the totalMaxBytes=25 case, we get something like this:
	//
	//    {"AGenericMap":{"key1":"on…","…":"…"},"Exported":"3.14","…":"…"}
	//
	// It's 74 runes, but only 33 are actual content (including ellipses!).
	checkJsonLength(wacky, 1024 /* plenty of field headroom */, 25, 25, 100)
	checkJsonLength(wacky, 1024 /* plenty of field headroom */, 125, 125, 310)

	// Now try restricting both field and total message length for a single
	// really long string.
	longStr64K := strings.Repeat("-", 1024*64)
	checkJsonLength(longStr64K, 100, 1024*128, 100, 110) // limit field length
	checkJsonLength(longStr64K, 1024*128, 100, 100, 110) // limit total length

	justNil := interface{}(nil)
	checkJson(justNil, nil, 100, 1000)

	justNilMap := (map[string]string)(nil)
	checkJson(justNilMap, nil, 100, 1000)

	justNilInMap := map[string]interface{}{"foo": nil}
	checkJson(justNilInMap, justNilInMap, 100, 1000)

	justInterfaceNil := interface{}(interface{}(nil))
	checkJson(justInterfaceNil, nil, 100, 1000)

	justPointerNil := (*int)(nil)
	checkJson(justPointerNil, nil, 100, 1000)

	type LinkNode struct {
		Value int
		Next  *LinkNode
	}

	objectA := &LinkNode{1, nil}
	objectB := &LinkNode{2, nil}
	objectA.Next = objectB
	objectB.Next = objectA

	// TODO: should test that this does not follow the circular reference
	// (which it does at the moment; that hasn't been addressed)
	truncator := &Truncator{1024, 4096}
	_, err := truncator.TruncateToJSON(objectA)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

//----------------------------------------------------------------------------//
// Helpers
//----------------------------------------------------------------------------//

func doBench(b *B, tr *Truncator, data interface{}) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tr.TruncateToJSON(data)
	}
}

func newTestTruncator() *Truncator {
	return NewTruncator(1024, 16*1024)
}

func makeStringSlice(n int) []string {
	data := make([]string, n)
	for j := range data {
		data[j] = fmt.Sprintf("index = %v", j)
	}
	return data
}

func mustUnmarshalJSON(jsonStr string) interface{} {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(err)
	} else {
		return data
	}
}

//----------------------------------------------------------------------------//
// Test data
//----------------------------------------------------------------------------//

var samplePayloadJSON map[string]string = map[string]string{
	"empty": "{}",
	"browser_payload": `
{
  "navigator": {
    "appCodeName": "Mozilla",
    "appName": "Netscape",
    "appVersion": "5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36",
    "connection": {},
    "cookieEnabled": "true",
    "doNotTrack": null,
    "geolocation": {},
    "hardwareConcurrency": "8",
    "language": "en-US",
    "languages": [
      "en-US",
      "en"
    ],
    "maxTouchPoints": "0",
    "mediaDevices": {},
    "mimeTypes": [
      {
        "description": "",
        "suffixes": "pdf",
        "type": "application/pdf"
      },
      {
        "description": "Widevine Content Decryption Module",
        "suffixes": "",
        "type": "application/x-ppapi-widevine-cdm"
      },
      {
        "description": "Shockwave Flash",
        "suffixes": "swf",
        "type": "application/x-shockwave-flash"
      },
      {
        "description": "Shockwave Flash",
        "suffixes": "spl",
        "type": "application/futuresplash"
      },
      {
        "description": "Native Client Executable",
        "suffixes": "",
        "type": "application/x-nacl"
      },
      {
        "description": "Portable Native Client Executable",
        "suffixes": "",
        "type": "application/x-pnacl"
      },
      {
        "description": "Portable Document Format",
        "suffixes": "pdf",
        "type": "application/x-google-chrome-pdf"
      }
    ],
    "nfc": {},
    "onLine": "true",
    "permissions": {},
    "platform": "MacIntel",
    "plugins": [
      {
        "description": "",
        "name": "Chrome PDF Viewer"
      },
      {
        "description": "Enables Widevine licenses for playback of HTML audio/video content. (version: 1.4.8.866)",
        "name": "Widevine Content Decryption Module"
      },
      {
        "description": "Shockwave Flash 20.0 r0",
        "name": "Shockwave Flash"
      },
      {
        "description": "",
        "name": "Native Client"
      },
      {
        "description": "Portable Document Format",
        "name": "Chrome PDF Viewer"
      }
    ],
    "presentation": {},
    "product": "Gecko",
    "productSub": "20030107",
    "serviceWorker": {
      "controller": null
    },
    "services": {},
    "storage": {},
    "storageQuota": {},
    "usb": {},
    "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36",
    "vendor": "Google Inc.",
    "vendorSub": "",
    "webkitPersistentStorage": {},
    "webkitTemporaryStorage": {}
  }
}`,

	"perf_snapshot": `
{
  "BytesFromSystem": "1070114864",
  "BytesInUse": "227855328",
  "HeapInUse": "227855328",
  "LastGCMicros": "1453838273864643",
  "NextGCHeapInUse": "347787336",
  "PerfSampleMicros": "1453838275217718"
}`,
}
