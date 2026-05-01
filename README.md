# Splitwise Card Reconciler

Splitwise is great for splitting household expenses with your partner. But it has no idea which card you used to pay for each thing.

At the end of the month, when the statements arrive, you end up doing the math by hand: "how much of this is on the Visa? how much on the debit card?". This project solves that.

## What it does

It connects to the Splitwise API, pulls the month's expenses from your household group, and assigns each expense to the card you paid with — either by category rule, by a tag you drop into the expense notes, or by a manual override. Then it generates a `.xlsx` with three sheets: the full expense detail, a per-card summary, and any unassigned expenses for manual review.

```
go run ./cmd/reconciler --month 2026-04
```

```
[INFO] Authenticated as Facundo Blanco (fblanco@...)
[INFO] Fetched 34 active expenses

=== Resumen 2026-04 ===
Total gastos procesados: 34

Por tarjeta (mi parte):
  Visa Galicia               48320.00
  American Express           12750.50
  Débito Santander            9100.00

  Sin asignar (2 gastos):    3200.00

Reporte generado: output/reporte_2026-04.xlsx
```

## Stack

Go · `excelize` · `godotenv` · `yaml.v3` · `net/http`

## Documentation

Full technical documentation lives in [`/docs`](docs/index.md):

- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Card mapping](docs/card-mapping.md)
- [Package reference](docs/packages.md)

## Quick start

```bash
cp .env.example .env                # add your SPLITWISE_API_KEY
cp config.example.yaml config.yaml  # add group_id, my_user_id and your cards
go run ./cmd/reconciler --month 2026-04
```
