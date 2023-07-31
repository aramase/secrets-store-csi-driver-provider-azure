package metrics

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	providerAttr = attribute.String("provider", "azure")
	osTypeAttr   = attribute.String("os_type", runtime.GOOS)
	// set service.name attribute explicitly to the provider name as the default service name is "unknown_service:<binary name>"
	serviceNameAttr = attribute.String("service.name", "csi-secrets-store-provider-azure")
	objectTypeKey   = "object_type"
	objectNameKey   = "object_name"
	errorKey        = "error"
	grpcMethodKey   = "grpc_method"
	grpcCodeKey     = "grpc_code"
	grpcMessageKey  = "grpc_message"
)

type reporter struct {
	keyvaultRequest metric.Float64Histogram
	grpcRequest     metric.Float64Histogram
}

// StatsReporter is the interface for reporting metrics
type StatsReporter interface {
	ReportKeyvaultRequest(ctx context.Context, duration float64, objectType, objectName, err string)
	ReportGRPCRequest(ctx context.Context, duration float64, method, code, message string)
}

// NewStatsReporter creates a new StatsReporter
func NewStatsReporter() (StatsReporter, error) {
	var err error

	r := &reporter{}
	meter := global.Meter("csi-secrets-store-provider-azure")

	if r.keyvaultRequest, err = meter.Float64Histogram("keyvault_request", metric.WithDescription("Distribution of how long it took to get from keyvault")); err != nil {
		return nil, err
	}
	if r.grpcRequest, err = meter.Float64Histogram("grpc_request", metric.WithDescription("Distribution of how long it took for the gRPC requests")); err != nil {
		return nil, err
	}
	return r, nil
}

// ReportKeyvaultRequest reports the duration of the keyvault request
// objectType and objectName are used to identify the object being accessed
// err is used to identify the error if any
func (r *reporter) ReportKeyvaultRequest(ctx context.Context, duration float64, objectType, objectName, err string) {
	opt := metric.WithAttributes{
		serviceNameAttr,
		providerAttr,
		osTypeAttr,
		attribute.Key(objectTypeKey).String(objectType),
		attribute.Key(objectNameKey).String(objectName),
		attribute.Key(errorKey).String(err),
	}
	r.keyvaultRequest.Record(ctx, duration, opt)
}

// ReportGRPCRequest reports the duration of the gRPC request
// method and code are used to identify the gRPC request
func (r *reporter) ReportGRPCRequest(ctx context.Context, duration float64, method, code, message string) {
	opt := metric.WithAttributes{
		serviceNameAttr,
		providerAttr,
		osTypeAttr,
		attribute.Key(grpcMethodKey).String(method),
		attribute.Key(grpcCodeKey).String(code),
		attribute.Key(grpcMessageKey).String(message),
	}
	r.grpcRequest.Record(ctx, duration, opt)
}
