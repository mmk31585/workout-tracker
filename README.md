# Workout Tracker

A learning-first REST API for personal workout tracking. Users build workout plans
from exercises with target sets/reps/weight, schedule them for specific times, and
record the **actual** performance when they complete them. Reports summarize training
volume, completion rates, and exercise frequency over time.

This project is built phase-by-phase as a **senior backend engineering curriculum**. It
is deliberately well-scoped: a layered Go backend, a relational database with real
constraints, contract-first APIs, and honest testing. Every non-trivial design decision
is documented — either in the README, in an ADR, or in the design docs produced along
the way — because the *reasoning* is the actual deliverable, not the code.

---

## 1. Project overview

A multi-user workout tracker with full data isolation (User A can never see User B's
data). A user creates **workout plans** (templates) containing exercises from a fixed
catalog, each with target sets, reps, and weight (kg). They **schedule** a workout plan
for a specific datetime, and when they complete it they record what they **actually**
did — real sets, reps, and weight per exercise. The API exposes CRUD for workout plans,
a schedule/complete lifecycle, and SQL-backed reporting (volume, completion counts,
exercise frequency, progress over time).

Authentication is stateless: users sign up, log in, and receive a short-lived JWT. All
protected endpoints require it, and ownership is enforced at two layers (service +
repository) as defense in depth. Everything is deployed as a small Docker stack (API +
Postgres) that boots with one command.

---

## 2. What this project teaches

This project is a vehicle for internalizing **how a senior Go backend engineer thinks**,
mapped to explicit phases:

| Skill area | What you build / learn | Phases |
|---|---|---|
| Scoping & requirements | Turning a vague spec into functional + non-functional requirements and acceptance criteria | 0 |
| Domain modeling | Entities vs value objects, aggregates, association tables with attributes | 1 |
| Data modeling | Cardinality, normalization, constraints, index justification, FK `ON DELETE` policy | 2, 6 |
| State machines | Workout lifecycle as an explicit typed state machine, not scattered `if`s | 3 |
| Security & auth | JWT structure/claims, bcrypt, BOLA/IDOR prevention, defense in depth | 4, 5, 14, 15 |
| Time handling | UTC-only storage, `timestamptz`, timezone bugs and why local time is wrong | 7 |
| API design | Contract-first REST, standard error envelopes, OpenAPI 3 | 8, 9, 19 |
| Architecture | Layered architecture, repository pattern, dependency inversion, DTO vs domain separation | 10, 11 |
| Database engineering | Versioned migrations, idempotent seeders, SQL aggregation/window functions | 12, 13, 17 |
| Go idioms | `database/sql` transactions, `%w` error wrapping, table-driven tests, `internal/`, `slog` | 13, 9, 11, 18 |
| Observability | Structured logging with request correlation IDs — and knowing what *not* to build | 18 |
| Process | ADRs, definition of done, scoping before designing, designing before coding | 21 |

---

## 3. Tech stack

| Component | Choice | Why |
|---|---|---|
| Language | Go 1.26.6 | Standard library ergonomics, explicit dependency injection, no framework magic, first-class concurrency — the language where "idiomatic" has an accepted shape. |
| Router | `chi` | Tiny, idiomatic, production-standard. Route groups, method+path patterns, and a clean middleware chain without the gravity of a full framework. |
| HTTP | `net/http` | chi is a thin layer on stdlib; we keep stdlib handlers, `ResponseWriter`/`Request`, and middleware wrapping. |
| Database | PostgreSQL 15 (`postgres:15-alpine`) | Real relational integrity: FK constraints, `UNIQUE`/`CHECK`, `timestamptz`, transactions. The constraint story *is* the lesson — an ORM would hide it. |
| SQL access | `database/sql` + `pgx/v5` stdlib driver | Hand-written SQL is a deliberate learning goal (indexes, `GROUP BY`, window functions). pgx v5 via its stdlib adapter gives us modern Postgres support under `database/sql`. |
| Migrations | `goose` | Simple, reliable migration tool for PostgreSQL. |
| Auth | `golang-jwt/jwt/v5` + `golang.org/x/crypto/bcrypt` | JWT HS256 for stateless auth; bcrypt (cost 12) for password hashing — one cost factor, ubiquitous, hard to misconfigure. argon2id is noted as a v2 upgrade. |
| Config | Plain env vars, small `config` package | 12-factor config. Six environment variables do not justify a config library. |
| Logging | stdlib `log/slog` | Structured, leveled, JSON output — built into the standard library, zero deps. |
| Testing | stdlib `testing` | Table-driven unit tests; integration tests against real Postgres. |
| Packaging | `docker-compose` (API + Postgres) | One-command local boot; multi-stage Dockerfile for the API image. |
| Docs | OpenAPI 3 (`openapi.yaml`) | Contract-first API surface, machine-validated, drifts detectable against implementation. |

**Deliberately excluded** (and why): ORMs (you must write the SQL to learn the schema),
Redis (not needed; a stateless auth design), message queues (no async workload),
Prometheus/Grafana (observability is slog-level for this project), microservices (a
single bounded domain does not justify distribution), and any web UI (the API is the
deliverable).

---

## 4. Architecture

Layered, with strict dependency direction: **handlers never import repositories;
repositories never import handlers; services never import `net/http`.**

```
                         ┌────────────────────────────┐
                         │          Client            │
                         │   (curl / any HTTP client) │
                         └────────────┬───────────────┐
                                      │ JSON over HTTP
                         ┌────────────▼───────────────┐
                         │        chi router          │
                         │   middleware chain:        │
                         │   requestID → slog → auth  │
                         └────────────┬───────────────┐
                                      │
                         ┌────────────▼───────────────┐
                         │        Handler layer       │  transport concerns ONLY:
                         │   request parsing/DTOs,    │  status codes, error envelope
                         │   validation, auth wiring  │
                         └────────────┬───────────────┐
                                      │
                         ┌────────────▼───────────────┐
                         │        Service layer       │  domain logic, ownership rules,
                         │   business rules, state    │  transactions boundaries
                         │   machine transitions      │
                         └────────────┬───────────────┐
                                      │  repository interfaces
                                      │  (owned by services)
                         ┌────────────▼───────────────┐
                         │      Repository layer      │  database/sql + pgx,
                         │   queries ALWAYS scoped     │  SQL, tx handling
                         │   by user_id (defense)     │
                         └────────────┬───────────────┐
                                      │
                         ┌────────────▼───────────────┐
                         │         PostgreSQL 15       │
                         │  migrations · constraints   │
                         │  · seeder (idempotent)      │
                         └────────────────────────────┘
```

**Auth request flow** (Phase 4 design, previewed here):

```
Request ──► JWT middleware ──► parse + verify token ──► extract user_id ──►
   attach identity to context ──► handler ──► service enforces ownership ──►
   repository scopes every query by user_id (defense in depth)
```

---

## 5. Domain model

Entities and the one-liner for each:

| Entity | One-liner |
|---|---|
| **User** | Account holder; owns all of their workout plans, scheduled workouts, and workout sessions; auth identity. |
| **Exercise** | Global catalog item (name, description, category, muscle group). Seeded; not user-created in v1. |
| **WorkoutPlan** | Reusable routine (e.g., "Push Day"): a named collection of exercises with target sets/reps/weight. |
| **WorkoutPlanItem** | Association table carrying *target* sets/reps/weight between a WorkoutPlan and an Exercise. |
| **ScheduledWorkout** | A planned run of a WorkoutPlan at a specific datetime; owns the lifecycle state machine. |
| **WorkoutSession** | A performed run of a WorkoutPlan (created when a scheduled workout is completed); the actual performance record. |
| **WorkoutSessionItem** | Association table carrying *actual* sets/reps/weight between a WorkoutSession and an Exercise. |

Relationships at a glance (full cardinality analysis in Phase 2):

| Relationship | Cardinality | Why |
|---|---|---|
| User → WorkoutPlan | 1:N | A user owns many workout plans. |
| User → ScheduledWorkout | 1:N | A user schedules many planned runs. |
| User → WorkoutSession | 1:N | A user performs many workout sessions. |
| WorkoutPlan ↔ Exercise | N:M via WorkoutPlanItem | Many exercises per plan; the same exercise appears in many plans. |
| WorkoutPlan → ScheduledWorkout | 1:N | A plan can be scheduled many times. |
| ScheduledWorkout → WorkoutSession | 1:1 (optional) | A scheduled run becomes at most one performed session. |
| WorkoutSession ↔ Exercise | N:M via WorkoutSessionItem | Actuals recorded per exercise per session. |

All timestamps stored as `timestamptz` in UTC; conversion to the user's local zone happens
only at the presentation layer.

---

## 6. API surface

Base path `/api/v1`. All resources are **scoped to the authenticated user** — a user can
only read/write their own workout plans, scheduled workouts, and workout sessions
(enforced at service + repository layers).

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `POST` | `/api/v1/auth/signup` | Public | Create an account; returns a JWT. |
| `POST` | `/api/v1/auth/login` | Public | Verify credentials; returns a JWT. |
| `POST` | `/api/v1/auth/logout` | JWT | Stateless logout: client discards the token. |
| `GET` | `/api/v1/exercises` | JWT | List the global exercise catalog. |
| `GET` | `/api/v1/exercises/{id}` | JWT | Get a specific exercise by ID. |
| `POST` | `/api/v1/workout-plans` | JWT | Create a workout plan with exercises + targets. |
| `GET` | `/api/v1/workout-plans` | JWT | List the user's workout plans. |
| `GET` | `/api/v1/workout-plans/{id}` | JWT | Get one workout plan (owned by caller). |
| `PUT` | `/api/v1/workout-plans/{id}` | JWT | Replace a workout plan (idempotent). |
| `DELETE` | `/api/v1/workout-plans/{id}` | JWT | Delete a workout plan. |
| `POST` | `/api/v1/workout-plans/{id}/schedule` | JWT | Schedule a workout plan for a datetime. |
| `GET` | `/api/v1/scheduled-workouts` | JWT | List the user's scheduled workouts. |
| `GET` | `/api/v1/scheduled-workouts/{id}` | JWT | Get one scheduled workout (owned by caller). |
| `POST` | `/api/v1/scheduled-workouts/{id}/complete` | JWT | Complete a scheduled workout, recording **actual** performance per exercise. |
| `POST` | `/api/v1/scheduled-workouts/{id}/cancel` | JWT | Cancel a scheduled workout. |
| `GET` | `/api/v1/workout-sessions/{id}` | JWT | Get a workout session (owned by caller). |
| `GET` | `/api/v1/reports/progress` | JWT | Progress report: exercise frequency, volume trends. |
| `GET` | `/api/v1/reports/summary` | JWT | Summary report: total plans, completion rates, etc. |

All responses use a consistent error envelope:
`{ "error": { "code": "WORKOUT_NOT_FOUND", "message": "Workout not found" } }`

---

## 7. Setup

Prerequisites: Go 1.26.6+, Docker (for the dev database and integration tests). No local
Postgres install is required — everything runs in containers.

```bash
# 1. Configure environment
cp .env.example .env        # edit values (see below)

# 2. Start Postgres
docker compose up -d postgres

# 3. Run migrations
make migrate                # or: go run ./cmd/api -migrate

# 4. Seed the exercise catalog (idempotent — safe to re-run)
make seed                   # or: go run ./cmd/seed/main.go

# 5. Run the API
make run                    # or: go run ./cmd/api
```

Environment variables (in `.env`):

| Variable | Purpose |
|---|---|
| `DB_ADDR` | Postgres DSN (`postgres://...`). |
| `JWT_SECRET` | HMAC secret for token signing — from env/secret manager, **never hardcoded or logged**. |
| `JWT_EXPIRATION` | JWT expiration time (e.g., 1h). |
| `HTTP_PORT` | Listen port, e.g. `8080`. |
| `HTTP_HOST` | Listen host, e.g. `0.0.0.0`. |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error`. |

Docker-only boot (everything including migrations):

```bash
docker compose up --build
```

Healthcheck: `GET /healthz` reports API liveness; the container healthcheck verifies
Postgres connectivity before the API reports healthy.

---

## 8. Testing strategy

Two classes, run differently:

- **Unit tests** (no Docker): service-layer business logic, the workout state machine,
  bcrypt hashing/verification, JWT generation/validation (valid/expired/malformed/tampered),
  DTO validation, error → HTTP mapping. Table-driven throughout.
  ```bash
  go test ./... -short
  ```

- **Integration tests** (require Docker): each test spins up a disposable Postgres via
  direct connection (Docker must be running). Covers repository CRUD, constraint
  violations, transaction rollback, handler request→DB round trips, migration up/down,
  seeder run-twice idempotency, and the **cross-user isolation suite** (User A cannot
  read/update/delete User B's resources — the most important test class in this project).
  ```bash
  go test ./...
  ```

Full gate before phase sign-off: `go test ./...`, `go vet ./...`, and `gofmt -l .` all clean.

---

## 9. Roadmap

The project proceeds through 22 phases; each phase produces working code, tests, a
"what you just learned" recap, and a checklist. Full details live in the phased plan;
condensed:

| Phase | Deliverable |
|---|---|
| 0 | Scoping & requirements (`docs/requirements.md`) |
| 1 | Domain analysis: entities vs value objects (`docs/domain.md`) |
| 2 | Relationships & cardinality |
| 3 | Workout lifecycle state machine & legal transitions |
| 4 | Auth & authorization design (`docs/authentication.md`) |
| 5 | Security baseline checklist |
| 6 | Database design — ERD, constraints, index justification |
| 7 | Time & timezone ADR (UTC `timestamptz`) |
| 8 | API contract design (`docs/api.md`) |
| 9 | Standard error model |
| 10 | Architecture decision: layers, auth placement |
| 11 | Package structure & scaffolding, config, logger, DB pool |
| 12 | Migrations & idempotent seeder |
| 13 | Repository layer + transaction boundaries |
| 14 | Auth implementation (signup/login/JWT middleware) |
| 15 | Authorization enforcement (BOLA prevention) |
| 16 | Workout plan CRUD + scheduling + completion |
| 17 | Reporting (SQL aggregation, progress over time) |
| 18 | Observability (structured request logs) |
| 19 | OpenAPI 3 spec |
| 20 | Docker & one-command dev environment |
| 21 | Final review & ADRs; full verification gate |

---

## 10. Glossary of concepts you'll learn

- **JWT** — a signed token with header/payload/signature; the payload's *claims* carry the
  user identity. Stateless: the server can verify without storing session state.
- **Stateless auth** — trust derived solely from a verifiable token, not server-side session
  memory. Tradeoff: you can't revoke a token server-side without extra machinery (a known
  gap, documented in Phase 4).
- **bcrypt** — deliberately slow, salted password hash; the cost factor trades CPU for
  resistance to brute force. Passwords are hashed at rest and never returned in any
  response.
- **BOLA / IDOR** — "Broken Object Level Authorization": accessing another user's object by
  changing an ID. Prevented here by scoping every query by `user_id` (defense in depth).
- **Association table with attributes** — a join table carrying extra columns
  (`WorkoutPlanItem`, `WorkoutSessionItem`): target/actual sets/reps/weight. The same
  pattern as order line items and playlist tracks.
- **Normalization** — organizing columns to eliminate redundancy (1NF–3NF); you'll see why
  at Phase 6, and where denormalization is *not* yet justified.
- **Indexes** — B-trees that speed up specific queries; every index you add must be
  justified by a real query and an expected `EXPLAIN ANALYZE` improvement.
- **Transaction / rollback** — grouping multi-step writes (create workout plan + its
  exercise rows) so all-or-nothing holds; the failure-path test is the one that proves it.
- **`timestamptz` / UTC** — storing instants in UTC and converting only at presentation;
  the fix for a whole class of off-by-one timezone bugs.
- **Contract-first API** — specifying every request/response/error on paper before writing
  handlers, so the API is the spec and clients can build against it in parallel.
- **Error envelope** — one consistent error shape across the API; domain errors become
  transport errors *only* at the handler boundary.
- **Layered architecture** — handler → service → repository; each layer has one job and
  imports only the layer below.
- **Repository pattern** — data access behind interfaces owned by the service layer;
  swapping Postgres for anything else is a swap of one implementation.
- **Dependency inversion** — high-level layers define the interfaces they need; low-level
  layers implement them. Nothing depends on `net/http` below the handler layer.
- **DTO vs domain model** — request/response shapes are not the same structs as domain
  objects; they diverge and must be allowed to.
- **`database/sql` transactions** — `BeginTx`/`Commit`/`Rollback` with context
  propagation; the mechanism behind the transaction boundary requirement.
- **Error wrapping (`%w`)** — `fmt.Errorf("...: %w", err)` to preserve the error chain for
  `errors.Is`/`errors.As`.
- **Table-driven tests** — a slice of test cases iterated in one test function; the
  idiomatic Go test shape used across this project.
- **`internal/` visibility** — Go's mechanism for "this package is not importable from
  outside the module"; the skeleton of our package-by-domain layout.
- **Structured logging** — machine-parseable key/value log lines (request id, method,
  path, status, duration) via `slog`; correlation IDs tie a request across layers.
- **ADR** — Architecture Decision Record: a short "context → decision → consequences"
  note that captures *why*, so settled decisions are never relitigated.

---

## 11. Definition of Done (project-level)

- All functional requirements implemented and manually verified.
- No user can read/write another user's data — verified by tests, not assumption.
- Passwords hashed; never returned in any response.
- JWT validated correctly for valid / expired / malformed / missing cases.
- All DB constraints exist and are enforced.
- Transactions exist wherever multi-step writes require atomicity.
- All timestamps stored in UTC; sorting/comparison behavior tested.
- Seeder is idempotent (run-twice test).
- Reports backed by real SQL, verified against hand-computed expected values.
- Unit tests for all service logic; integration tests for all repository + handler
  behavior.
- OpenAPI spec complete and matching implementation.
- `docker compose up` boots a fully working stack from scratch.
- Standard error format across all endpoints; structured logging on every request.
- README complete enough for a stranger to run and understand the project.
- Key ADRs written; `go test ./...`, `go vet ./...`, `gofmt -l .` all clean.

---

#Roadmap-project 
    https://roadmap.sh/projects/fitness-workout-tracker