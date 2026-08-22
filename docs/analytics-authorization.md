# Analytics Authorization

UIForge analytics uses GrokifyQL as the safe query language for saved Questions
and ad-hoc analytics execution. In server mode, GrokifyQL text is parsed into an
AST, checked against structural policy, optionally checked against SystemForge
authorization, and only then dispatched to the configured analytics source.

## Flow

```text
question/query text
  -> grokifyql.Parse
  -> UIForge analytics catalog schema
  -> SystemForge authz.Authorizer
  -> GrokifyQL policy
  -> source QueryProvider
```

The policy provider lives in `internal/server/api/grokifyql_policy.go`.

When UIForge has no authorization service or no authenticated principal in the
request context, the provider falls back to the baseline GrokifyQL safety policy:

- read-only operations
- expression depth limit
- expression node count limit
- maximum `IN` list size

When SystemForge authorization is configured and a principal is present, UIForge
builds a request-scoped policy from SpiceDB/SystemForge decisions.

## Resource Model

UIForge extends its SpiceDB schema with analytics resources:

```zed
definition analytics_source {
    relation org: organization
    relation owner: principal
    relation data_source: data_source

    permission manage = owner + org->admin + data_source->manage
    permission view = manage + org->viewer + data_source->view
    permission query = manage + org->editor + data_source->use
}

definition analytics_dataset {
    relation source: analytics_source
    relation owner: principal

    permission manage = owner + source->manage
    permission read = manage + source->query
    permission list = read
    permission sort = read
}

definition analytics_field {
    relation dataset: analytics_dataset
    relation owner: principal

    permission manage = owner + dataset->manage
    permission read = manage + dataset->read
    permission list = read
    permission sort = read
}
```

Datasets and fields map to deterministic UUID resources derived from:

```text
uiforge:analytics:<source_id>:<dataset_query_name>
uiforge:analytics:<source_id>:<dataset_query_name>:<field_query_name>
```

This gives SystemForge's SpiceDB provider concrete type/id/permission tuples to
check.

## GrokifyQL Permission Mapping

UIForge compiles SpiceDB decisions into GrokifyQL field policy:

| GrokifyQL use | SystemForge action | SpiceDB permission |
|---------------|--------------------|--------------------|
| Dataset access | `read` | `analytics_dataset#read` |
| Select field | `read` | `analytics_field#read` |
| Filter field | `list` | `analytics_field#list` |
| Sort/group field | `sort` | `analytics_field#sort` |

Saved Questions and ad-hoc query execution use the same policy provider. This
prevents unsaved queries from bypassing field-level authorization.

## Saved Questions

Questions are persisted by UIForge's backend and compiled at save time. The
compiled metadata includes the parsed GrokifyQL AST, fingerprint, datasets,
fields, read-only flag, and limit. Persisting the AST gives UIForge a stable
representation for audit and future re-validation, but the query should still be
checked against the current policy before execution because permissions can
change.

## Relationship Sync

The schema defines the relationships; product code must still write them to
SpiceDB. Typical relationships are:

- `analytics_source:<id>#org@organization:<org_id>`
- `analytics_dataset:<id>#source@analytics_source:<source_id>`
- `analytics_field:<id>#dataset@analytics_dataset:<dataset_id>`

Direct `owner` relationships can be added for source, dataset, or field-specific
administration. Organization and data source relationships should be preferred
for broad product access.
