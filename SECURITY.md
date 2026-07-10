# Security Model - NorthstarDTL

NorthstarDTL modela un router de liquidez con admisión económica, reservas de
fondos y ejecución diferida. La revisión de seguridad debe cubrir consistencia
entre rutas, pricing, límites de exposición, colas de settlement y saldos del
ledger.

## Invariantes Esperadas

- Cada ticket admitido tiene una reserva suficiente en el asset fuente.
- Cada ruta mantiene exposición menor o igual a `maxExposure`.
- La liquidez de salida de una ruta no debe quedar negativa.
- Los recibos deben corresponder a tickets existentes y ejecutados una sola vez.
- Los balances disponibles y reservados no pueden ser negativos.
- Las rutas pausadas no deben admitir nuevos tickets.

## Superficie De Revisión

- `routing.Selector` y `routing.Scorer` para ranking de rutas.
- `settlement.Engine` para transición de ticket y contabilización.
- `ledger.Book` para reservas, consumo y transferencias internas.
- `api.Service` para snapshots y cambios operativos de ruta.
- Fixtures JSON en `tests/fixtures` para flujos reproducibles.

## Validación Automatizada

```bash
go test ./...
npm run check
npm test
npm run loc
```

El CI ejecuta formato Go, tests Go, `go vet`, typecheck TypeScript, tests
TypeScript y conteo de LOC de `src/`.

## Dependencias

El proyecto no usa dependencias Go externas. La capa TypeScript usa únicamente
herramientas de test y typecheck de desarrollo. Dependabot cubre `gomod`, `npm`
y GitHub Actions.

## Reporte Interno

Un reporte debe incluir:

- escenario reproducible;
- fixture o secuencia de llamadas;
- rutas y epochs afectados;
- impacto económico estimado;
- propuesta de test de regresión;
- cambio mínimo sugerido.
