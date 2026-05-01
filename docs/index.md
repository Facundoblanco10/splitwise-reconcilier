# Splitwise Card Reconciler — Documentation

A Go CLI tool that fetches expenses from a Splitwise household group, maps each one to the payment card it was charged to, and generates a monthly Excel report.

## Contents

- [Architecture](architecture.md) — package structure and data flow
- [Configuration](configuration.md) — full reference for `config.yaml` and `.env`
- [Card mapping](card-mapping.md) — how the tool resolves which card belongs to each expense
- [Package reference](packages.md) — internal API of each Go package

## Quick start

```bash
# 1. Copy and fill in the config files
cp .env.example .env                # add SPLITWISE_API_KEY
cp config.example.yaml config.yaml  # add group_id and my_user_id

# 2. Run for the desired month
go run ./cmd/reconciler --month 2026-04

# 3. Report is saved to:
#    output/report_2026-04.xlsx
```

## Requirements

- Go 1.21 or later
- A Splitwise account with API access (see [Configuration](configuration.md#getting-the-api-key))
