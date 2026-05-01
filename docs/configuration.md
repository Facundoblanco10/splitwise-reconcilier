# Configuration

The tool is configured through two files: `.env` for secrets and `config.yaml` for business logic.

---

## `.env`

Contains only the Splitwise API key. Excluded from the repository via `.gitignore`.

```env
SPLITWISE_API_KEY=your_api_key_here
```

If the variable is already present in the shell environment, the `.env` file is optional: `godotenv` attempts to load it but does not fail if it is missing.

### Getting the API key

1. Go to **https://secure.splitwise.com/apps**
2. Click **Register your application** (or open an existing one)
3. Copy the **API Key** value

---

## `config.yaml`

Controls which group to query, the known cards, and the mapping rules.

### Full reference

```yaml
splitwise:
  group_id: 12345678    # numeric ID of your Splitwise group
  my_user_id: 111       # your numeric Splitwise user ID
```

**`group_id`** — Found in the URL when viewing the group in Splitwise:
`https://secure.splitwise.com/#/groups/12345678`

**`my_user_id`** — Retrieve it from the API:
```bash
curl -H "Authorization: Bearer <API_KEY>" \
     https://secure.splitwise.com/api/v3.0/get_current_user
```
The `id` field in the response is your user ID.

---

```yaml
cards:
  - id: VISA_GALICIA          # internal identifier (used in rules and overrides)
    name: "Visa Galicia"      # display name shown in the Excel report
    closing_day: 25           # statement closing day (null for debit cards)
  - id: AMEX
    name: "American Express"
    closing_day: 5
  - id: DEBITO_SANTANDER
    name: "Débito Santander"
    closing_day: null
```

**`id`** — Uppercase string with underscores. This is the identifier you use in `rules_by_category`, `overrides_by_expense_id`, and in `[CARD:ID]` tags inside Splitwise notes.

**`closing_day`** — Informational for now; does not affect calculations. Useful to know which billing cycle a given expense falls into.

---

```yaml
rules_by_category:
  Groceries: VISA_GALICIA
  Utilities: DEBITO_SANTANDER
  Entertainment: AMEX
```

Maps a **Splitwise category name** to a card `id`. The match is exact and case-sensitive. See [Card mapping](card-mapping.md) for the full priority logic.

Common Splitwise category names:

| Category | Description |
|----------|-------------|
| `Groceries` | Supermarket, food shopping |
| `Utilities` | Electricity, gas, water, internet |
| `Rent` | Rent or mortgage |
| `Entertainment` | Outings, subscriptions |
| `Electronics` | Electronics and appliances |
| `General` | Default when no category is set |

---

```yaml
overrides_by_expense_id:
  98765432: AMEX
  98765433: VISA_GALICIA
```

Forces a specific card for one expense, ignoring any rule or tag. The expense ID can be found in the Splitwise URL when opening the expense, or in the **Expense ID** column of the generated Excel.

This has the highest priority in resolution. See [Card mapping](card-mapping.md).

---

## CLI flags

```
go run ./cmd/reconciler [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--month` | *(required)* | Month to process, format `YYYY-MM` (e.g. `2026-04`) |
| `--config` | `config.yaml` | Path to the configuration file |

### Examples

```bash
# Current month
go run ./cmd/reconciler --month 2026-05

# Previous month with config in a custom location
go run ./cmd/reconciler --month 2026-04 --config /home/user/my-configs/splitwise.yaml

# Build a binary and run it
go build -o reconciler ./cmd/reconciler
./reconciler --month 2026-04
```
