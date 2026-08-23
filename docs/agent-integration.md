# Agent Integration

DashForge can host agent-assisted analytics workflows, but the reusable agent
marketplace model belongs in
[`github.com/plexusone/omniagent`](https://github.com/plexusone/omniagent).
DashForge consumes OmniAgent marketplace listings and exposes DashForge-specific
capabilities to those agents instead of defining a separate marketplace model.

## Boundary

- **OmniAgent owns marketplace primitives**: agent listings, skill listings,
  provider interfaces, filtering, featured/listed/private visibility, and
  adapters from OmniAgent runtime objects.
- **DashForge owns analytics capabilities**: saved Questions, GrokifyQL execution,
  field value lookup, result summaries, dashboard widgets, and dashboard layout.
- **SystemForge/SpiceDB owns authorization** when fine-grained checks are
  enabled. DashForge maps analytics resources to permissions and asks SystemForge
  before executing privileged actions.

This keeps agents portable across DashForge, OmniRoadmap, VisionStudio, and future
applications.

## Capability Names

DashForge should publish stable capability names when registering marketplace
agents or skills:

| Capability | Purpose |
|------------|---------|
| `dashforge.question.read` | Read saved Questions and their metadata |
| `dashforge.question.write` | Create, update, or delete saved Questions |
| `dashforge.query.run` | Execute read-only GrokifyQL against an analytics source |
| `dashforge.query.explain` | Summarize or explain a GrokifyQL statement |
| `dashforge.field_values.read` | Retrieve distinct values for a field |
| `dashforge.dashboard.read` | Read dashboard definitions and widget references |
| `dashforge.dashboard.write` | Add or update dashboard widgets |

Capabilities are descriptive contracts. The DashForge backend still enforces the
actual policy before executing the action.

## Question Assistant

The Question builder includes an LLM chat panel for analytical query creation.
The assistant receives:

- current source and dataset metadata
- catalog fields and field types
- current GrokifyQL
- user prompt

It can return:

- a human-readable summary
- a complete replacement GrokifyQL query
- an optional title
- notes about assumptions or follow-up work

When an assistant response is applied, DashForge switches the SQL pane into edit
mode so the user can review the generated query before saving or running it.

## Recommended Flow

1. DashForge loads an OmniAgent marketplace provider.
2. DashForge filters listings by capabilities such as `dashforge.query.run`.
3. The user chooses an agent for the Question builder.
4. DashForge sends the agent only the current analytics context required for the
   requested action.
5. The backend validates GrokifyQL policy before saving or executing any query.
6. Dashboard widgets reference saved Questions by ID and choose their own
   rendering.

## Security Notes

- Agents should not receive raw database credentials.
- Analytics source configuration should store OmniVault/VaultGuard secret
  references, not raw DSNs or tokens.
- Agents should not bypass GrokifyQL policy validation.
- Saved Questions remain read-only analytics artifacts.
- Field value browsing must use the same dataset and field authorization checks
  as normal query execution.
- Result summarization should be capped by row count and token budget.

## Secrets

DashForge should resolve analytics source credentials through OmniVault-compatible
references such as `env://DASHFORGE_OMNIROADMAP_DSN`, `keyring://dashforge/source`,
`aws-sm://dashforge/analytics/omniroadmap#dsn`, or
`sql://analytics-sources/omniroadmap-local`. VaultGuard can sit in front of
OmniVault when the deployment needs environment detection, posture checks, and
provider policy.

The `sql://` form is the fallback for self-contained installs. It should use
OmniVault's SQL store provider rather than a DashForge-specific encryption path,
so migration to an external vault only changes provider configuration and stored
references.

The UI should show whether a credential reference is configured and testable,
but it should not echo resolved secret values back to the browser.
