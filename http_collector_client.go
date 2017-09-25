package lightstep

import (
	"math/rand"
	"time"
	"golang.org/x/net/context"
	"github.com/lightstep/lightstep-tracer-go/collectorpb"
	"net/http"
	"golang.org/x/net/http2"
	"io/ioutil"
	"github.com/golang/protobuf/proto"
	"io"
	"bytes"
)

// grpcCollectorClient specifies how to send reports back to a LightStep
// collector via grpc.
type httpCollectorClient struct {
	// auth and runtime information
	attributes map[string]string
	reporterID uint64

	// accessToken is the access token used for explicit trace collection requests.
	accessToken string
	maxReportingPeriod time.Duration // set by GrpcOptions.MaxReportingPeriod
	reconnectPeriod    time.Duration // set by GrpcOptions.ReconnectPeriod
	reportingTimeout   time.Duration // set by GrpcOptions.ReportTimeout

	// Remote service that will receive reports.
	socketAddress string
	client        http.Client
	connTimestamp time.Time

	// converters
	converter *protoConverter

	// For testing purposes only
	httpConnectorFactory ConnectorFactory
}

func newHttpCollectorClient(opts Options, reporterID uint64, attributes map[string]string) *httpCollectorClient {
	return &httpCollectorClient{
		accessToken:          opts.AccessToken,
		attributes:           attributes,
		maxReportingPeriod:   opts.ReportingPeriod,
		reportingTimeout:     opts.ReportTimeout,
		reporterID:           reporterID,
		socketAddress:        opts.Collector.HostPort(),
		reconnectPeriod:      time.Duration(float64(opts.ReconnectPeriod) * (1 + 0.2*rand.Float64())),
		converter:            newProtoConverter(opts),
		httpConnectorFactory: opts.ConnFactory,
	}
}

func (client *httpCollectorClient) ConnectClient() (Connection, error) {
	client.connTimestamp = time.Now()
	client.client = http.Client{
		Transport: &http2.Transport{},
	}
	return nil, nil
}

func (client *httpCollectorClient) ShouldReconnect() bool {
	return time.Now().Sub(client.connTimestamp) > client.reconnectPeriod
}

func (client *httpCollectorClient) Report(ctx context.Context, buffer *reportBuffer) (collectorResponse, error) {
	bytes, err := client.toRequest(buffer)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", client.socketAddress + "/v1/report", bytes)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/octet-stream")

	response, err := client.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return client.toResponse(response.Body)
}

func (client *httpCollectorClient) toRequest(buffer *reportBuffer) (io.Reader, error) {
	protoRequest := client.converter.toReportRequest(
		client.reporterID,
		client.attributes,
		client.accessToken,
		buffer,
	)

	buf, err := proto.Marshal(protoRequest)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

func (client *httpCollectorClient) toResponse(reader io.Reader) (collectorResponse, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	protoResponse := &collectorpb.ReportResponse{}
	if err := proto.Unmarshal(body, protoResponse); err != nil {
		return nil, err
	}

	return protoResponse, nil
}
