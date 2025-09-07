# 📑 Indexing Strategy for Signalbus

## API Keys

### Why

- Every request is authenticated via an API key.
- We look up the tenant by API key hash on every `/publish/:topic` call.
- Needs to be fast → index required.

### Model

```go
type APIKey struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
    Hash      string    `gorm:"not null;uniqueIndex"`
    Scopes    pq.StringArray `gorm:"type:text[]"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

    Tenant    Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}
```

### Generated SQL

```sql
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(hash);
```

---

## Policies

### Why

- Policies define how each tenant handles events (topics).
- A tenant cannot have duplicate policies for the same topic.
- Enforce uniqueness across `(tenant_id, topic)`.

### Model

```go
type Policy struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TenantID  uuid.UUID `gorm:"type:uuid;not null;index:idx_tenant_topic,unique"`
    Topic     string    `gorm:"size:100;not null;index:idx_tenant_topic,unique"`
    Channels  pq.StringArray `gorm:"type:text[];not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

    Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}
```

### Generated SQL

```sql
CREATE UNIQUE INDEX idx_tenant_topic ON policies(tenant_id, topic);
```

---

## Tenants

### Why

- Core entity for all records (notifications, api_keys, policies).
- Frequently joined and filtered.
- Tenant `id` is primary key and already indexed by default.

---

## Recommendations

- **Start simple**: DB indexes (as above) + query plan analysis with `EXPLAIN ANALYZE`.
- **High throughput**: Add **Redis cache** for `{api_key_hash → tenant_id}` lookups.
- **Future option**: Self-contained JWT-style API keys (no DB hit per request).

---
