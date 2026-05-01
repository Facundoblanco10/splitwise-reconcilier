# Card mapping

Assigning a card to each expense is the core of the tool. It is implemented in `internal/mapping/mapper.go`.

## Card key format

Cards are not pre-defined in config. They are identified at runtime as `ENTITY:PERSON` strings — for example `SCOTIA:John` or `SANTANDER:Jane`. The entity is the bank name (uppercase) and the person is the first name of whoever holds that card.

## Priority order

For each expense, the following sources are evaluated in order. The first one that produces a result wins.

```
1. [ENTITY:PERSON] tag    →  in the expense "details" (Notes) field
2. Payer's default entity →  config.yaml › members[].default_entity + payer first name
3. Unassigned             →  "UNASSIGNED"
```

### 1. `[ENTITY:PERSON]` tag in the Notes field

In Splitwise, the **Notes** field of an expense accepts free text. If it contains a tag of the form:

```
[ENTITY:PERSON]
```

the tool uses `ENTITY:PERSON` as the card key directly. The person segment can be a custom alias or the user's actual first name — both are treated identically.

The tag is **case-insensitive** — `[scotia]`, `[Scotia]`, and `[SCOTIA]` all produce the same result. The entity is always normalized to uppercase internally.

If the person segment is **omitted or empty**, it is auto-filled with the first name of whoever paid the expense:

```
[SCOTIA]    →   SCOTIA:John   (if John paid)
[SCOTIA:]   →   SCOTIA:John   (if John paid)
```

**Examples:**

```
[SCOTIA:FATI]                      →  SCOTIA:FATI
[VISA:John] weekly groceries       →  VISA:John
appliance installments [AMEX:Jane] →  AMEX:Jane
[SANTANDER:]                       →  SANTANDER:<payer first name>
[itau-d]                           →  ITAU-D:<payer first name>
```

The tag can appear anywhere in the Notes field. If multiple tags are present, the first one wins.

**Regex used:** `(?i)\[([A-Z0-9_-]+)(?::([^\]]*))?\]`

### 2. Payer's default entity

```yaml
# config.yaml
members:
  - splitwise_id: 111
    default_entity: SCOTIA
  - splitwise_id: 222
    default_entity: SANTANDER
```

If no tag is present, the tool looks at who paid the expense (`paid_share > 0`) and checks whether that user has a `default_entity` configured. If so, the card is `DEFAULT_ENTITY:PAYER_FIRST_NAME`.

This is the fallback for expenses without any tag or matching category rule, as long as the payer is configured.

> **Note:** for expenses where multiple users share the payment, the first payer found in the list is used.

### 3. Unassigned (`UNASSIGNED`)

If none of the sources above produce a result, the expense is marked as `UNASSIGNED` and appears in the **"Unassigned"** sheet for manual review.

## Resolving unassigned expenses

After generating the report, check the **"Unassigned"** sheet. For each expense you have two options:

**Option A — Tag in Splitwise (recommended):**
Edit the expense in Splitwise and add `[ENTITY:NAME]` to the Notes field. Re-running the command will pick it up automatically.

**Option B — Add a default entity for the payer in config.yaml:**
If the payer doesn't have a `default_entity` configured yet, add one under `members:`.

## Share calculation

Once a card is assigned, two amounts are calculated per expense:

- **My share (`MyShare`)**: the `owed_share` of the user whose ID matches `my_user_id`.
- **Their share (`TheirShare`)**: the sum of `owed_share` for all other users.

The total expense cost (`cost`) is also shown in the Excel but is not used in internal calculations.

## Full example

Given this expense:

```json
{
  "id": 99001,
  "description": "Supermarket",
  "cost": "8000.00",
  "date": "2026-04-10T18:00:00Z",
  "category": { "name": "Groceries" },
  "details": "[SCOTIA:FATI]",
  "users": [
    { "user_id": 111, "paid_share": "8000.00", "owed_share": "5000.00" },
    { "user_id": 222, "paid_share": "0.00",    "owed_share": "3000.00" }
  ]
}
```

| Step | Result |
|------|--------|
| Tag in Notes | `[SCOTIA:FATI]` found → `SCOTIA:FATI` ✓ |
| Assigned card | `SCOTIA:FATI` |
| My share | 5000.00 |
| Their share | 3000.00 |

If the notes were empty and `members` had `splitwise_id: 111, default_entity: SCOTIA`, the card would be resolved as `SCOTIA:John` (user 111's first name).
