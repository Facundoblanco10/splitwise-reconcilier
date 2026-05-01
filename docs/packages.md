# Package reference

## `internal/splitwise`

Handles all communication with the Splitwise REST API.

**Base URL:** `https://secure.splitwise.com/api/v3.0`  
**Auth:** `Authorization: Bearer <API_KEY>` header

---

### `Client`

```go
func NewClient(apiKey string) *Client
```

Creates an HTTP client ready to use. Internally instantiates a default `*http.Client` (no explicit timeout; the tool is interactive and data volume is small).

---

### `Client.GetCurrentUser() (*User, error)`

Calls `GET /get_current_user`. Used at startup to validate the API key before doing anything else.

```go
type User struct {
    ID        int
    FirstName string
    LastName  string
    Email     string
}
```

---

### `Client.GetGroups() ([]Group, error)`

Calls `GET /get_groups`. Useful for listing available groups and finding the right `group_id`. Not used in the main flow currently.

```go
type Group struct {
    ID   int
    Name string
}
```

---

### `Client.GetExpenses(groupID int, from, to time.Time) ([]Expense, error)`

Calls `GET /get_expenses` with the following query params:

| Param | Value |
|-------|-------|
| `group_id` | Group ID |
| `dated_after` | `from` in RFC3339 format |
| `dated_before` | `to` in RFC3339 format |
| `limit` | `500` (fixed) |

Automatically filters out expenses where `deleted_at != null` before returning.

```go
type Expense struct {
    ID          int
    Description string
    Cost        string       // string from the API, e.g. "5000.00"
    Date        string       // RFC3339
    Details     string       // Splitwise Notes field
    DeletedAt   *string      // nil if the expense is active
    Category    Category
    Users       []ExpenseUser
}

type Category struct {
    ID   int
    Name string
}

type ExpenseUser struct {
    UserID    int
    PaidShare string   // how much this user paid
    OwedShare string   // how much this user owes
}
```

Monetary fields come as `string` from the API and are parsed to `float64` in the `mapping` package.

---

## `internal/mapping`

Loads configuration and applies the card assignment logic.

---

### `LoadConfig(path string) (*Config, error)`

Reads and parses `config.yaml`. Returns an error if the file does not exist or has invalid YAML syntax.

```go
type Config struct {
    Splitwise struct {
        GroupID  int
        MyUserID int
    }
    Cards                []Card
    RulesByCategory      map[string]string
    OverridesByExpenseID map[int]string
}

type Card struct {
    ID         string
    Name       string
    ClosingDay *int    // nil for debit cards
}
```

---

### `Mapper`

```go
func NewMapper(cfg *Config) *Mapper
func (m *Mapper) MapAll(expenses []splitwise.Expense) []MappedExpense
```

`MapAll` iterates over all expenses and applies `mapOne` to each, which in turn calls `resolveCard` and `resolveShares`.

```go
type MappedExpense struct {
    Expense    splitwise.Expense
    Card       string    // card ID or "UNASSIGNED"
    MyShare    float64
    TheirShare float64
}
```

---

### `FormatCardName(cfg *Config, cardID string) string`

Converts an internal `cardID` to its display name for the Excel output. Returns `(ID)` in parentheses if the ID is not in the config. Returns `"Sin asignar"` for `UNASSIGNED`.

---

### `ParseCost(s string) float64`

Parses a monetary string from the API (`"5000.00"`) to `float64`. Returns `0` if the string is not a valid number.

---

### Constant `Unassigned`

```go
const Unassigned = "UNASSIGNED"
```

Sentinel value indicating no card was found for an expense.

---

## `internal/report`

Generates the Excel file with three sheets.

---

### `Generate(expenses []MappedExpense, cfg *Config, month time.Time, outputPath string) error`

The package's single entry point. Creates the file at `outputPath` with the three sheets described below. Overwrites the file if it already exists.

---

### Generated sheets

#### "Gastos detallados" (Expense detail)

One row per expense. Columns:

| Column | Content |
|--------|---------|
| Fecha | `YYYY-MM-DD` |
| Descripción | `expense.Description` |
| Categoría | `expense.Category.Name` |
| Monto total | `expense.Cost` (float, `#,##0.00` format) |
| Mi parte | `MappedExpense.MyShare` |
| Parte pareja | `MappedExpense.TheirShare` |
| Tarjeta | Card display name |
| Expense ID | `expense.ID` |
| Notes | `expense.Details` (original Splitwise Notes field) |

#### "Resumen por tarjeta" (Per-card summary)

Two columns: card display name and the sum of `MyShare` across all expenses assigned to it. Order follows the `cards` list in `config.yaml`; `UNASSIGNED` is always last.

#### "Sin asignar" (Unassigned)

Same columns as the detail sheet, minus "Parte pareja" and "Tarjeta" (irrelevant here). Only includes expenses where `Card == "UNASSIGNED"`.

---

### Applied styles

| Element | Style |
|---------|-------|
| Headers on all sheets | Bold + light blue background (`#D9E1F2`) |
| Monetary columns | `#,##0.00` format (NumFmt 7) |
| Column widths | Set per content type (dates: 12, descriptions: 35, etc.) |
