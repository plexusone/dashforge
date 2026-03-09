# Transforms

Transforms modify data after it's fetched from a data source. They can be applied at the data source level or widget level.

## Usage

### Data Source Level

Transforms applied here affect all widgets using this data source:

```json
{
  "id": "products",
  "type": "postgres",
  "query": "SELECT * FROM products",
  "transform": [
    { "type": "filter", "config": { "field": "active", "operator": "eq", "value": true } },
    { "type": "sort", "config": { "field": "sales", "direction": "desc" } }
  ]
}
```

### Widget Level

Transforms applied here only affect this specific widget:

```json
{
  "id": "top-5-chart",
  "type": "chart",
  "dataSourceId": "products",
  "transform": [
    { "type": "limit", "config": { "count": 5 } }
  ]
}
```

## Transform Types

### filter

Filter rows based on a condition.

```json
{
  "type": "filter",
  "config": {
    "field": "status",
    "operator": "eq",
    "value": "active"
  }
}
```

#### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| eq | Equal to | `"operator": "eq", "value": "active"` |
| neq | Not equal to | `"operator": "neq", "value": "deleted"` |
| gt | Greater than | `"operator": "gt", "value": 100` |
| gte | Greater than or equal | `"operator": "gte", "value": 100` |
| lt | Less than | `"operator": "lt", "value": 50` |
| lte | Less than or equal | `"operator": "lte", "value": 50` |
| in | In array | `"operator": "in", "value": ["a", "b", "c"]` |
| notIn | Not in array | `"operator": "notIn", "value": ["x", "y"]` |
| contains | String contains | `"operator": "contains", "value": "test"` |
| startsWith | String starts with | `"operator": "startsWith", "value": "pre"` |
| endsWith | String ends with | `"operator": "endsWith", "value": "ing"` |
| isNull | Is null | `"operator": "isNull"` |
| isNotNull | Is not null | `"operator": "isNotNull"` |

#### Multiple Conditions

```json
{
  "type": "filter",
  "config": {
    "logic": "and",
    "conditions": [
      { "field": "status", "operator": "eq", "value": "active" },
      { "field": "sales", "operator": "gt", "value": 1000 }
    ]
  }
}
```

### sort

Sort rows by one or more fields.

```json
{
  "type": "sort",
  "config": {
    "field": "sales",
    "direction": "desc"
  }
}
```

#### Multi-field Sort

```json
{
  "type": "sort",
  "config": {
    "fields": [
      { "field": "category", "direction": "asc" },
      { "field": "sales", "direction": "desc" }
    ]
  }
}
```

### limit

Limit the number of rows.

```json
{
  "type": "limit",
  "config": {
    "count": 10,
    "offset": 0
  }
}
```

### extract

Extract a nested path from the data.

```json
{
  "type": "extract",
  "config": {
    "path": "response.data.items"
  }
}
```

Useful for API responses:

```json
// Input: { "status": "ok", "data": { "users": [...] } }
// Output: [...]
{
  "type": "extract",
  "config": { "path": "data.users" }
}
```

### select

Select specific fields (projection).

```json
{
  "type": "select",
  "config": {
    "fields": ["name", "email", "created_at"]
  }
}
```

#### With Aliases

```json
{
  "type": "select",
  "config": {
    "fields": [
      "name",
      { "field": "created_at", "as": "date" },
      { "field": "total_amount", "as": "revenue" }
    ]
  }
}
```

### compute

Add computed fields.

```json
{
  "type": "compute",
  "config": {
    "fields": [
      {
        "name": "profit_margin",
        "expression": "(revenue - cost) / revenue * 100"
      },
      {
        "name": "full_name",
        "expression": "concat(first_name, ' ', last_name)"
      }
    ]
  }
}
```

#### Supported Functions

| Function | Description | Example |
|----------|-------------|---------|
| concat | Concatenate strings | `concat(a, b)` |
| upper | Uppercase | `upper(name)` |
| lower | Lowercase | `lower(name)` |
| round | Round number | `round(value, 2)` |
| floor | Floor | `floor(value)` |
| ceil | Ceiling | `ceil(value)` |
| abs | Absolute value | `abs(value)` |
| coalesce | First non-null | `coalesce(a, b, 0)` |

### aggregate

Aggregate data with grouping.

```json
{
  "type": "aggregate",
  "config": {
    "groupBy": ["category", "region"],
    "aggregations": [
      { "field": "amount", "function": "sum", "as": "total" },
      { "field": "amount", "function": "avg", "as": "average" },
      { "field": "id", "function": "count", "as": "count" }
    ]
  }
}
```

#### Aggregation Functions

| Function | Description |
|----------|-------------|
| sum | Sum of values |
| avg | Average |
| min | Minimum |
| max | Maximum |
| count | Count of rows |
| countDistinct | Count of distinct values |
| first | First value |
| last | Last value |

### pivot

Pivot data from rows to columns.

```json
{
  "type": "pivot",
  "config": {
    "rowField": "product",
    "columnField": "month",
    "valueField": "sales",
    "aggregation": "sum"
  }
}
```

### join

Join with another data source.

```json
{
  "type": "join",
  "config": {
    "dataSourceId": "categories",
    "type": "left",
    "leftField": "category_id",
    "rightField": "id"
  }
}
```

## Chaining Transforms

Transforms are applied in order:

```json
"transform": [
  { "type": "extract", "config": { "path": "data.items" } },
  { "type": "filter", "config": { "field": "active", "operator": "eq", "value": true } },
  { "type": "compute", "config": { "fields": [{ "name": "margin", "expression": "profit / revenue" }] } },
  { "type": "sort", "config": { "field": "margin", "direction": "desc" } },
  { "type": "limit", "config": { "count": 10 } }
]
```
