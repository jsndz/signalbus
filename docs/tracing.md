I am using opentelemetry for tracing 

what it does is,

create span and links it

```go

func InitTracer(serviceName string,logr *zap.Logger) func() {
	ctx:= context.Background()

// create exporter to send the traces
	exporter,err :=otlptracegrpc.New(ctx)
	if err!=nil{
		logr.Fatal(err.Error(), zap.Error(err))
	}
//new tracing resources
	res,err := resource.New(ctx,resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		logr.Fatal(err.Error(), zap.Error(err))
    }

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
	)
// convert it to complete tracer
	otel.SetTracerProvider(tp)
//function to close it
	return func() {
        ctx, cancel := context.WithTimeout(ctx, time.Second*5)
        defer cancel()
        if err := tp.Shutdown(ctx); err != nil {
			logr.Fatal(err.Error(), zap.Error(err))
        }
    }
}	

```


use this by running it on main.go which initializes the tracer

and 

create a trcer using
'
	tracer := otel.Tracer("notification_api")
'

on this tracer create spans::

tracer_context, span := tracer.Start(c, "handle-sending")
		defer span.End()

and t=you can update these based on condition that run eg:
	if !matched {
			log.Warn("no policy matched event", zap.String("event_type", req.EventType), zap.String("tenant", tenant.ID.String()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "no policy matches event type"})
			span.SetStatus(codes.Error, "no policy matched event")
			return
		}


can pass this through kafka
```go
headers := make(map[string]string)
otel.GetTextMapPropagator().Inject(tracer_context, propagation.MapCarrier(headers))
kafkaCtx, kafkaSpan := tracer.Start(tracer_context, "publish-kafka")
if err := p.Publish(kafkaCtx, topic, key, msgBytes); err != nil {
    kafkaSpan.RecordError(err)
    kafkaSpan.SetStatus(codes.Error, err.Error())
    log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
} else {
    log.Debug("published message", zap.String("topic", topic), zap.String("tenant", tenant.ID.String()))
    
}
```