# Configuration

The tool is configured through two files: `.env` for secrets and `config.yaml` for business logic.

Run `make setup` (or `bash setup.sh`) for an interactive prompt that creates both files for you.

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

Controls which group to query and the member-to-entity mappings.

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
members:
  - splitwise_id: 111       # numeric Splitwise user ID
    default_entity: SCOTIA  # bank entity used when no [ENTITY] tag is present
  - splitwise_id: 222
    default_entity: SANTANDER
```

**`splitwise_id`** — The numeric Splitwise user ID. Get it from `GET /get_current_user` (printed at startup) for yourself; for other members, inspect the group response or the expense user list.

**`default_entity`** — Bank entity used when an expense has no `[ENTITY:PERSON]` tag. Combined at runtime with the payer's Splitwise first name to form `ENTITY:FirstName`.

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
# Current month (auto-detected)
make run

# Specific month
make run MONTH=2026-04

# Specific month with a custom config path
go run ./cmd/reconciler --month 2026-04 --config /home/user/my-configs/splitwise.yaml

# Build a binary and run it directly
make build
./reconciler --month 2026-04
```
