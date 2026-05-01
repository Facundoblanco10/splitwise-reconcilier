# Configuración

La herramienta se configura con dos archivos: `.env` para secretos y `config.yaml` para la lógica de negocio.

---

## `.env`

Contiene únicamente el API key de Splitwise. Está excluido del repositorio vía `.gitignore`.

```env
SPLITWISE_API_KEY=your_api_key_here
```

Si la variable está presente en el entorno del shell, el archivo `.env` es opcional: `godotenv` lo intenta cargar pero no falla si no existe.

### Obtener el API key

1. Ir a **https://secure.splitwise.com/apps**
2. Hacer clic en **Register your application** (o abrir una app existente)
3. Copiar el valor de **API Key**

---

## `config.yaml`

Controla el grupo a consultar, las tarjetas conocidas y las reglas de mapeo.

### Referencia completa

```yaml
splitwise:
  group_id: 12345678    # ID numérico del grupo de Splitwise
  my_user_id: 111       # ID numérico de tu usuario en Splitwise
```

**`group_id`** — Se encuentra en la URL al ver el grupo en Splitwise:
`https://secure.splitwise.com/#/groups/12345678`

**`my_user_id`** — Se puede obtener de la salida de la API:
```bash
curl -H "Authorization: Bearer <API_KEY>" \
     https://secure.splitwise.com/api/v3.0/get_current_user
```
El campo `id` en la respuesta es tu user ID. La herramienta también lo imprime al arrancar si agregás un log del struct `User`.

---

```yaml
cards:
  - id: VISA_GALICIA          # identificador interno (usado en rules y overrides)
    name: "Visa Galicia"      # nombre que aparece en el Excel
    closing_day: 25           # día de cierre del resumen (null para débito)
  - id: AMEX
    name: "American Express"
    closing_day: 5
  - id: DEBITO_SANTANDER
    name: "Débito Santander"
    closing_day: null
```

**`id`** — String en mayúsculas/guiones bajos. Es el identificador que usás en `rules_by_category`, `overrides_by_expense_id`, y en los tags `[CARD:ID]` dentro de Splitwise.

**`closing_day`** — Informativo por ahora; no afecta el cálculo. Sirve para saber a qué resumen pertenece un gasto según su fecha.

---

```yaml
rules_by_category:
  Groceries: VISA_GALICIA
  Utilities: DEBITO_SANTANDER
  Entertainment: AMEX
```

Mapea el **nombre de categoría de Splitwise** al `id` de la tarjeta. El nombre debe coincidir exactamente con el que usa Splitwise (sensible a mayúsculas). Ver [Mapeo de tarjetas](card-mapping.md) para la lógica completa de prioridades.

Categorías comunes de Splitwise (en inglés):

| Categoría Splitwise | Descripción |
|---------------------|-------------|
| `Groceries` | Supermercado |
| `Utilities` | Servicios (luz, gas, agua, internet) |
| `Rent` | Alquiler |
| `Entertainment` | Entretenimiento, salidas |
| `Electronics` | Electrónica |
| `General` | Categoría por defecto si no se especifica |

---

```yaml
overrides_by_expense_id:
  98765432: AMEX
  98765433: VISA_GALICIA
```

Fuerza la tarjeta para un gasto específico, ignorando cualquier regla o tag. El ID del gasto se puede ver en la URL de Splitwise al abrir el gasto, o en la columna **Expense ID** del Excel generado.

Tiene la mayor prioridad en la resolución. Ver [Mapeo de tarjetas](card-mapping.md).

---

## Flags del CLI

```
go run ./cmd/reconciler [flags]
```

| Flag | Por defecto | Descripción |
|------|-------------|-------------|
| `--month` | *(requerido)* | Mes a procesar, formato `YYYY-MM` (ej. `2026-04`) |
| `--config` | `config.yaml` | Ruta al archivo de configuración |

### Ejemplos

```bash
# Mes actual
go run ./cmd/reconciler --month 2026-05

# Mes anterior con config en otra ubicación
go run ./cmd/reconciler --month 2026-04 --config /home/user/mis-configs/splitwise.yaml

# Compilar y ejecutar binario
go build -o reconciler ./cmd/reconciler
./reconciler --month 2026-04
```
