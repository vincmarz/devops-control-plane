# PostgreSQL Integration Notes

## Stato di questo step

Questo step sostituisce il database placeholder con una connessione PostgreSQL reale basata su `pgxpool`.

## File modificati o aggiunti

```text
go.mod
Makefile
cmd/devops-control-plane/main.go
cmd/devops-control-plane-migrate/main.go
internal/database/postgres.go
internal/database/repositories.go
docs/postgresql-integration-notes.md
```

## Variabile obbligatoria

```bash
export DATABASE_URL="postgres://devops_cp:devops_cp@localhost:5432/devops_control_plane?sslmode=disable"
```

## Comandi di validazione

```bash
go mod tidy
go test ./...
make migrate-up
make run
```

## Verifica readiness

```bash
curl -i http://localhost:8080/readyz
```

Output atteso:

```json
{"data":{"checks":{"configuration":"ok","database":"ok"},"status":"ready"},"error":null}
```

## Nota importante

Da questo step in poi `go run ./cmd/devops-control-plane` fallisce se PostgreSQL non è raggiungibile o se `DATABASE_URL` non è configurata.
