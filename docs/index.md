# Splitwise Card Reconciler — Documentación

Herramienta de línea de comandos en Go que consulta los gastos de un grupo de Splitwise, los mapea a las tarjetas de pago con las que fueron abonados, y genera un reporte mensual en Excel.

## Contenido

- [Arquitectura](architecture.md) — estructura de paquetes y flujo de datos
- [Configuración](configuration.md) — referencia completa de `config.yaml` y `.env`
- [Mapeo de tarjetas](card-mapping.md) — cómo se resuelve qué tarjeta corresponde a cada gasto
- [Referencia de paquetes](packages.md) — API interna de cada paquete Go

## Inicio rápido

```bash
# 1. Copiar y completar los archivos de configuración
cp .env.example .env          # agregar SPLITWISE_API_KEY
cp config.example.yaml config.yaml  # agregar group_id y my_user_id

# 2. Correr para el mes deseado
go run ./cmd/reconciler --month 2026-04

# 3. El reporte queda en:
#    output/reporte_2026-04.xlsx
```

## Requisitos

- Go 1.21 o superior
- Cuenta de Splitwise con acceso a la API (ver [Configuración](configuration.md#obtener-el-api-key))
