# Composite Unique Index in Postgres (Simple Explanation)

## Without Composite Index

Imagine you have a table with millions of rows.

If you run a query like:

```sql
SELECT *
FROM templates
WHERE tenant_id = 't1'
  AND channel = 'email'
  AND name = 'welcome_email'
  AND locale = 'en-US'
  AND content_type = 'text/html'
LIMIT 1;
```

If you only have single-column indexes (on `tenant_id`, `channel`, etc.):

* Postgres can use only **one index** (for example, `tenant_id`).
* It finds all rows for tenant `t1`.
* Then it still has to filter through them to match `channel`, `name`, `locale`, and `content_type`.

That might mean scanning thousands of rows before finding the right one.

---

## With Composite Unique Index

If you create a composite index:

```sql
CREATE UNIQUE INDEX idx_templates_lookup
    ON templates (tenant_id, channel, name, locale, content_type);
```

Postgres builds a **sorted B-tree** using the combination of all these columns.
So when you query with all five fields, Postgres can do a direct lookup in the B-tree and go straight to the correct row, without scanning.

This makes lookups much faster.

---

## The Unique Part

Adding **UNIQUE** means Postgres will not allow duplicate rows for the same combination of `(tenant_id, channel, name, locale, content_type)`.

* If you try to insert another row with the same values for these columns, it will fail.
* This ensures there is only one template for each tenant, channel, name, locale, and content type.

---

## Proof with `EXPLAIN`

With only single-column indexes, `EXPLAIN` might show:

```
Bitmap Heap Scan on templates
  Recheck Cond: tenant_id = 't1'
  Filter: (channel='email' AND name='welcome_email' ...)
```

That means it used one index and then filtered the rest.

With the composite index, it shows:

```
Index Scan using idx_templates_lookup on templates
  Index Cond: (tenant_id = 't1' AND channel='email' AND name='welcome_email' AND locale='en-US' AND content_type='text/html')
```

That means it directly used the multi-column index. No filtering needed.

---

## Why This Helps

1. **Performance**: Queries using all five fields can be resolved directly with the index.
2. **Correctness**: The unique index prevents duplicate templates with the same identifiers.

---

## Best Approach

### 1. Composite Unique Index

```sql
CREATE UNIQUE INDEX idx_templates_lookup
    ON templates (tenant_id, channel, name, locale, content_type);
```

This gives fast lookups and enforces uniqueness.

### 2. GORM Model

You can define the unique index in your GORM model:

```go
type Template struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"size:100;not null"`
	Channel     string    `gorm:"type:varchar(20);not null"`
	ContentType string    `gorm:"type:varchar(20);not null"`
	Locale      string    `gorm:"size:10;default:'en-US'"`
	Content     string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	// Composite unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_template_unique,priority:1"`
}
```

You would repeat the `uniqueIndex` tag with different `priority` values for each field in the combination.

### 3. Repository Query

```go
func (r *TemplateRepository) GetByLookup(tenantID uuid.UUID, channel, name, locale, contentType string) (*models.Template, error) {
	var template models.Template
	err := r.db.Where(
		"tenant_id = ? AND channel = ? AND name = ? AND locale = ? AND content_type = ?",
		tenantID, channel, name, locale, contentType,
	).First(&template).Error

	if err != nil {
		return nil, err
	}
	return &template, nil
}
```

---

