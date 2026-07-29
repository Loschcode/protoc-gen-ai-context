# protoc-gen-ai-context

A protoc plugin that generates AI-consumable knowledge files from proto annotations.

**The missing bridge between gRPC contracts and AI agent context.**

## Problem

AI agents that work with gRPC APIs need to understand what each capability does, what payloads look like, and what users might call things in natural language. Today this knowledge is maintained in hand-written markdown files that drift from the actual proto contracts. When you add a new enum value to your proto, the AI doesn't know about it until someone manually updates the knowledge file.

## Solution

Annotate your proto files with AI context metadata. The plugin generates structured markdown knowledge files that AI agents can load on demand — split by topic to minimize token usage.

```proto
import "ai/context/v1/annotations.proto";

enum FormFieldType {
  option (ai.context.v1.enum_knowledge) = {
    topic: "forms"
    title: "Forms"
    summary: "Form step for data collection."
  };

  FORM_FIELD_TYPE_TEXT = 1 [(ai.context.v1.enum_value_context) = {
    human_name: "Text"
    common_queries: "text input, free text, short answer"
  }];

  FORM_FIELD_TYPE_RATING = 10 [(ai.context.v1.enum_value_context) = {
    human_name: "Rating"
    common_queries: "star rating, score, NPS, satisfaction"
  }];
}

message FormPayload {
  option (ai.context.v1.message_knowledge) = {
    topic: "forms"
    summary: "Payload for creating a form step."
    example: "{\"form_title\": \"Rate Us\", \"form_fields\": [{\"type\": \"RATING\", \"label\": \"How would you rate us?\"}]}"
  };

  repeated FormField form_fields = 1;
  optional string form_title = 2;
}
```

Running `protoc` with this plugin generates `forms.ai.md`:

```markdown
# Forms

Form step for data collection. Payload for creating a form step.
Capabilities: Text, Rating. Common queries: text input, free text,
short answer, star rating, score, NPS, satisfaction.

## FormFieldType
- **Text** (`FORM_FIELD_TYPE_TEXT`) — (users may say: text input, free text, short answer)
- **Rating** (`FORM_FIELD_TYPE_RATING`) — (users may say: star rating, score, NPS, satisfaction)

## Payload for creating a form step (FormPayload)
Fields:
- `form_fields` (repeated FormField)
- `form_title` (string) *(optional)*

Example payload:
...
```

## Key Features

- **Single source of truth** — proto files define both the API contract and the AI knowledge
- **Topic-based splitting** — each `topic` value produces one `.md` file, keeping token usage minimal
- **Capability summaries** — the generated first paragraph contains all keywords for topic index discoverability
- **Zero drift** — add an enum value with an annotation, run `make proto`, knowledge updates automatically
- **Works with buf** — standard protoc plugin, compatible with buf.gen.yaml

## Installation

```bash
go install github.com/Loschcode/protoc-gen-ai-context/cmd/protoc-gen-ai-context@latest
```

## Usage with buf

Add to your `buf.gen.yaml`:

```yaml
plugins:
  - plugin: protoc-gen-ai-context
    out: internal/agents/knowledge
```

Then run:

```bash
buf generate
```

## Usage with protoc

```bash
protoc \
  --ai-context_out=internal/agents/knowledge \
  --proto_path=proto \
  proto/api/v1/*.proto
```

## Proto Annotations Reference

### `enum_knowledge` / `message_knowledge` (KnowledgeTopic)

Apply on enums or messages to assign them to a knowledge topic.

| Field | Type | Description |
|-------|------|-------------|
| `topic` | string | Topic identifier — becomes the output filename (e.g. `"forms"` → `forms.ai.md`) |
| `title` | string | Human-readable title for the markdown heading |
| `summary` | string | Keywords and description appended to the topic's index summary |
| `example` | string | JSON payload example included in the generated file |

### `enum_value_context` (EnumValueContext)

Apply on individual enum values for AI discoverability.

| Field | Type | Description |
|-------|------|-------------|
| `human_name` | string | Natural-language name (e.g. `"Rating"` instead of `"FORM_FIELD_TYPE_RATING"`) |
| `common_queries` | string | Comma-separated user queries that should match this value |
| `topic` | string | Override topic (inherits from parent enum if empty) |

### `field_context` (FieldContext)

Apply on message fields for richer documentation.

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable field description |
| `example` | string | Example value |

## Companion Libraries

- [grpc-mcp-gateway](https://github.com/Loschcode/grpc-mcp-gateway) — Proto annotations → MCP tool server (how external systems **call** your API)
- **protoc-gen-ai-context** (this library) — Proto annotations → AI knowledge files (how AI agents **understand** your API)

## License

MIT
