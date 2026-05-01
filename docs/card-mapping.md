# Card mapping

Assigning a card to each expense is the core of the tool. It is implemented in `internal/mapping/mapper.go`.

## Priority order

For each expense, the following sources are evaluated in order. The first one that produces a result wins; the rest are not checked.

```
1. Override by expense ID   →  config.yaml › overrides_by_expense_id
2. Tag in the Notes field   →  [CARD:ID] in the expense "details" field
3. Rule by category         →  config.yaml › rules_by_category
4. Unassigned               →  "UNASSIGNED"
```

### 1. Override by expense ID

```yaml
# config.yaml
overrides_by_expense_id:
  98765432: AMEX
```

If the expense `id` appears in this map, that card is used regardless of anything else. Useful for one-off cases where the category rule or tag does not apply (e.g. a grocery run paid exceptionally with a different card).

### 2. Tag in the Notes field (`details`)

In Splitwise, the **Notes** field of an expense accepts free text. If that text contains the pattern:

```
[CARD:CARD_ID]
```

the tool extracts the ID and uses it as the card. The ID must be uppercase and may contain letters, digits, and underscores. The rest of the Notes field is preserved as-is in the Excel output.

**Valid examples:**

```
Weekly groceries [CARD:VISA_GALICIA]
[CARD:AMEX] appliance installments
work lunch [CARD:DEBIT_SANTANDER] reimbursable
```

The tag can appear anywhere in the field.

**Regex used:** `\[CARD:([A-Z0-9_]+)\]`

This mechanism is the most convenient way to assign cards expense by expense directly from the Splitwise app, without editing the config file.

### 3. Rule by category

```yaml
# config.yaml
rules_by_category:
  Groceries: VISA_GALICIA
  Utilities: DEBIT_SANTANDER
```

The expense's category name (the `category.name` field from the API) is compared against the map keys. The match is exact and case-sensitive.

Ideal for expenses that always go to the same card by nature (e.g. groceries always on Visa).

### 4. Unassigned (`UNASSIGNED`)

If none of the sources above produce a result, the expense is marked as `UNASSIGNED`. It appears in the **"Unassigned"** sheet of the Excel for manual review.

## Resolving unassigned expenses

After generating the report, check the **"Unassigned"** sheet. For each expense you have two options:

**Option A — Tag in Splitwise (recommended):**
Edit the expense in Splitwise and add `[CARD:YOUR_CARD_ID]` to the Notes field. Re-running the command will pick it up automatically.

**Option B — Override in config.yaml:**
Copy the Expense ID from the corresponding column and add it under `overrides_by_expense_id`:

```yaml
overrides_by_expense_id:
  12345678: VISA_GALICIA
```

Re-run the command to regenerate the report.

## Share calculation

Once a card is assigned, two amounts are calculated per expense:

- **My share (`MyShare`)**: the `owed_share` of the user whose ID matches `my_user_id` in the expense's user list.
- **Their share (`TheirShare`)**: the sum of `owed_share` for all other users.

These values are calculated by Splitwise according to the split ratio configured in the group; the tool reads them as-is.

The total expense cost (`cost`) is also shown in the Excel but is not used in internal calculations.

## Full example

Given this expense from Splitwise:

```json
{
  "id": 99001,
  "description": "Supermercado Disco",
  "cost": "8000.00",
  "date": "2026-04-10T18:00:00Z",
  "category": { "name": "Groceries" },
  "details": "weekly shop",
  "deleted_at": null,
  "users": [
    { "user_id": 111, "paid_share": "8000.00", "owed_share": "5000.00" },
    { "user_id": 222, "paid_share": "0.00",    "owed_share": "3000.00" }
  ]
}
```

With `my_user_id: 111` and the rule `Groceries: VISA_GALICIA`:

| Step | Result |
|------|--------|
| Override by ID | not found |
| Tag in Notes | no `[CARD:...]` present |
| Rule by category | `Groceries` → `VISA_GALICIA` ✓ |
| Assigned card | `VISA_GALICIA` |
| My share | 5000.00 |
| Their share | 3000.00 |
