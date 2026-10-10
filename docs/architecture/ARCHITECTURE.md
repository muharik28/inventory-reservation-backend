# Architecture — Inventory Reservation Backend

Stack: **Go 1.27 + Gin + PostgreSQL 16** (Bun ORM / pgdriver), deployed via Docker Compose as three services: `app` (REST API), `scheduler` (expiry worker), `postgres`.

## 1. Schema design, routing, locking, and transaction strategy

**Schema** (`migrations/20261010_init.sql`)

- `inventories(item_id PK, total_stock, created_at, updated_at)` with `CHECK (total_stock >= 0)`.
- `reservations(reservation_id PK, user_id, item_id FK → inventories ON DELETE CASCADE, quantity, status, expires_at, confirmed_at, ...)` with `CHECK (quantity > 0)`.
- Indexes: `(status, expires_at)` for the expiry worker scan and active-reservation aggregation, `(item_id)` for per-item lookups.

**Stock model.** `total_stock` is the only persisted stock number; it decreases only on confirmation. Reserved stock is *derived*, never stored:

```
reserved_stock  = SUM(quantity) FROM reservations
                  WHERE item_id = ? AND status = 'PENDING' AND expires_at > NOW()
available_stock = total_stock - reserved_stock
```

`NOW()` is always the PostgreSQL clock, fetched inside the transaction. Because availability is a function of committed rows, an expired reservation stops reducing stock **immediately** — correctness does not depend on the worker's schedule. The worker only performs bookkeeping (flip `PENDING → EXPIRED`).

**Layering.** `routes → handlers (HTTP/DTO) → use cases (business rules + transactions) → repositories (Bun queries)`. Transactions are opened through a `TxManager` port (consumer-defined interface, `BunTxManager` implementation): the use case never touches `*bun.DB`, and repositories join the ambient transaction from `context`. Domain entities carry the rules (`CanReserve`, `IsPending`, `IsExpired`); handlers translate domain errors to HTTP + machine-readable `code` (`INSUFFICIENT_STOCK`, `RESERVATION_EXPIRED`, ...).

**Concurrency strategy.**

- *Reserve* — single transaction: `SELECT ... FOR UPDATE` on the inventory row (serializes all reservations per item), compute active reserved with the DB clock, reject if `quantity > available`, insert the `PENDING` reservation with `expires_at = now() + 5m`, commit. Concurrent overselling is impossible: the second transaction always reads the first one's committed insert under the row lock.
- *Confirm* — single transaction: lock the reservation row → verify `PENDING` and not expired (DB clock) → lock the inventory row → `total_stock -= quantity` → set `CONFIRMED` + `confirmed_at`. Lock order is consistent across flows (reservation row, then inventory row; reserve touches only the inventory row plus a fresh insert), so there is no lock cycle and no deadlock path.
- Database CHECK constraints are the last line of defense against invalid quantities, even under bugs.

## 2. Ten instances sharing PostgreSQL

The service is already stateless: no in-memory session/timer decides reservation validity — only PostgreSQL state and the DB clock do. Running 10 API replicas behind a load balancer works unchanged. What changes and what becomes a bottleneck:

- **Connection pressure.** 10 × 25 pool connections can exceed `max_connections`. Put **PgBouncer** (transaction pooling) in front, or shrink per-instance pools.
- **Per-item lock contention.** The `FOR UPDATE` row lock serializes reservations *per item*. A single hot item becomes the throughput ceiling (in exchange for strict oversell protection). Mitigations: collapse the two-step check into one atomic guarded statement (`UPDATE ... WHERE total_stock - active_reserved >= qty`), batch with `SKIP LOCKED`, or shard hot items.
- **Expiry worker.** Run it as a singleton. If it must be replicated, take a PostgreSQL advisory lock (`pg_try_advisory_lock`) per tick for leader election — safe because it is idempotent bookkeeping anyway.
- **Clock trust.** All TTL math uses the DB clock, so replica clock skew cannot extend or shorten reservations.
- **Optional next steps:** idempotency keys for client retries, read replicas for stock dashboards (accepting small staleness), and per-API-key rate limiting.

## 3. Timebox trade-offs (Go vs React) and a corrected AI suggestion

Backend was weighted 80%, so the timebox went to correctness first: atomic reserve/confirm, DB-clock expiry, constraints, integration + concurrency tests (`go test -race -v ./...`), Swagger docs, and a clean two-command deployment. Non-essential backend ideas (authentication, pagination, rate limiting, metrics) were cut. The React + Vite dashboard was the sacrifice of the timebox: the API supports every workflow (stock read, reserve, confirm, clear error codes for expired/insufficient states), but the UI itself was not built within the time limit.

**Truthful example of a flawed AI suggestion and the fix.** An AI-assisted draft stored a persisted `reserved_stock` counter on `inventories`, incremented in the reserve transaction and decremented by the expiry worker and on confirmation. It passed naive tests, but it silently violated the expiry requirement: in the window between `expires_at` passing and the worker's next tick, the stale counter still reduced availability and could reject a legitimate reservation. The fix was to remove the counter column entirely and derive reserved stock from the reservation rows (`SUM(quantity) ... WHERE status = 'PENDING' AND expires_at > NOW()`) inside the row-locked transaction using the database clock — making availability correct at any instant, with the worker reduced to idempotent status cleanup. The trade-off is one indexed aggregate query per reserve/stock read, which is acceptable at this scale and was validated by the concurrency tests (50 parallel reservations against 10 units: exactly 10 succeed, zero oversell, zero negative stock).
