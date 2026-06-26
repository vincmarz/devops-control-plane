# Phase 1 - PostgreSQL ChangeRequest repository

## Obiettivo

Sostituire la persistenza in-memory delle ChangeRequest con repository PostgreSQL reali per:

- `change_requests`
- `change_events`

## Endpoint validati attesi

```text
POST /api/v1/changes
GET  /api/v1/changes
GET  /api/v1/changes/{id}
GET  /api/v1/changes/{id}/events
POST /api/v1/changes/{id}/validate
```

## Nota

L'ID restituito da questo step è l'UUID PostgreSQL della tabella `change_requests`.
Il campo `changeNumber` rimane l'identificativo funzionale leggibile, ad esempio `CHG-2026-0001`.
