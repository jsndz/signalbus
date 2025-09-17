

# Kafka Producer Idempotency Explained

Kafka producer idempotency (`enable.idempotence=true`) ensures that the same message, when retried due to network or broker issues, is not appended multiple times to a partition. The critical point is that Kafka does not compare message payloads for duplicates. Instead, it relies on protocol-level sequence numbers tied to the producer session.

---

## How Kafka Enforces Idempotency Internally

### 1. Producer ID (PID)

When a producer with `idempotence=true` connects, the broker assigns it a unique 64-bit Producer ID (PID). This PID is valid for the lifetime of that producer session.

### 2. Sequence Numbers per Partition

Each `(PID, partition)` pair has a monotonic sequence number.

* First message = sequence 0, then 1, 2, 3…
* The producer increments the sequence number before sending each new message.

### 3. Broker Validation

The broker tracks the last sequence number accepted for each `(PID, partition)`. When a new message arrives:

* If `seq == lastSeq + 1`, it is accepted.
* If `seq <= lastSeq`, it is identified as a duplicate and silently discarded.
* If `seq > lastSeq + 1`, the broker detects a gap and signals an error, requiring the producer to retry.

### 4. Retry Scenario

If a producer sends `seq=5` but the network times out, the producer retries by sending `seq=5` again. The broker, which already stored `seq=5`, discards the duplicate and keeps only one copy.

---

## Consumer-Side Behavior

An idempotent producer guarantees that duplicates are not written into the Kafka log. As a result, consumers will never see the same `(PID, sequence number)` written twice.

However, if two separate API calls produce the same payload intentionally, Kafka will treat them as distinct messages. This is because they carry different sequence numbers, even if the message body is identical. Consumers will receive both copies. Kafka does not perform payload-level deduplication.

To achieve semantic idempotency, such as ensuring that only one notification is sent per `event_id`, the message payload should include an explicit idempotency key. Deduplication can then be enforced at the consumer or in downstream storage.

---

## Summary

* **Transport-level idempotency**

  * Enabled with `enable.idempotence=true`.
  * Works through `(Producer ID + partition + sequence number)`.
  * Prevents duplicates caused by retries or broker failovers.

* **Business-level idempotency**

  * Requires adding an `event_id` (or similar key) in the payload.
  * Deduplication must be handled at the consumer or in storage systems.

In practice, enabling producer idempotency prevents transport-layer duplication. To prevent business-level duplication, you must design messages with unique identifiers and enforce deduplication downstream.



This unfortunately is not available in segmentio/kafka.