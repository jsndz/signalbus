A span is just a named piece of work in your system.

"API received request" → span

"Produced message to Kafka" → span

Each span has:

Name → what’s happening ("POST /notify", "Kafka publish")

Start & end time → how long it took

Attributes (tags) → metadata (tenant_id, channel="email", provider="sendgrid")

Status → success or error

Multiple spans linked together = a trace

Trace: Notification Lifecycle
└── Span A: API POST /notify (root span)
    ├── Span B: Save to DB
    ├── Span C: Produce to Kafka
    └── Span D: Worker consume from Kafka
        ├── Span E: Render template
        └── Span F: Send email via SendGrid


Then comes with context propagation.
Each span has trace_id
For example:
When the API creates Span A, it generates a trace_id.
It puts this trace_id into Kafka message headers.
The worker picks up the message, extracts the trace_id, and continues the trace with Spans D–F.
With Grafana stick everything and you get a end-to-end journey as one trace view.