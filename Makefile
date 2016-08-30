# LightStep-specific: rebuilds the LightStep thrift protocol files.  Assumes
# the command is run within the LightStep development environment (i.e. the
# LIGHTSTEP_HOME environment variable is set).
.PHONY: thrift proto
thrift:
	thrift --gen go:package_prefix='github.com/lightstep/lightstep-tracer-go/',thrift_import='github.com/lightstep/lightstep-tracer-go/thrift_0_9_2/lib/go/thrift' -out . $(LIGHTSTEP_HOME)/go/src/crouton/crouton.thrift
	rm -rf lightstep_thrift/reporting_service-remote

proto:
	@git submodule update --init
	cd lightstep-tracer-common && protoc --gofast_out=plugins=grpc:. --gofast_out=../collectorpb/ collector.proto
	@rm lightstep-tracer-common/collector.pb.go
