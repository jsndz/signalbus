package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)


func InitTracer(serviceName string,logr *zap.Logger) func() {
	ctx:= context.Background()


	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("jaeger:4317"),
	)
		if err!=nil{
		logr.Fatal(err.Error(), zap.Error(err))
	}

	res,err := resource.New(ctx,resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		logr.Fatal(err.Error(), zap.Error(err))
    }

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return func() {
        ctx, cancel := context.WithTimeout(ctx, time.Second*5)
        defer cancel()
        if err := tp.Shutdown(ctx); err != nil {
			logr.Fatal(err.Error(), zap.Error(err))
        }
    }
}	