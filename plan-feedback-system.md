# Plan: Feedback System

## Metadata
- **Date**: 2026-03-21
- **Goal**: Add a delightful, non-intrusive feedback widget to the results screen that collects user satisfaction, qualitative feedback, and tracks shopping link clicks -- enabling product validation for the soft launch
- **Scope**: Feedback widget UI, feedback submission API endpoint, shopping link click tracking, Neon PostgreSQL storage via GORM, database migrations | **Out of Scope**: Analytics dashboard/admin UI, A/B testing framework, user accounts, localStorage/sessionStorage usage, push notifications, email follow-ups
- **Affected Areas**: `giftsense-frontend/src/components/results/`, `giftsense-frontend/src/screens/ResultsScreen.jsx`, `giftsense-frontend/src/api/giftsense.js`, `giftsense-frontend/src/hooks/`, `giftsense-backend/internal/delivery/`, `giftsense-backend/internal/domain/`, `giftsense-backend/internal/port/`, `giftsense-backend/internal/usecase/`, `giftsense-backend/internal/adapter/`, `giftsense-backend/internal/database/`, `giftsense-backend/config/`, `giftsense-backend/cmd/server/main.go`, `giftsense-backend/api/index.go`
- **Estimated Tasks**: 26
- **Deployment**: Vercel only (serverless backend + static frontend)

---

## Codebase Context

- **Tech Stack**: Go 1.22 + Gin 1.9 + OpenAI Go SDK + Pinecone Go SDK (backend); React 18 + Vite 8 + Tailwind 3.4 + Lucide React (frontend)
- **Architecture Pattern**: Clean Architecture -- domain -> port (interfaces) -> usecase -> adapter -> delivery/http. Frontend: hooks + screens + components
- **Deployment**: Vercel serverless (`api/index.go` with `sync.Once` cold-start pattern). Ephemeral filesystem -- no file-based storage possible.
- **Storage Decision**: **Neon DB free tier** (PostgreSQL, 0.5GB, always-on compute) via **GORM** ORM.
- **ORM Decision**: **GORM** (`gorm.io/gorm` + `gorm.io/driver/postgres`) — Go's most popular ORM. Provides struct-based schema definition, auto-migration, parameterized queries, and connection pooling out of the box. Uses the `postgres` driver which wraps `pgx` internally.
- **Migration Strategy**: Migrations run on application startup via a dedicated `database/` package. GORM's `AutoMigrate` handles table creation and additive schema changes (new columns, new indexes). For destructive changes in the future, numbered SQL migration files can be added alongside.
- **Existing Patterns**:
  - `config.Config` struct loads all env vars in `config/config.go` -- add `DatabaseURL` here
  - `api/index.go` uses `sync.Once` to init Gin router on cold start -- wire DB + migrations here
  - `cmd/server/main.go` for local dev -- wire DB + migrations here too
  - `dto.AnalyzeRequest` / `dto.AnalyzeResponse` / `dto.ErrorResponse` -- pattern for new DTOs
  - `port.VectorStore` interface -- pattern for new `FeedbackStore` port interface
  - `giftsense.js` API client with `analyzeConversation()` -- pattern for new API functions
  - `useAnalyze.js` hook -- pattern for new hooks
- **Key Constraints**:
  - Privacy-first: feedback storage must NOT include conversation content or PII
  - Zero-account model: no user identity beyond session_id
  - 80%+ mobile users, 18-35 age group, India
  - Budget tier already available in frontend form state (no API response change needed)
  - No localStorage/sessionStorage
  - GORM models live in the adapter layer only -- domain types remain ORM-free (Clean Architecture)

---

## Decisions Made

| Question | Decision | Reasoning |
|----------|----------|-----------|
| Storage backend | **Neon DB (PostgreSQL)** free tier | Vercel has ephemeral filesystem -- file-based storage is impossible. Neon free tier gives 0.5GB always-on PostgreSQL at $0. |
| ORM | **GORM** (`gorm.io/gorm`) | Most popular Go ORM. Struct-based models, built-in migration, connection pooling, parameterized queries. `gorm.io/driver/postgres` wraps `pgx` internally. |
| Migration strategy | **GORM AutoMigrate on startup** | Runs in a dedicated `database/` package. Creates tables + indexes if they don't exist. Additive-only (adds columns/indexes, never drops). Safe to run on every cold start — idempotent. Future destructive changes handled by numbered SQL files. |
| GORM models vs domain types | **Separate GORM models in adapter layer** | Domain types (`domain.Feedback`, `domain.AnalyticsEvent`) stay ORM-free. GORM model structs live in `adapter/feedbackstore/` and map to/from domain types. This keeps the Clean Architecture boundary intact — no `gorm` import in domain or usecase. |
| Render support | **Removed** -- Vercel only | User confirmed Vercel-only deployment. Simplifies wiring. |
| Budget tier source | **Frontend form state** | Already available in `App.jsx` state from the input form. No need to modify the analyze API response. |
| Scroll depth tracking | **Not tracked as separate event** | Feedback submission + link clicks provide sufficient signal. Extra network calls on every scroll hurt mobile UX. Fewer events = faster experience on 4G/3G. |
| Event type validation | **`link_click` only** | Removed `scroll_depth` from valid event types since we decided not to track it. |

---

## Design

### UX Flow -- Text-Based Wireframe

The feedback experience unfolds in three progressive stages. Triggered by user behavior (scroll) or a fallback timer -- never immediately on load.

```
==========================================================
RESULTS SCREEN (existing)
==========================================================

  [GiftSense logo]                    [Start over]

  -- Who they are --
  [ InsightCard ]
  [ InsightCard ]

  -- Gift ideas --
  [ GiftCard + ShoppingLinks ]        <-- click tracking
  [ GiftCard + ShoppingLinks ]        <-- click tracking
  [ GiftCard + ShoppingLinks ]

==========================================================
STAGE 1: SOFT PROMPT (appears after user scrolls past
          50% of gift cards OR after 8 seconds, whichever
          comes first)
==========================================================

  +-------------------------------------------------+
  |                                                 |
  |   Were these suggestions helpful?               |
  |                                                 |
  |     [ :) Yes ]      [ :/ Not really ]           |
  |                                                 |
  +-------------------------------------------------+

  Slides up from bottom with a gentle spring animation.
  Two large touch-friendly buttons (min 48px height).
  Emoji + text label, not emoji alone (accessibility).
  Tapping either one transitions to Stage 2.

==========================================================
STAGE 2: FOLLOW-UP (inline, replaces Stage 1)
==========================================================

  IF "Yes" was tapped:
  +-------------------------------------------------+
  |  Glad to hear it! Would you actually buy any    |
  |  of these?                                      |
  |                                                 |
  |     [ Definitely ]  [ Maybe ]  [ Probably not ] |
  |                                                 |
  |  (optional) Anything we could do better?        |
  |  [                                         ]    |
  |  [                                         ]    |
  |                                                 |
  |               [ Send feedback ]                 |
  +-------------------------------------------------+

  IF "Not really" was tapped:
  +-------------------------------------------------+
  |  Sorry about that. What went wrong?             |
  |                                                 |
  |  [ ] Suggestions don't match their personality  |
  |  [ ] Prices are off                             |
  |  [ ] I wanted different categories              |
  |  [ ] Something else                             |
  |                                                 |
  |  (optional) Tell us more:                       |
  |  [                                         ]    |
  |  [                                         ]    |
  |                                                 |
  |               [ Send feedback ]                 |
  +-------------------------------------------------+

  Multi-select checkboxes for structured data.
  Optional textarea (max 500 chars, no placeholder
  that makes it feel required).
  "Send feedback" button is always enabled (text is
  optional).
  Textarea: 2 rows, auto-grows to 4 rows max on mobile.

==========================================================
STAGE 3: THANK YOU (replaces Stage 2, auto-dismiss)
==========================================================

  +-------------------------------------------------+
  |                                                 |
  |   Thank you! Your feedback helps us improve.    |
  |                                                 |
  +-------------------------------------------------+

  Stays visible for 3 seconds, then fades out.
  The widget area collapses smoothly after fade.

==========================================================
DISMISSED STATE (widget never reappears in this session)
==========================================================

  The privacy footer remains at the bottom:
  "Your conversation was anonymised and has not been stored."
```

**Shopping link click tracking** is invisible to the user. Fire-and-forget `POST /api/v1/events` — analytics never degrades UX.

### Components

| Component | New / Existing | Responsibility | File Path |
|-----------|---------------|----------------|-----------|
| `Feedback` (domain type) | New | Domain model for feedback | `giftsense-backend/internal/domain/feedback.go` |
| `AnalyticsEvent` (domain type) | New | Domain model for click events | `giftsense-backend/internal/domain/analytics.go` |
| `FeedbackStore` (port) | New | Interface for persisting feedback | `giftsense-backend/internal/port/feedback.go` |
| `FeedbackService` (usecase) | New | Validate + persist feedback | `giftsense-backend/internal/usecase/feedback.go` |
| `database` (package) | New | GORM connection + migration runner | `giftsense-backend/internal/database/database.go` |
| `migration` (package) | New | Migration definitions (GORM models + AutoMigrate) | `giftsense-backend/internal/database/migration/migration.go` |
| `GormFeedbackStore` (adapter) | New | GORM-based PostgreSQL adapter | `giftsense-backend/internal/adapter/feedbackstore/gorm_store.go` |
| `FeedbackHandler` (delivery) | New | HTTP handler for feedback + events | `giftsense-backend/internal/delivery/http/feedback_handler.go` |
| `FeedbackRequest` / `EventRequest` (dto) | New | Request DTOs | `giftsense-backend/internal/delivery/dto/feedback.go` |
| `FeedbackWidget` | New | 3-stage feedback UI orchestrator | `giftsense-frontend/src/components/results/FeedbackWidget.jsx` |
| `SatisfactionPrompt` | New | Stage 1: yes/no buttons | `giftsense-frontend/src/components/results/feedback/SatisfactionPrompt.jsx` |
| `PositiveFeedbackForm` | New | Stage 2a: purchase intent + text | `giftsense-frontend/src/components/results/feedback/PositiveFeedbackForm.jsx` |
| `NegativeFeedbackForm` | New | Stage 2b: issue checkboxes + text | `giftsense-frontend/src/components/results/feedback/NegativeFeedbackForm.jsx` |
| `ThankYouMessage` | New | Stage 3: confirmation | `giftsense-frontend/src/components/results/feedback/ThankYouMessage.jsx` |
| `useFeedback` | New | Feedback widget state machine hook | `giftsense-frontend/src/hooks/useFeedback.js` |
| `useScrollDepth` | New | Track scroll depth for trigger | `giftsense-frontend/src/hooks/useScrollDepth.js` |
| `giftsense.js` | Existing (modify) | Add `submitFeedback()` and `trackEvent()` | `giftsense-frontend/src/api/giftsense.js` |
| `ShoppingLinks.jsx` | Existing (modify) | Add onClick tracking | `giftsense-frontend/src/components/results/ShoppingLinks.jsx` |
| `ResultsScreen.jsx` | Existing (modify) | Mount FeedbackWidget | `giftsense-frontend/src/screens/ResultsScreen.jsx` |
| `App.jsx` | Existing (modify) | Pass sessionId + budgetTier to ResultsScreen | `giftsense-frontend/src/App.jsx` |
| `config.go` | Existing (modify) | Add `DatabaseURL` env var | `giftsense-backend/config/config.go` |
| `main.go` | Existing (modify) | Wire DB + migrations + feedback handler | `giftsense-backend/cmd/server/main.go` |
| `api/index.go` | Existing (modify) | Wire DB + migrations + feedback handler in Vercel entry | `giftsense-backend/api/index.go` |

### Database Architecture

```
giftsense-backend/internal/database/
├── database.go              -- Connect() function: opens GORM connection, returns *gorm.DB
└── migration/
    └── migration.go         -- GORM models + RunMigrations(*gorm.DB) function

giftsense-backend/internal/adapter/feedbackstore/
└── gorm_store.go            -- GormFeedbackStore: implements port.FeedbackStore using *gorm.DB
```

**Separation of concerns:**
- `database/database.go` — pure connection logic. Accepts `DATABASE_URL`, returns `*gorm.DB`. Configures connection pool (max 5 open, max 3 idle for serverless).
- `database/migration/migration.go` — defines GORM model structs (with `gorm:` struct tags) and a `RunMigrations(*gorm.DB)` function that calls `db.AutoMigrate()`. These models are **not** the domain types — they are ORM-specific and live in the database layer.
- `adapter/feedbackstore/gorm_store.go` — implements `port.FeedbackStore`. Receives `*gorm.DB`, maps domain types to/from GORM models, executes queries.

**Why this split matters:** The GORM model structs have `gorm:` tags for column names, indexes, and constraints. Domain types have `json:` tags. Mixing them violates Clean Architecture. The adapter layer handles the mapping.

### Interactions

1. **On application startup** (cold start on Vercel, or `main.go` locally):
   - `database.Connect(databaseURL)` opens a GORM PostgreSQL connection
   - `migration.RunMigrations(db)` runs `AutoMigrate` — creates/updates tables idempotently
   - `feedbackstore.NewGormFeedbackStore(db)` creates the adapter
   - `FeedbackService` and `FeedbackHandler` are constructed and routes registered
2. User lands on ResultsScreen after successful analysis. Personality insights and gift cards render.
3. User scrolls through results. `useScrollDepth` hook detects when 50% of gift cards are in viewport (or 8-second timer fires, whichever is first).
4. `FeedbackWidget` transitions from `hidden` -> `prompt` stage. `SatisfactionPrompt` slides up.
5. User taps "Yes" or "Not really". `FeedbackWidget` transitions to `followup` stage.
6. User optionally selects checkboxes and/or types free text, then taps "Send feedback".
7. `useFeedback` hook calls `submitFeedback()` from `giftsense.js` -> `POST /api/v1/feedback`.
8. Backend `FeedbackHandler` validates -> maps to domain -> calls `FeedbackService.SubmitFeedback()`.
9. `FeedbackService` validates -> calls `FeedbackStore.SaveFeedback()` -> GORM inserts row into `feedbacks` table.
10. Frontend transitions to `thankyou` **optimistically** (before server confirms). After 3s, fades to `dismissed`.
11. **Shopping link clicks**: `ShoppingLinks.jsx` fires `trackEvent()` (fire-and-forget) -> `POST /api/v1/events` -> GORM inserts into `analytics_events` table.
12. **On failure**: All feedback/event submissions are fire-and-forget on the frontend. Analytics never degrades UX.

### Contracts

**GORM Models (in `database/migration/migration.go`):**

```go
package migration

import (
    "time"
    "gorm.io/gorm"
    "github.com/lib/pq"
)

type FeedbackModel struct {
    ID              uint           `gorm:"primaryKey"`
    SessionID       string         `gorm:"type:uuid;not null;index:idx_feedback_session"`
    Satisfaction    string         `gorm:"type:varchar(20);not null"`
    PurchaseIntent  string         `gorm:"type:varchar(20)"`
    Issues          pq.StringArray `gorm:"type:text[]"`
    FreeText        string         `gorm:"type:varchar(500)"`
    BudgetTier      string         `gorm:"type:varchar(20);not null"`
    SuggestionCount int            `gorm:"not null;default:0"`
    CreatedAt       time.Time      `gorm:"autoCreateTime;index:idx_feedback_created_at"`
}

func (FeedbackModel) TableName() string { return "feedbacks" }

type AnalyticsEventModel struct {
    ID        uint      `gorm:"primaryKey"`
    SessionID string    `gorm:"type:uuid;not null;index:idx_events_session"`
    EventType string    `gorm:"type:varchar(20);not null"`
    Target    string    `gorm:"type:varchar(100);not null"`
    Metadata  string    `gorm:"type:jsonb"`
    CreatedAt time.Time `gorm:"autoCreateTime;index:idx_events_created_at"`
}

func (AnalyticsEventModel) TableName() string { return "analytics_events" }

func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(&FeedbackModel{}, &AnalyticsEventModel{})
}
```

**Backend domain types (ORM-free):**

```go
// domain/feedback.go
package domain

import "time"

type SatisfactionRating string

const (
    SatisfactionHelpful    SatisfactionRating = "helpful"
    SatisfactionNotHelpful SatisfactionRating = "not_helpful"
)

type PurchaseIntent string

const (
    PurchaseIntentDefinitely   PurchaseIntent = "definitely"
    PurchaseIntentMaybe        PurchaseIntent = "maybe"
    PurchaseIntentProbablyNot  PurchaseIntent = "probably_not"
)

type Feedback struct {
    SessionID       string             `json:"session_id"`
    Satisfaction    SatisfactionRating `json:"satisfaction"`
    PurchaseIntent  PurchaseIntent     `json:"purchase_intent,omitempty"`
    Issues          []string           `json:"issues,omitempty"`
    FreeText        string             `json:"free_text,omitempty"`
    BudgetTier      string             `json:"budget_tier"`
    SuggestionCount int                `json:"suggestion_count"`
    Timestamp       time.Time          `json:"timestamp"`
}
```

```go
// domain/analytics.go
package domain

import "time"

type AnalyticsEvent struct {
    SessionID  string            `json:"session_id"`
    EventType  string            `json:"event_type"`
    Target     string            `json:"target"`
    Metadata   map[string]string `json:"metadata,omitempty"`
    Timestamp  time.Time         `json:"timestamp"`
}
```

**Backend port interface:**

```go
// port/feedback.go
type FeedbackStore interface {
    SaveFeedback(ctx context.Context, feedback domain.Feedback) error
    SaveEvent(ctx context.Context, event domain.AnalyticsEvent) error
}
```

**Backend DTOs:**

```go
// dto/feedback.go
type FeedbackRequest struct {
    SessionID       string   `json:"session_id" binding:"required,uuid"`
    Satisfaction    string   `json:"satisfaction" binding:"required,oneof=helpful not_helpful"`
    PurchaseIntent  string   `json:"purchase_intent,omitempty" binding:"omitempty,oneof=definitely maybe probably_not"`
    Issues          []string `json:"issues,omitempty"`
    FreeText        string   `json:"free_text,omitempty" binding:"max=500"`
    BudgetTier      string   `json:"budget_tier" binding:"required"`
    SuggestionCount int      `json:"suggestion_count" binding:"min=0"`
}

type EventRequest struct {
    SessionID string            `json:"session_id" binding:"required,uuid"`
    EventType string            `json:"event_type" binding:"required,oneof=link_click"`
    Target    string            `json:"target" binding:"required"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

type FeedbackResponse struct {
    Message string `json:"message"`
}
```

**API endpoints:**

```
POST /api/v1/feedback
Content-Type: application/json
{
    "session_id": "uuid",
    "satisfaction": "helpful" | "not_helpful",
    "purchase_intent": "definitely" | "maybe" | "probably_not",
    "issues": ["personality_mismatch", "price_mismatch", "wrong_categories", "other"],
    "free_text": "optional text, max 500 chars",
    "budget_tier": "BUDGET" | "MID_RANGE" | "PREMIUM" | "LUXURY",
    "suggestion_count": 5
}
Response 201: { "message": "Feedback received" }

POST /api/v1/events
Content-Type: application/json
{
    "session_id": "uuid",
    "event_type": "link_click",
    "target": "amazon",
    "metadata": { "gift_name": "Pottery Kit", "gift_index": "0" }
}
Response 204: (no body)
```

**Frontend API functions:**

```javascript
// Added to giftsense.js
export async function submitFeedback(payload) { ... }
export async function trackEvent(payload) { ... }  // fire-and-forget, never throws
```

**Frontend hook state machine:**

```javascript
// useFeedback.js
// States: hidden -> prompt -> followup -> submitting -> thankyou -> dismissed
```

---

## Implementation Tasks

> Tasks are sequenced by dependency. Execute in order.

### Phase A: Backend Domain + Port (Foundation)

- [ ] **Task 1** -- `Create Feedback domain type with satisfaction rating and purchase intent enums`
  - **File**: `giftsense-backend/internal/domain/feedback.go`
  - **Action**: Create new file. Define `SatisfactionRating` string type with constants `SatisfactionHelpful` ("helpful") and `SatisfactionNotHelpful` ("not_helpful"). Define `PurchaseIntent` string type with constants `PurchaseIntentDefinitely` ("definitely"), `PurchaseIntentMaybe` ("maybe"), `PurchaseIntentProbablyNot` ("probably_not"). Define `Feedback` struct with fields: `SessionID string`, `Satisfaction SatisfactionRating`, `PurchaseIntent PurchaseIntent`, `Issues []string`, `FreeText string`, `BudgetTier string`, `SuggestionCount int`, `Timestamp time.Time`. All fields have `json` struct tags only — no `gorm` tags. No third-party imports.
  - **Done When**: File compiles. `go vet ./internal/domain/...` passes. No external dependencies.

- [ ] **Task 2** -- `Create AnalyticsEvent domain type for click events`
  - **File**: `giftsense-backend/internal/domain/analytics.go`
  - **Action**: Create new file. Define `AnalyticsEvent` struct with fields: `SessionID string`, `EventType string`, `Target string`, `Metadata map[string]string`, `Timestamp time.Time`. All fields have `json` struct tags only. No third-party imports.
  - **Done When**: File compiles. No external dependencies.

- [ ] **Task 3** -- `Define FeedbackStore port interface`
  - **File**: `giftsense-backend/internal/port/feedback.go`
  - **Action**: Create new file. Define `FeedbackStore` interface with two methods: `SaveFeedback(ctx context.Context, feedback domain.Feedback) error` and `SaveEvent(ctx context.Context, event domain.AnalyticsEvent) error`. Import only `context` and `domain` package.
  - **Done When**: File compiles. Interface has exactly two methods. No third-party imports.

### Phase B: Backend Use Case + Tests

- [ ] **Task 4** -- `Create FeedbackService use case with validation logic`
  - **File**: `giftsense-backend/internal/usecase/feedback.go`
  - **Action**: Create new file. Define `FeedbackService` struct holding a `port.FeedbackStore` dependency. Constructor: `NewFeedbackService(store port.FeedbackStore) *FeedbackService`. Method `SubmitFeedback(ctx context.Context, fb domain.Feedback) error`: validate `Satisfaction` is one of the two valid values, validate `FreeText` <= 500 chars, set `Timestamp` to `time.Now()`, call `store.SaveFeedback()`. Method `TrackEvent(ctx context.Context, evt domain.AnalyticsEvent) error`: validate `EventType` is `link_click`, set `Timestamp` to `time.Now()`, call `store.SaveEvent()`. No third-party imports.
  - **Done When**: File compiles. Both methods validate and delegate.

- [ ] **Task 5** -- `Write tests for FeedbackService use case`
  - **File**: `giftsense-backend/internal/usecase/feedback_test.go`
  - **Action**: Create new file. Define `mockFeedbackStore` struct in test file. Table-driven tests:
    - `TestSubmitFeedback_ShouldSaveFeedback_WhenSatisfactionIsHelpful`
    - `TestSubmitFeedback_ShouldSaveFeedback_WhenSatisfactionIsNotHelpful`
    - `TestSubmitFeedback_ShouldReturnError_WhenSatisfactionIsInvalid`
    - `TestSubmitFeedback_ShouldReturnError_WhenFreeTextExceedsMaxLength`
    - `TestSubmitFeedback_ShouldSetTimestamp_WhenCalled`
    - `TestTrackEvent_ShouldSaveEvent_WhenEventTypeIsLinkClick`
    - `TestTrackEvent_ShouldReturnError_WhenEventTypeIsInvalid`
  - **Done When**: All tests pass.

### Phase C: Backend Database + GORM Migration

- [ ] **Task 6** -- `Add GORM and postgres driver dependencies to go.mod`
  - **File**: `giftsense-backend/go.mod`
  - **Action**: Run `go get gorm.io/gorm gorm.io/driver/postgres github.com/lib/pq` to add GORM, the Postgres driver (wraps pgx internally), and `lib/pq` for `pq.StringArray` type (PostgreSQL text[] support in GORM).
  - **Done When**: `go.mod` contains all three packages. `go build ./...` succeeds.

- [ ] **Task 7** -- `Create database connection package`
  - **File**: `giftsense-backend/internal/database/database.go`
  - **Action**: Create new file. Define `Connect(databaseURL string) (*gorm.DB, error)` function. Opens a GORM connection with `gorm.io/driver/postgres.Open(databaseURL)`. Configure: `&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}` (suppress SQL logs in production — never log session data). Get the underlying `*sql.DB` via `db.DB()` and set: `SetMaxOpenConns(5)`, `SetMaxIdleConns(3)`, `SetConnMaxLifetime(5 * time.Minute)` — tuned for serverless where connections are short-lived. Return `*gorm.DB`.
  - **Done When**: File compiles. Returns a configured `*gorm.DB`. Connection pool settings appropriate for serverless.

- [ ] **Task 8** -- `Create migration package with GORM models and RunMigrations`
  - **File**: `giftsense-backend/internal/database/migration/migration.go`
  - **Action**: Create new file. Define `FeedbackModel` struct with GORM tags: `ID uint` (primaryKey), `SessionID string` (type:uuid, not null, indexed), `Satisfaction string` (varchar(20), not null), `PurchaseIntent string` (varchar(20)), `Issues pq.StringArray` (type:text[]), `FreeText string` (varchar(500)), `BudgetTier string` (varchar(20), not null), `SuggestionCount int` (not null, default:0), `CreatedAt time.Time` (autoCreateTime, indexed). Custom `TableName() string` returns `"feedbacks"`. Define `AnalyticsEventModel` struct: `ID uint` (primaryKey), `SessionID string` (type:uuid, not null, indexed), `EventType string` (varchar(20), not null), `Target string` (varchar(100), not null), `Metadata string` (type:jsonb), `CreatedAt time.Time` (autoCreateTime, indexed). Custom `TableName()` returns `"analytics_events"`. Define `RunMigrations(db *gorm.DB) error` that calls `db.AutoMigrate(&FeedbackModel{}, &AnalyticsEventModel{})`. This is idempotent — safe to run on every cold start.
  - **Done When**: File compiles. `RunMigrations` creates both tables with correct columns, types, and indexes. Models have custom table names.

- [ ] **Task 9** -- `Write tests for database connection and migrations`
  - **File**: `giftsense-backend/internal/database/database_test.go`
  - **Action**: Create new file. Tests check for `TEST_DATABASE_URL` env var — if not set, skip with `t.Skip("TEST_DATABASE_URL not set")`. Tests:
    - `TestConnect_ShouldReturnGormDB_WhenValidURLProvided`
    - `TestConnect_ShouldReturnError_WhenInvalidURLProvided`
    - `TestRunMigrations_ShouldCreateTables_WhenCalledOnEmptyDatabase`
    - `TestRunMigrations_ShouldBeIdempotent_WhenCalledMultipleTimes`
  - **Done When**: Tests pass when `TEST_DATABASE_URL` is set. Skip cleanly when not.

### Phase D: Backend Adapter (GORM FeedbackStore)

- [ ] **Task 10** -- `Create GormFeedbackStore adapter`
  - **File**: `giftsense-backend/internal/adapter/feedbackstore/gorm_store.go`
  - **Action**: Create new file. Define `GormFeedbackStore` struct holding `*gorm.DB`. Constructor: `NewGormFeedbackStore(db *gorm.DB) *GormFeedbackStore`. `SaveFeedback(ctx, fb domain.Feedback)`: map `domain.Feedback` to `migration.FeedbackModel` (convert `Issues []string` to `pq.StringArray`), call `db.WithContext(ctx).Create(&model)`. `SaveEvent(ctx, evt domain.AnalyticsEvent)`: map `domain.AnalyticsEvent` to `migration.AnalyticsEventModel` (marshal `Metadata map[string]string` to JSON string for the JSONB column), call `db.WithContext(ctx).Create(&model)`. Never log session_id.
  - **Done When**: File compiles. Implements `port.FeedbackStore`. Uses `db.WithContext(ctx)` for all queries. Maps domain <-> GORM models cleanly.

- [ ] **Task 11** -- `Write tests for GormFeedbackStore adapter`
  - **File**: `giftsense-backend/internal/adapter/feedbackstore/gorm_store_test.go`
  - **Action**: Create new file. Skip if `TEST_DATABASE_URL` not set. Run migrations in test setup. Tests:
    - `TestGormFeedbackStore_ShouldInsertFeedback_WhenSaveFeedbackCalled`
    - `TestGormFeedbackStore_ShouldInsertEvent_WhenSaveEventCalled`
    - `TestGormFeedbackStore_ShouldStoreIssuesAsArray_WhenMultipleIssuesProvided`
    - `TestGormFeedbackStore_ShouldStoreMetadataAsJSON_WhenMetadataProvided`
    - `TestGormFeedbackStore_ShouldHandleConcurrentInserts_WhenCalledFromMultipleGoroutines`
  - **Done When**: Tests pass with `TEST_DATABASE_URL`. Skip cleanly without.

### Phase E: Backend HTTP Layer (DTOs + Handlers)

- [ ] **Task 12** -- `Create feedback and event request/response DTOs`
  - **File**: `giftsense-backend/internal/delivery/dto/feedback.go`
  - **Action**: Create new file. Define `FeedbackRequest` with Gin binding tags (session_id required+uuid, satisfaction required+oneof, purchase_intent omitempty+oneof, free_text max=500, budget_tier required, suggestion_count min=0). Define `EventRequest` (session_id required+uuid, event_type required+oneof=link_click, target required, metadata optional). Define `FeedbackResponse` with Message string.
  - **Done When**: File compiles. Struct tags match the API contract.

- [ ] **Task 13** -- `Create FeedbackHandler with POST endpoints`
  - **File**: `giftsense-backend/internal/delivery/http/feedback_handler.go`
  - **Action**: Create new file. `FeedbackHandler` struct holds `*usecase.FeedbackService`. `SubmitFeedback(c *gin.Context)`: bind JSON -> map to domain -> call service -> return 201. `TrackEvent(c *gin.Context)`: bind JSON -> map to domain -> call service -> return 204. On validation error: 400. On service error in TrackEvent: log and return 204 anyway (analytics never fails visibly).
  - **Done When**: File compiles. Follows existing `AnalyzeHandler` error-handling pattern.

- [ ] **Task 14** -- `Write tests for FeedbackHandler`
  - **File**: `giftsense-backend/internal/delivery/http/feedback_handler_test.go`
  - **Action**: Use `httptest` + `gin.CreateTestContext` + in-memory mock store. Tests:
    - `TestSubmitFeedback_ShouldReturn201_WhenValidFeedbackProvided`
    - `TestSubmitFeedback_ShouldReturn400_WhenSatisfactionMissing`
    - `TestSubmitFeedback_ShouldReturn400_WhenSessionIDInvalid`
    - `TestSubmitFeedback_ShouldReturn400_WhenFreeTextTooLong`
    - `TestTrackEvent_ShouldReturn204_WhenValidEventProvided`
    - `TestTrackEvent_ShouldReturn400_WhenEventTypeInvalid`
  - **Done When**: All tests pass.

### Phase F: Backend Wiring (Config + Entry Points)

- [ ] **Task 15** -- `Add DATABASE_URL env var to config`
  - **File**: `giftsense-backend/config/config.go`
  - **Action**: Add `DatabaseURL string` field to `Config` struct. Load from `DATABASE_URL` env var. This is **optional** — if not set, feedback features are disabled (the app still works for analyze-only). Add a helper method `Config.HasDatabase() bool` that returns `DatabaseURL != ""`.
  - **Done When**: `config.Load()` populates `DatabaseURL`. Existing tests pass. Missing `DATABASE_URL` does not crash the app.

- [ ] **Task 16** -- `Wire DB, migrations, and feedback handler in cmd/server/main.go`
  - **File**: `giftsense-backend/cmd/server/main.go`
  - **Action**: After existing handler construction, check `cfg.HasDatabase()`. If true: call `database.Connect(cfg.DatabaseURL)` to get `*gorm.DB`, call `migration.RunMigrations(db)` to apply migrations, construct `GormFeedbackStore` with `db`, construct `FeedbackService`, construct `FeedbackHandler`, register `v1.POST("/feedback", ...)` and `v1.POST("/events", ...)`. If false: log a warning that feedback endpoints are disabled. Add imports for `database`, `migration`, and `feedbackstore` packages.
  - **Done When**: `go build ./...` succeeds. With `DATABASE_URL` set: migrations run on startup, feedback endpoints return 400 (missing body). Without: endpoints return 404.

- [ ] **Task 17** -- `Wire DB, migrations, and feedback handler in api/index.go (Vercel entry)`
  - **File**: `giftsense-backend/api/index.go`
  - **Action**: Same wiring as `main.go` but inside the `sync.Once` block. Check `cfg.HasDatabase()`, call `database.Connect()`, call `migration.RunMigrations()`, construct the store/service/handler, register routes. Important: the `*gorm.DB` must be created once and reused across invocations (held in a package-level var alongside the existing Gin engine). Migrations only run once per cold start (inside `sync.Once`), so there is no per-request overhead.
  - **Done When**: `go build ./...` succeeds. Vercel handler includes feedback routes when `DATABASE_URL` is set. Migrations run exactly once per cold start.

- [ ] **Task 18** -- `Update .env.example`
  - **File**: `giftsense-backend/.env.example`
  - **Action**: Add `DATABASE_URL=` with comment: `# Neon PostgreSQL connection string (optional — feedback features disabled if not set)`. Format: `postgresql://user:pass@host/dbname?sslmode=require`.
  - **Done When**: `.env.example` documents the new variable.

### Phase G: Frontend API Client + Hooks

- [ ] **Task 19** -- `Add submitFeedback() and trackEvent() to API client`
  - **File**: `giftsense-frontend/src/api/giftsense.js`
  - **Action**: Add two exported async functions. `submitFeedback(payload)`: POST JSON to `${API_URL}/api/v1/feedback`, return response JSON, throw on non-2xx. `trackEvent(payload)`: POST JSON to `${API_URL}/api/v1/events`, fire-and-forget — wrap in try/catch that silently swallows all errors. Use `navigator.sendBeacon()` as enhancement for page-unload scenarios.
  - **Done When**: Both exported. `trackEvent` never throws. `submitFeedback` throws on HTTP errors.

- [ ] **Task 20** -- `Create useScrollDepth hook`
  - **File**: `giftsense-frontend/src/hooks/useScrollDepth.js`
  - **Action**: Export `useScrollDepth(threshold = 0.5)`. Attach passive scroll listener. Compute scroll ratio: `(window.scrollY + window.innerHeight) / document.documentElement.scrollHeight`. When ratio > threshold, set `hasScrolledPast` to true (once, never resets). Throttle with `requestAnimationFrame`. Clean up on unmount. Return `{ hasScrolledPast }`.
  - **Done When**: Returns `hasScrolledPast: true` when scrolled past 50%. Listener cleaned up.

- [ ] **Task 21** -- `Create useFeedback state machine hook`
  - **File**: `giftsense-frontend/src/hooks/useFeedback.js`
  - **Action**: Export `useFeedback(sessionId)`. States: `hidden` -> `prompt` -> `followup` -> `submitting` -> `thankyou` -> `dismissed`. Expose: `stage`, `satisfaction`, `showPrompt()`, `submitSatisfaction(rating)`, `submitFeedback({ purchaseIntent, issues, freeText, budgetTier, suggestionCount })`. `submitFeedback` transitions to `thankyou` **optimistically** (immediately), then fires the API call. After 3 seconds in `thankyou`, auto-transition to `dismissed`. All errors silently caught.
  - **Done When**: Full lifecycle works. Errors during submission are swallowed.

### Phase H: Frontend Components

- [ ] **Task 22** -- `Create SatisfactionPrompt component (Stage 1)`
  - **File**: `giftsense-frontend/src/components/results/feedback/SatisfactionPrompt.jsx`
  - **Action**: Create `feedback/` subdirectory. Export `SatisfactionPrompt({ onSelect })`. Card with "Were these suggestions helpful?" + two buttons (smile emoji + "Yes", neutral face + "Not really"). Buttons call `onSelect("helpful")` / `onSelect("not_helpful")`. Styling: rounded-xl, bg-white, shadow-sm, p-4. Buttons: min-h-[48px], `grid grid-cols-2 gap-3`, border hover states. Slide-up entry animation (`translate-y-4 opacity-0` -> `translate-y-0 opacity-100`). `role="group"` + `aria-label="Rate suggestions"`.
  - **Done When**: Two labeled buttons, >= 48px touch targets, slide-up animation.

- [ ] **Task 23** -- `Create PositiveFeedbackForm component (Stage 2a)`
  - **File**: `giftsense-frontend/src/components/results/feedback/PositiveFeedbackForm.jsx`
  - **Action**: Export `PositiveFeedbackForm({ onSubmit })`. Heading + 3 purchase intent pills (definitely/maybe/probably not) + optional textarea (max 500 chars, 2 rows, auto-grow to 4) + "Send feedback" button. Selected pill gets purple border+fill. Submit calls `onSubmit({ purchaseIntent, freeText })`.
  - **Done When**: Collects purchase intent + optional text. Submit fires with correct shape.

- [ ] **Task 24** -- `Create NegativeFeedbackForm component (Stage 2b)`
  - **File**: `giftsense-frontend/src/components/results/feedback/NegativeFeedbackForm.jsx`
  - **Action**: Export `NegativeFeedbackForm({ onSubmit })`. Heading + 4 checkboxes (personality_mismatch, price_mismatch, wrong_categories, other) with human-readable labels + optional textarea + "Send feedback" button (always enabled). Custom Tailwind checkboxes with Lucide Check icon, min 44px touch target. Submit calls `onSubmit({ issues, freeText })`.
  - **Done When**: Multi-select issues + optional text. Submit fires with correct shape.

- [ ] **Task 25** -- `Create ThankYouMessage component (Stage 3) and FeedbackWidget orchestrator`
  - **Files**: `giftsense-frontend/src/components/results/feedback/ThankYouMessage.jsx`, `giftsense-frontend/src/components/results/FeedbackWidget.jsx`
  - **Action**:
    - **ThankYouMessage.jsx**: Export `ThankYouMessage()`. Centered text "Thank you! Your feedback helps us improve." with Lucide `CheckCircle2` icon. Fade-in animation. Same card style.
    - **FeedbackWidget.jsx**: Export `FeedbackWidget({ sessionId, budgetTier, suggestionCount })`. Uses `useFeedback(sessionId)` + `useScrollDepth(0.5)`. When `hasScrolledPast` OR 8-second timer fires, call `showPrompt()`. Render by stage: hidden->null, prompt->SatisfactionPrompt, followup->Positive/NegativeFeedbackForm, submitting->form disabled, thankyou->ThankYouMessage, dismissed->null. `aria-live="polite"` wrapper. Pass `budgetTier` and `suggestionCount` through to submission.
  - **Done When**: FeedbackWidget orchestrates all 3 stages. Trigger on scroll or timeout.

### Phase I: Frontend Integration (Wire Everything)

- [ ] **Task 26** -- `Mount FeedbackWidget in ResultsScreen and add click tracking to ShoppingLinks`
  - **Files**: `giftsense-frontend/src/App.jsx`, `giftsense-frontend/src/screens/ResultsScreen.jsx`, `giftsense-frontend/src/components/results/ShoppingLinks.jsx`
  - **Action**:
    - **App.jsx**: Pass `sessionId` and `budgetTier` (from form state) as props to `ResultsScreen`.
    - **ResultsScreen.jsx**: Accept `sessionId` and `budgetTier` props. Import `FeedbackWidget`. Render after gift suggestions, before privacy footer: `<FeedbackWidget sessionId={sessionId} budgetTier={budgetTier} suggestionCount={data.gift_suggestions.length} />`. Pass `sessionId` down to each `GiftCard` -> `ShoppingLinks`.
    - **ShoppingLinks.jsx**: Import `trackEvent` from `giftsense.js`. On link click, call `trackEvent({ session_id: sessionId, event_type: 'link_click', target: storeName, metadata: { gift_name: giftName } })`. Do NOT `preventDefault` — let the link navigate normally.
  - **Done When**: FeedbackWidget appears after scroll/8s. Link clicks tracked. `npm run build` succeeds.

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Neon free tier connection limits (max 5 concurrent) | Medium | Medium | GORM pool configured with max 5 open conns. Serverless functions are short-lived. |
| GORM AutoMigrate adds latency to cold start | Low | Low | AutoMigrate is idempotent and fast when tables already exist (~50ms). Runs once per cold start inside `sync.Once`. |
| GORM AutoMigrate doesn't handle destructive changes (column drops, type changes) | Medium | Low | Acceptable for soft launch. Future breaking changes handled via numbered SQL migration files run in `RunMigrations()` before `AutoMigrate`. |
| Feedback table grows over time | Low | Low | 0.5GB free tier. Each row ~200 bytes = ~2.5M rows. More than enough for soft launch. |
| Users never scroll far enough | Medium | Medium | 8-second fallback timer ensures prompt appears regardless. |
| CORS blocks new JSON endpoints | Low | Medium | New endpoints under same `/api/v1` prefix, same CORS middleware. `Content-Type: application/json` already allowed. |
| `DATABASE_URL` not set on Vercel | Low | High | Graceful degradation — feedback endpoints not registered. Analyze still works. Clear log warning. |
| `pq.StringArray` compatibility with GORM | Low | Low | Well-tested combination. `lib/pq` is the standard Go Postgres array type. |

---

## Environment Setup Required

Before deploying:
1. **Create Neon DB** at [neon.tech](https://neon.tech) — free tier, region: Singapore (closest to India)
2. **Copy the connection string** (format: `postgresql://user:pass@host/dbname?sslmode=require`)
3. **Add to Vercel environment variables**: `DATABASE_URL=<connection string>`
4. **Tables are auto-created on first cold start** — GORM `AutoMigrate` runs inside `sync.Once`, creating `feedbacks` and `analytics_events` tables with all columns and indexes. No manual SQL required.
