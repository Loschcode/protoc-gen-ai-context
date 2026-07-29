# Examples

Practical examples showing how to annotate proto files with `protoc-gen-ai-context` so that AI agents can understand your API.

## Examples

### E-commerce (`ecommerce/product_catalog.proto`)

A product catalog and order system demonstrating:

- **Topic splitting** — `ProductCategory` and `ProductPayload` share the `"products"` topic, while `PaymentMethod` and `OrderPayload` share the `"orders"` topic. Two knowledge files are generated from one proto file.
- **`field_context`** — every message field has a human-readable description and example value.
- **`usage_notes`** on messages — business rules like "price_cents is always in the smallest currency unit" that are not derivable from the schema.

### Helpdesk (`helpdesk/tickets.proto`)

A support ticket system demonstrating:

- **Behavioral rules in `usage_notes`** — SLA enforcement ("HIGH priority tickets must be assigned within 1 hour"), valid state transitions ("OPEN -> IN_PROGRESS -> WAITING_ON_CUSTOMER"), and auto-close rules.
- **Multiple enums in one topic** — `TicketPriority` and `TicketStatus` both contribute to the `"tickets"` topic, so their usage notes and capabilities appear in a single knowledge file.
- **Agent-visible vs customer-visible fields** — `internal_note` is annotated to make clear it is never shown to the customer.

### Healthcare (`healthcare/appointments.proto`)

A medical appointment scheduling system demonstrating:

- **`common_queries` for natural-language mapping** — patients say "checkup" or "follow-up with my doctor" and the AI maps to the correct `AppointmentType` enum value.
- **Symptom-to-specialty routing** via `usage_notes` — "chest pain -> CARDIOLOGY, skin rash -> DERMATOLOGY".
- **Cross-field validation rules** — `previous_appointment_id` is required for `FOLLOW_UP` type, insurance fields are required for `PROCEDURE` type.

## Running the examples

### With buf

From the repository root, create a `buf.gen.yaml` that includes the examples:

```yaml
version: v2
plugins:
  - local: protoc-gen-ai-context
    out: output
inputs:
  - directory: examples
  - directory: proto
```

Then run:

```bash
buf generate
```

### With protoc

```bash
protoc \
  --ai-context_out=output \
  --proto_path=proto \
  --proto_path=examples \
  examples/ecommerce/product_catalog.proto \
  examples/helpdesk/tickets.proto \
  examples/healthcare/appointments.proto
```

### Expected output

The plugin generates one `.ai.md` file per topic:

| Topic | File | Contents |
|-------|------|----------|
| `products` | `products.ai.md` | ProductCategory enum + ProductPayload message |
| `orders` | `orders.ai.md` | PaymentMethod enum + OrderPayload message |
| `tickets` | `tickets.ai.md` | TicketPriority + TicketStatus enums + TicketPayload message |
| `appointments` | `appointments.ai.md` | AppointmentType + SpecialtyType enums + AppointmentPayload message |

Each generated file contains:

1. A summary paragraph with keywords for topic discovery
2. Enum capability tables with human names and common queries
3. Message field documentation with types and examples
4. Usage notes sections with behavioral rules
5. JSON payload examples
