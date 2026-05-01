# Arquitectura

## Estructura de directorios

```
splitwise-reconcilier/
├── cmd/
│   └── reconciler/
│       └── main.go              # Entrypoint y orquestación del flujo
├── internal/
│   ├── splitwise/
│   │   ├── client.go            # Cliente HTTP con autenticación Bearer
│   │   ├── models.go            # Structs que representan la respuesta de la API
│   │   └── expenses.go          # GetExpenses con filtros de fecha y grupo
│   ├── mapping/
│   │   ├── rules.go             # Config struct y carga del config.yaml
│   │   └── mapper.go            # Lógica de asignación de tarjeta por gasto
│   └── report/
│       └── excel.go             # Generación del .xlsx con tres hojas
├── docs/                        # Esta documentación
├── output/                      # Reportes generados (ignorado por git)
├── config.example.yaml
├── .env.example
└── go.mod
```

Los paquetes viven bajo `internal/` para que no sean importables desde fuera del módulo. El único punto de entrada público es `cmd/reconciler/main.go`.

## Flujo de datos

```
.env + config.yaml
       │
       ▼
  main.go (cmd/reconciler)
       │
       ├─► splitwise.Client.GetCurrentUser()   →  valida el API key
       │
       ├─► splitwise.Client.GetExpenses()      →  []Expense  (filtrados: sin deleted_at)
       │
       ├─► mapping.Mapper.MapAll()             →  []MappedExpense  (con Card asignada)
       │
       └─► report.Generate()                  →  output/reporte_YYYY-MM.xlsx
```

## Responsabilidades por paquete

| Paquete | Responsabilidad |
|---------|----------------|
| `cmd/reconciler` | Parseo de flags, carga de env/config, orquestación, impresión del resumen |
| `internal/splitwise` | Comunicación con la API de Splitwise, deserialización JSON |
| `internal/mapping` | Carga de `config.yaml`, resolución de tarjeta por gasto, cálculo de shares |
| `internal/report` | Construcción del archivo Excel con estilos y tres hojas |

## Dependencias externas

| Librería | Uso |
|----------|-----|
| `github.com/xuri/excelize/v2` | Generación del archivo `.xlsx` |
| `github.com/joho/godotenv` | Carga del archivo `.env` |
| `gopkg.in/yaml.v3` | Parseo de `config.yaml` |

El cliente HTTP de Splitwise usa únicamente `net/http` de la stdlib, sin SDK de terceros.

## Decisiones de diseño

**`internal/` en lugar de paquetes exportados.** La herramienta es un CLI de uso personal; no hay razón para que los paquetes sean importables externamente. `internal/` lo refuerza a nivel del compilador.

**Sin caché ni persistencia.** Cada ejecución consulta la API de nuevo. Dado que el volumen de gastos mensuales de un hogar es bajo (< 500), el `limit=500` cubre todo en una sola llamada y la latencia es despreciable.

**Campos monetarios como `float64` internamente.** La API devuelve montos como strings (`"5000.00"`). Se parsean a `float64` al momento de mapear. Para un reporte personal esto es suficiente; si se necesitara aritmética financiera precisa habría que usar `decimal`.

**Sin reintentos.** Si la API falla, el error se propaga y el proceso termina. El caso de uso es interactivo (el usuario corre el comando y ve el error), por lo que los reintentos automáticos no agregan valor en esta versión.
