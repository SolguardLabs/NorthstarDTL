# API y operación local

## Arranque

```bash
npm ci
go run ./cmd/northstardtl serve --addr 127.0.0.1:8091
```

La API solo debe exponerse en loopback: `POST /route` modifica el catálogo en memoria y el servicio no añade una capa de autenticación. Reiniciar el proceso recupera el bootstrap inicial. Para escenarios reproducibles, ejecutar `go run ./cmd/northstardtl run tests/fixtures/settlement.json` y conservar el JSON resultante junto al SHA de la revisión.

| Ruta        | Método | Efecto                            |
| ----------- | ------ | --------------------------------- |
| `/snapshot` | GET    | Estado agregado del proceso       |
| `/capacity` | GET    | Proyección por ruta, sin mutación |
| `/quote`    | POST   | Cotizaciones de un intent válido  |
| `/submit`   | POST   | Reserva y entrada en cola         |
| `/execute`  | POST   | Ejecución de tickets listos       |
| `/route`    | POST   | Cambio operativo local de ruta    |

Ejemplo de lectura: `curl http://127.0.0.1:8091/capacity`. Para un cambio de comportamiento, revisar primero el fixture, los asientos y el snapshot, y después ejecutar `bash scripts/ci.sh`. El CI exige Go format, `go test`, `go vet`, `npm run check`, `npm test`, Prettier y métrica informativa de LOC. No usar la salida de `GET /capacity` como reserva de liquidez o confirmación de settlement.
