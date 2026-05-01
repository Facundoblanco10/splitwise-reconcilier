# Architecture

## Directory structure

```
splitwise-reconcilier/
├── cmd/
│   └── reconciler/
│       └── main.go              # Entrypoint and flow orchestration
├── internal/
│   ├── splitwise/
│   │   ├── client.go            # HTTP client with Bearer auth
│   │   ├── models.go            # Structs representing the API response
│   │   └── expenses.go          # GetExpenses with date and group filters
│   ├── mapping/
│   │   ├── rules.go             # Config struct and config.yaml loader
│   │   └── mapper.go            # Card assignment logic per expense
│   └── report/
│       └── excel.go             # .xlsx generation with three sheets
├── docs/                        # This documentation
├── output/                      # Generated reports (git-ignored)
├── config.example.yaml
├── .env.example
└── go.mod
```

All packages live under `internal/` so they cannot be imported from outside the module. The only public entry point is `cmd/reconciler/main.go`.

## Data flow

```
.env + config.yaml
       │
       ▼
  main.go (cmd/reconciler)
       │
       ├─► splitwise.Client.GetCurrentUser()   →  validates the API key
       │
       ├─► splitwise.Client.GetExpenses()      →  []Expense  (deleted ones filtered out)
       │
       ├─► mapping.Mapper.MapAll()             →  []MappedExpense  (with Card assigned)
       │
       └─► report.Generate()                  →  output/report_YYYY-MM.xlsx
```

## Package responsibilities

| Package | Responsibility |
|---------|---------------|
| `cmd/reconciler` | Flag parsing, env/config loading, orchestration, stdout summary |
| `internal/splitwise` | Splitwise API communication, JSON deserialization |
| `internal/mapping` | `config.yaml` loading, card resolution per expense, share calculation |
| `internal/report` | Excel file construction with styles and three sheets |

## External dependencies

| Library | Purpose |
|---------|---------|
| `github.com/xuri/excelize/v2` | `.xlsx` file generation |
| `github.com/joho/godotenv` | `.env` file loading |
| `gopkg.in/yaml.v3` | `config.yaml` parsing |

The Splitwise HTTP client uses only `net/http` from the standard library — no third-party SDK.

## Design decisions

**`internal/` instead of exported packages.** This is a personal CLI tool; there is no reason for its packages to be importable externally. `internal/` enforces that at the compiler level.

**No caching or persistence.** Every run re-fetches from the API. Household monthly expense volume is low (< 500), so `limit=500` covers everything in a single call and latency is negligible.

**Monetary fields as `float64` internally.** The API returns amounts as strings (`"5000.00"`). They are parsed to `float64` at mapping time. For a personal report this is sufficient; if precise financial arithmetic were needed, `decimal` would be the right choice.

**No retries.** If the API fails, the error propagates and the process exits. The use case is interactive — the user runs the command and sees the error — so automatic retries add no value in this version.
