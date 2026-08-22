# Agent Integration

UIForge can host agent-assisted analytics workflows, but the reusable agent
marketplace model belongs in
[`github.com/plexusone/omniagent`](https://github.com/plexusone/omniagent).
UIForge consumes OmniAgent marketplace listings and exposes UIForge-specific
capabilities to those agents instead of defining a separate marketplace model.

## Boundary

- **OmniAgent owns marketplace primitives**: agent listings, skill listings,
  provider interfaces, filtering, featured/listed/private visibility, and
  adapters from OmniAgent runtime objects.
- **UIForge owns analytics capabilities**: saved Questions, GrokifyQL execution,
  field value lookup, result summaries, dashboard widgets, and dashboard layout.
- **SystemForge/SpiceDB owns authorization** when fine-grained checks are
  enabled. UIForge maps analytics resources to permissions and asks SystemForge
  before executing privileged actions.

This keeps agents portable across UIForge, OmniRoadmap, VisionStudio, and future
applications.

## Capability Names

UIForge should publish stable capability names when registering marketplace
agents or skills:

| Capability | Purpose |
|------------|---------|
| `uiforge.question.read` | Read saved Questions and their metadata |
| `uiforge.question.write` | Create, update, or delete saved Questions |
| `uiforge.query.run` | Execute read-only GrokifyQL against an analytics source |
| `uiforge.query.explain` | Summarize or explain a GrokifyQL statement |
| `uiforge.field_values.read` | Retrieve distinct values for a field |
| `uiforge.dashboard.read` | Read dashboard definitions and widget references |
| `uiforge.dashboard.write` | Add or update dashboard widgets |

Capabilities are descriptive contracts. The UIForge backend still enforces the
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

When an assistant response is applied, UIForge switches the SQL pane into edit
mode so the user can review the generated query before saving or running it.

## Recommended Flow

1. UIForge loads an OmniAgent marketplace provider.
2. UIForge filters listings by capabilities such as `uiforge.query.run`.
3. The user chooses an agent for the Question builder.
4. UIForge sends the agent only the current analytics context required for the
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

UIForge should resolve analytics source credentials through OmniVault-compatible
references such as `env://UIFORGE_OMNIROADMAP_DSN`, `keyring://uiforge/source`,
`aws-sm://uiforge/analytics/omniroadmap#dsn`, or
`sql://analytics-sources/omniroadmap-local`. VaultGuard can sit in front of
OmniVault when the deployment needs environment detection, posture checks, and
provider policy.

The `sql://` form is the fallback for self-contained installs. It should use
OmniVault's SQL store provider rather than a UIForge-specific encryption path,
so migration to an external vault only changes provider configuration and stored
references.

The UI should show whether a credential reference is configured and testable,
but it should not echo resolved secret values back to the browser.
