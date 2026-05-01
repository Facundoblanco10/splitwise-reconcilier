# Testing

## Running the tests

```bash
make test          # run all tests
make test-v   # verbose output (shows each test case)
```

All tests complete in under a second — there are no network calls or file I/O beyond temp files.

---

## What is tested

Tests live in `internal/mapping/`, the only package with domain logic worth isolating.

### `mapper_test.go` — card resolution and share calculation

#### `TestResolveCard_Tag`

Verifies the regex-based tag parser against ten cases:

| Case | Input notes | Expected card |
|------|-------------|---------------|
| Explicit tag | `[SCOTIA:FATI]` | `SCOTIA:FATI` |
| Mixed-case entity | `[scotia:Fati]` | `SCOTIA:Fati` |
| Hyphen in entity | `[ITAU-D:John]` | `ITAU-D:John` |
| Lowercase + hyphen, no person | `[itau-d]` | `ITAU-D:John` (payer auto-fill) |
| No colon, no person | `[SCOTIA]` | `SCOTIA:John` (payer auto-fill) |
| Empty person segment | `[SCOTIA:]` | `SCOTIA:John` (payer auto-fill) |
| Payer is second user | `[AMEX]` | `AMEX:Jane` |
| Tag mid-sentence | `monthly bill [AMEX:Jane] installment` | `AMEX:Jane` |
| Multiple tags | `[VISA:A] [AMEX:B]` | `VISA:A` (first wins) |
| Unknown payer for auto-fill | `[SCOTIA]` (payer ID not in name map) | `SCOTIA` |

#### `TestResolveCard_DefaultEntity`

Verifies the priority-2 fallback: when there is no tag, the card is derived from the payer's `default_entity` + their first name.

#### `TestResolveCard_Unassigned`

Covers the three ways an expense ends up as `UNASSIGNED`:
- Payer is not listed in `config.yaml` members
- Payer is listed but has no `default_entity`
- Expense has no users at all

#### `TestResolveShares`

Verifies the `owed_share` split across four scenarios: two-way split, multiple other users (their shares are summed), expense paid only by me, and expense paid only by the other person.

#### `TestMapAll`

Integration smoke test: feeds two expenses through `MapAll` and checks that both `Card` and `MyShare` are resolved end-to-end.

#### `TestFormatCardName`

Checks that `UNASSIGNED` formats to `"Unassigned"` and all other card IDs pass through unchanged.

#### `TestParseCost`

Checks parsing of valid floats, zero, empty string, non-numeric string, and negative values.

---

### `rules_test.go` — config loading

#### `TestLoadConfig`

| Sub-test | Verifies |
|----------|----------|
| Valid config | All fields (`group_id`, `my_user_id`, both members) are parsed correctly |
| Missing file | Returns an error |
| Invalid YAML | Returns an error |
| Wrong field type | Returns an error (e.g. `group_id: not_a_number`) |
| Empty members list | Parses without error, `Members` is an empty slice |

---

## What is not tested

| Package | Reason |
|---------|--------|
| `internal/splitwise` | Thin HTTP wrapper around `net/http`. Testing it meaningfully requires either a live API key or an HTTP mock server — both add complexity with little marginal value for a personal tool. |
| `internal/report` | Delegates entirely to `excelize`. The interesting behaviour (correct cell values, styles, sheet names) is best verified by opening the generated file. |
| `cmd/reconciler` | Orchestration-only; it has no logic beyond wiring the other packages together. |
