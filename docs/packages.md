# Referencia de paquetes

## `internal/splitwise`

Maneja toda la comunicación con la API REST de Splitwise.

**Base URL:** `https://secure.splitwise.com/api/v3.0`  
**Autenticación:** header `Authorization: Bearer <API_KEY>`

---

### `Client`

```go
func NewClient(apiKey string) *Client
```

Crea un cliente HTTP listo para usar. Internamente instancia un `*http.Client` con configuración por defecto (sin timeout explícito; la herramienta es interactiva y el volumen de datos es pequeño).

---

### `Client.GetCurrentUser() (*User, error)`

Llama a `GET /get_current_user`. Se usa al arrancar para validar que el API key es correcto antes de hacer cualquier otra cosa.

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

Llama a `GET /get_groups`. Útil para listar los grupos disponibles y encontrar el `group_id` correcto. No se usa en el flujo principal actualmente.

```go
type Group struct {
    ID   int
    Name string
}
```

---

### `Client.GetExpenses(groupID int, from, to time.Time) ([]Expense, error)`

Llama a `GET /get_expenses` con los query params:

| Param | Valor |
|-------|-------|
| `group_id` | ID del grupo |
| `dated_after` | `from` en formato RFC3339 |
| `dated_before` | `to` en formato RFC3339 |
| `limit` | `500` (fijo) |

Filtra automáticamente los gastos con `deleted_at != null` antes de retornar.

```go
type Expense struct {
    ID          int
    Description string
    Cost        string       // string en la API, ej. "5000.00"
    Date        string       // RFC3339
    Details     string       // campo Notes de Splitwise
    DeletedAt   *string      // nil si el gasto está activo
    Category    Category
    Users       []ExpenseUser
}

type Category struct {
    ID   int
    Name string
}

type ExpenseUser struct {
    UserID    int
    PaidShare string   // cuánto pagó este usuario
    OwedShare string   // cuánto le corresponde pagar
}
```

Los campos monetarios vienen como `string` desde la API y se parsean a `float64` en el paquete `mapping`.

---

## `internal/mapping`

Carga la configuración y aplica la lógica de asignación de tarjetas.

---

### `LoadConfig(path string) (*Config, error)`

Lee y parsea el archivo `config.yaml`. Retorna error si el archivo no existe o tiene sintaxis YAML inválida.

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
    ClosingDay *int    // nil para tarjetas de débito
}
```

---

### `Mapper`

```go
func NewMapper(cfg *Config) *Mapper
func (m *Mapper) MapAll(expenses []splitwise.Expense) []MappedExpense
```

`MapAll` itera sobre todos los gastos y aplica `mapOne` a cada uno, que a su vez llama a `resolveCard` y `resolveShares`.

```go
type MappedExpense struct {
    Expense    splitwise.Expense
    Card       string    // ID de tarjeta o "UNASSIGNED"
    MyShare    float64
    TheirShare float64
}
```

---

### `FormatCardName(cfg *Config, cardID string) string`

Convierte un `cardID` interno al nombre para mostrar en el Excel. Si el ID no está en el config retorna `(ID)` entre paréntesis. Para `UNASSIGNED` retorna `"Sin asignar"`.

---

### `ParseCost(s string) float64`

Parsea un string monetario de la API (`"5000.00"`) a `float64`. Retorna `0` si el string no es un número válido.

---

### Constante `Unassigned`

```go
const Unassigned = "UNASSIGNED"
```

Valor centinela que indica que no se encontró tarjeta para el gasto.

---

## `internal/report`

Genera el archivo Excel con tres hojas.

---

### `Generate(expenses []MappedExpense, cfg *Config, month time.Time, outputPath string) error`

Punto de entrada único del paquete. Crea el archivo en `outputPath` con las tres hojas descritas abajo. Si el archivo ya existe, lo sobreescribe.

---

### Hojas generadas

#### "Gastos detallados"

Una fila por gasto. Columnas:

| Columna | Contenido |
|---------|-----------|
| Fecha | `YYYY-MM-DD` |
| Descripción | `expense.Description` |
| Categoría | `expense.Category.Name` |
| Monto total | `expense.Cost` (float, formato `#,##0.00`) |
| Mi parte | `MappedExpense.MyShare` |
| Parte pareja | `MappedExpense.TheirShare` |
| Tarjeta | nombre display de la tarjeta |
| Expense ID | `expense.ID` |
| Notes | `expense.Details` (campo original de Splitwise) |

#### "Resumen por tarjeta"

Dos columnas: nombre de tarjeta y suma de `MyShare` de todos los gastos asignados a ella. El orden sigue el definido en `config.yaml › cards`; `UNASSIGNED` va siempre al final.

#### "Sin asignar"

Mismas columnas que "Gastos detallados" excepto "Parte pareja" y "Tarjeta" (que son irrelevantes). Solo incluye gastos donde `Card == "UNASSIGNED"`.

---

### Estilos aplicados

| Elemento | Estilo |
|----------|--------|
| Headers de todas las hojas | Negrita + fondo azul claro (`#D9E1F2`) |
| Columnas monetarias | Formato `#,##0.00` (NumFmt 7) |
| Anchos de columna | Ajustados por tipo de contenido (fechas: 12, descripciones: 35, etc.) |
