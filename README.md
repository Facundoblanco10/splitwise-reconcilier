# Splitwise Card Reconciler

Maps Splitwise household expenses to payment cards and generates a monthly Excel report.

## Requirements

- Go 1.21+
- A Splitwise account with API access

## Getting your Splitwise API key

1. Go to https://secure.splitwise.com/apps
2. Click **Register your application** (or use an existing one)
3. Copy the **API Key** shown in your app's settings

## Setup

```bash
# 1. Clone the repo and install dependencies
go mod download

# 2. Create your .env file
cp .env.example .env
# Edit .env and paste your API key

# 3. Create your config file
cp config.example.yaml config.yaml
# Edit config.yaml:
#   - Set splitwise.group_id  (run --list-groups to find it, see below)
#   - Set splitwise.my_user_id (your numeric Splitwise user ID)
#   - Define your cards and mapping rules
```

### Finding your group_id

You can inspect the Splitwise URL when viewing your group:
`https://secure.splitwise.com/#/groups/12345678` → group_id is `12345678`

### Finding your user_id

Run the tool once; it prints your user info on startup including the ID, or check
`https://secure.splitwise.com/api/v3.0/get_current_user` in a browser with your key.

## Configuration reference

```yaml
splitwise:
  group_id: 12345678      # numeric ID of your household group
  my_user_id: 111         # your numeric Splitwise user ID

cards:
  - id: VISA_GALICIA      # internal ID used in rules and overrides
    name: "Visa Galicia"  # display name in the Excel report
    closing_day: 25       # statement closing day (null for debit)

rules_by_category:
  Groceries: VISA_GALICIA   # Splitwise category name → card ID

overrides_by_expense_id:
  98765432: AMEX            # force a specific card for one expense
```

### Card assignment priority

1. **Override by expense ID** (`overrides_by_expense_id` in config)
2. **Tag in the expense Notes field**: `[CARD:VISA_GALICIA]`
3. **Rule by category** (`rules_by_category` in config)
4. **UNASSIGNED** — appears in the "Sin asignar" sheet for manual review

## Running

```bash
go run ./cmd/reconciler --month 2026-04
```

Options:

| Flag | Default | Description |
|------|---------|-------------|
| `--month` | *(required)* | Month to process, format `YYYY-MM` |
| `--config` | `config.yaml` | Path to your config file |

## Output

The report is saved to `output/reporte_YYYY-MM.xlsx` with three sheets:

| Sheet | Contents |
|-------|----------|
| Gastos detallados | One row per expense with date, description, amounts, card |
| Resumen por tarjeta | Total amount owed per card |
| Sin asignar | Expenses with no card matched, for manual assignment |

## Handling unassigned expenses

1. Open the generated Excel and check the **Sin asignar** sheet
2. For each expense, either:
   - Add a `[CARD:YOUR_CARD_ID]` tag to the Notes field in Splitwise, or
   - Add an entry under `overrides_by_expense_id` in `config.yaml`
3. Re-run the command — it re-fetches and re-maps everything

## Building a binary

```bash
go build -o reconciler ./cmd/reconciler
./reconciler --month 2026-04
```
