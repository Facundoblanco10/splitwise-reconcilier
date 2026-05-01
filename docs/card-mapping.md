# Mapeo de tarjetas

El proceso de asignar una tarjeta a cada gasto es el núcleo de la herramienta. Se implementa en `internal/mapping/mapper.go`.

## Orden de prioridad

Para cada gasto, se evalúan las siguientes fuentes en orden. La primera que produce un resultado gana; las demás no se consultan.

```
1. Override por ID de gasto   →  config.yaml › overrides_by_expense_id
2. Tag en el campo Notes      →  [CARD:ID] en el campo "details" del gasto
3. Regla por categoría        →  config.yaml › rules_by_category
4. Sin asignar                →  "UNASSIGNED"
```

### 1. Override por ID de gasto

```yaml
# config.yaml
overrides_by_expense_id:
  98765432: AMEX
```

Si el `id` del gasto aparece en este mapa, se usa esa tarjeta sin importar nada más. Útil para casos puntuales donde la categoría o el tag no aplican (ej. un gasto de supermercado pagado excepcionalmente con otra tarjeta).

### 2. Tag en el campo Notes (`details`)

En Splitwise, al agregar o editar un gasto, el campo **Notes** acepta texto libre. Si ese texto contiene el patrón:

```
[CARD:ID_DE_TARJETA]
```

la herramienta extrae el ID y lo usa como tarjeta. El ID debe estar en mayúsculas y puede contener letras, números y guiones bajos. El resto del campo Notes se preserva en el Excel.

**Ejemplos válidos:**

```
Compra en Disco [CARD:VISA_GALICIA]
[CARD:AMEX] cuotas electrodoméstico
ticket restaurant [CARD:DEBITO_SANTANDER] almuerzo trabajo
```

El tag puede estar en cualquier posición dentro del campo.

**Regex utilizada:** `\[CARD:([A-Z0-9_]+)\]`

Este mecanismo es la forma más cómoda de asignar tarjeta gasto a gasto desde la app de Splitwise, sin necesidad de tocar el config.

### 3. Regla por categoría

```yaml
# config.yaml
rules_by_category:
  Groceries: VISA_GALICIA
  Utilities: DEBITO_SANTANDER
```

Se compara el nombre de categoría del gasto (campo `category.name` de la API) con las claves del mapa. La comparación es exacta y sensible a mayúsculas.

Ideal para gastos que siempre van a la misma tarjeta según su naturaleza (ej. siempre pagás el supermercado con Visa).

### 4. Sin asignar (`UNASSIGNED`)

Si ninguna de las fuentes anteriores produce resultado, el gasto queda marcado como `UNASSIGNED`. Aparece en la hoja **"Sin asignar"** del Excel para revisión manual.

## Cómo resolver gastos sin asignar

Después de generar el reporte, revisá la hoja **"Sin asignar"**. Para cada gasto tenés dos opciones:

**Opción A — Tag en Splitwise (recomendada):**
Editá el gasto en Splitwise y agregá `[CARD:TU_TARJETA]` en el campo Notes. Al volver a correr el comando, ese gasto quedará asignado.

**Opción B — Override en config.yaml:**
Anotá el Expense ID de la columna correspondiente y agregalo en `overrides_by_expense_id`:

```yaml
overrides_by_expense_id:
  12345678: VISA_GALICIA
```

Volvé a correr el comando para regenerar el reporte.

## Cálculo de shares

Una vez asignada la tarjeta, se calculan dos montos por gasto:

- **Mi parte (`MyShare`)**: el campo `owed_share` del usuario con `my_user_id` en la lista de usuarios del gasto.
- **Parte pareja (`TheirShare`)**: suma de `owed_share` de todos los demás usuarios.

Estos valores los calcula Splitwise según el porcentaje de división configurado en el grupo; la herramienta los lee tal cual.

El monto total del gasto (`cost`) también se muestra en el Excel pero no se usa para cálculos internos.

## Ejemplo completo

Dado este gasto en Splitwise:

```json
{
  "id": 99001,
  "description": "Supermercado Disco",
  "cost": "8000.00",
  "date": "2026-04-10T18:00:00Z",
  "category": { "name": "Groceries" },
  "details": "compra semanal",
  "deleted_at": null,
  "users": [
    { "user_id": 111, "paid_share": "8000.00", "owed_share": "5000.00" },
    { "user_id": 222, "paid_share": "0.00",    "owed_share": "3000.00" }
  ]
}
```

Y con `my_user_id: 111` y la regla `Groceries: VISA_GALICIA`:

| Campo | Valor |
|-------|-------|
| Override por ID | no encontrado |
| Tag en Notes | no hay `[CARD:...]` |
| Regla por categoría | `Groceries` → `VISA_GALICIA` ✓ |
| Tarjeta asignada | `VISA_GALICIA` |
| Mi parte | 5000.00 |
| Parte pareja | 3000.00 |
