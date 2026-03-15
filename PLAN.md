# GiftSense Implementation Plan
**Status:** IN_PROGRESS

---

## Project Directory Structure

```
gift-sense/                          ← repo root (this directory)
├── CLAUDE.md
├── PLAN.md
├── PROGRESS.md                      ← created by orchestrator
├── docs/
│   ├── architecture.md
│   └── coding-rules.md
│
├── giftsense-backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              ← entry point; wires all deps, starts Gin
│   ├── config/
│   │   └── config.go                ← ONLY file that calls os.Getenv()
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── conversation.go      ← Message, Chunk, Session types
│   │   │   ├── recipient.go         ← RecipientDetails, BudgetRange, BudgetTier
│   │   │   ├── suggestion.go        ← GiftSuggestion, PersonalityInsight
│   │   │   └── errors.go            ← sentinel errors (ErrFileTooLarge, etc.)
│   │   ├── port/
│   │   │   ├── embedder.go          ← Embedder interface
│   │   │   ├── llm.go               ← LLMClient interface + CompletionOptions
│   │   │   └── vectorstore.go       ← VectorStore interface + MetadataFilter
│   │   ├── usecase/
│   │   │   ├── analyze.go           ← AnalyzeConversation orchestrator (13-step pipeline)
│   │   │   ├── parse.go             ← WhatsApp + plain-text parser, smart sampling
│   │   │   ├── anonymize.go         ← NER (sender-seeded + regex) + pseudonymization
│   │   │   ├── chunk.go             ← Sliding window chunker + metadata heuristics
│   │   │   └── retrieve.go          ← Multi-query construction + re-ranking
│   │   ├── adapter/
│   │   │   ├── openai/
│   │   │   │   ├── embedder.go      ← Embedder impl via OpenAI text-embedding-3-small
│   │   │   │   └── llm.go           ← LLMClient impl via OpenAI GPT-4o (JSON mode)
│   │   │   ├── vectorstore/
│   │   │   │   ├── pinecone.go      ← VectorStore impl: Upsert/Query/DeleteSession
│   │   │   │   └── memory.go        ← In-memory VectorStore for tests (no-op delete)
│   │   │   └── linkgen/
│   │   │       └── shopping.go      ← Amazon/Flipkart/GoogleShopping URL builder
│   │   └── delivery/
│   │       ├── http/
│   │       │   ├── handler.go       ← POST /api/v1/analyze + GET /health + GET /metrics
│   │       │   ├── middleware.go    ← CORS, structured logging, rate limiting
│   │       │   └── validator.go     ← file size, UUID format, field validation
│   │       └── dto/
│   │           ├── request.go       ← AnalyzeRequest (multipart form fields)
│   │           └── response.go      ← AnalysisResponse, ErrorResponse
│   ├── .env.example
│   ├── go.mod
│   └── go.sum
│
└── giftsense-frontend/
    ├── src/
    │   ├── main.jsx
    │   ├── App.jsx                  ← three-screen router + global state
    │   ├── hooks/
    │   │   ├── useSession.js        ← crypto.randomUUID() in useRef (one per window)
    │   │   └── useAnalyze.js        ← API call state machine (idle/loading/done/error)
    │   ├── components/
    │   │   ├── upload/
    │   │   │   ├── UploadZone.jsx   ← drag-drop (desktop) + tap-to-pick (mobile)
    │   │   │   └── TextPaste.jsx    ← collapsible textarea for paste input
    │   │   ├── form/
    │   │   │   ├── RecipientForm.jsx ← name, relation, gender, occasion fields
    │   │   │   └── BudgetSelector.jsx ← 4-tier card selector (responsive grid)
    │   │   ├── results/
    │   │   │   ├── InsightCard.jsx   ← single personality insight (insight + evidence)
    │   │   │   ├── GiftCard.jsx      ← gift suggestion card (name, reason, price badge)
    │   │   │   └── ShoppingLinks.jsx ← Amazon / Flipkart / Google Shopping link buttons
    │   │   └── shared/
    │   │       ├── LoadingScreen.jsx ← rotating text animation
    │   │       ├── ErrorMessage.jsx  ← retry-able error display
    │   │       └── PrivacyNotice.jsx ← pre-upload + post-results banners
    │   ├── screens/
    │   │   ├── InputScreen.jsx      ← upload + form + submit
    │   │   ├── LoadingScreen.jsx    ← progress animation wrapper
    │   │   └── ResultsScreen.jsx    ← insights + gifts + back button
    │   └── api/
    │       └── giftsense.js         ← fetch wrapper; attaches session_id to every call
    ├── index.html
    ├── vite.config.js
    ├── tailwind.config.js
    ├── postcss.config.js
    ├── .env.example                 ← VITE_API_URL=
    └── package.json
```

---

## Environment Variables Registry

| Variable | Type | Default | Component | Required |
|---|---|---|---|---|
| `OPENAI_API_KEY` | string | — | `config.go` | **YES** |
| `PINECONE_API_KEY` | string | — | `config.go` | **YES** |
| `PINECONE_ENVIRONMENT` | string | — | `config.go` | **YES** |
| `CHAT_MODEL` | string | `gpt-4o` | `adapter/openai/llm.go` | no |
| `EMBEDDING_MODEL` | string | `text-embedding-3-small` | `adapter/openai/embedder.go` | no |
| `EMBEDDING_DIMENSIONS` | int | `1536` | `adapter/openai/embedder.go`, `adapter/vectorstore/pinecone.go` | no |
| `MAX_TOKENS` | int | `1000` | `adapter/openai/llm.go` | no |
| `TOP_K` | int | `3` | `usecase/retrieve.go` | no |
| `NUM_RETRIEVAL_QUERIES` | int | `4` | `usecase/retrieve.go` | no |
| `PINECONE_INDEX_NAME` | string | `giftsense` | `adapter/vectorstore/pinecone.go` | no |
| `MAX_FILE_SIZE_BYTES` | int64 | `2097152` | `delivery/http/handler.go`, `delivery/http/middleware.go` | no |
| `MAX_PROCESSED_MESSAGES` | int | `400` | `usecase/parse.go` | no |
| `CHUNK_WINDOW_SIZE` | int | `8` | `usecase/chunk.go` | no |
| `CHUNK_OVERLAP_SIZE` | int | `3` | `usecase/chunk.go` | no |
| `PORT` | string | `8080` | `cmd/server/main.go` | set by Render |
| `ALLOWED_ORIGINS` | []string | `http://localhost:5173` | `delivery/http/middleware.go` | no |

---

## Go Interface Contracts

These interfaces MUST exist in `internal/port/` before any implementation or use-case work begins. They are the binding contract between all teams.

```go
// port/embedder.go
package port

import "context"

type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// port/llm.go
package port

import "context"

type CompletionOptions struct {
    MaxTokens   int
    JSONMode    bool
}

type LLMClient interface {
    Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, error)
}

// port/vectorstore.go
package port

import (
    "context"
    "github.com/gaurav/giftsense/internal/domain"
)

type MetadataFilter struct {
    HasPreference *bool
    HasWish       *bool
    Topics        []string
}

type VectorStore interface {
    Upsert(ctx context.Context, sessionID string, chunks []domain.Chunk, vectors [][]float32) error
    Query(ctx context.Context, sessionID string, queryVector []float32, topK int, filter MetadataFilter) ([]domain.Chunk, error)
    DeleteSession(ctx context.Context, sessionID string) error
}
```

---

## Task Board

### Phase 1 — Foundation (Interfaces + Config + Project Scaffold)

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| BE-001 | Backend | Initialize giftsense-backend Go module and directory scaffold | — | PENDING |
| BE-002 | Backend | Implement config system with env var loading and fail-fast validation | BE-001 | PENDING |
| BE-003 | Backend | Define all domain types and sentinel errors | BE-001 | PENDING |
| BE-004 | Backend | Define all port interfaces (Embedder, LLMClient, VectorStore) | BE-003 | PENDING |

---

#### BE-001 — Initialize giftsense-backend Go module and directory scaffold

**Team:** Backend
**Title:** Create giftsense-backend Go module with full directory tree
**Scope:**
```
giftsense-backend/
├── go.mod           (module: github.com/<user>/giftsense-backend, go 1.22)
├── cmd/server/main.go        (empty stub — just package main + func main)
├── config/config.go          (empty stub)
├── internal/domain/          (directory only, no files yet)
├── internal/port/            (directory only, no files yet)
├── internal/usecase/         (directory only, no files yet)
├── internal/adapter/openai/  (directory only)
├── internal/adapter/vectorstore/ (directory only)
├── internal/adapter/linkgen/ (directory only)
├── internal/delivery/http/   (directory only)
├── internal/delivery/dto/    (directory only)
└── .env.example              (empty placeholder)
```
**Dependencies:** none
**Acceptance Criteria:**
- `go build ./...` succeeds (only stubs, no logic)
- `go.mod` declares module path and Go 1.22+
- All required directories exist
**Test Requirement:** None for scaffold task — verified by `go build`

---

#### BE-002 — Implement config system

**Team:** Backend
**Title:** Config system: load all 16 env vars, fail-fast on missing secrets
**Scope:** `config/config.go`
**Dependencies:** BE-001
**Implementation:**
```go
type Config struct {
    OpenAIAPIKey         string
    ChatModel            string
    EmbeddingModel       string
    EmbeddingDimensions  int
    MaxTokens            int
    TopK                 int
    NumRetrievalQueries  int
    PineconeAPIKey       string
    PineconeIndexName    string
    PineconeEnvironment  string
    MaxFileSizeBytes     int64
    MaxProcessedMessages int
    ChunkWindowSize      int
    ChunkOverlapSize     int
    Port                 string
    AllowedOrigins       []string
}

func Load() (*Config, error) { ... }
```
- `Load()` calls `os.Getenv()` for every field (it is the ONLY file permitted to do so)
- Returns error if `OPENAI_API_KEY` or `PINECONE_API_KEY` or `PINECONE_ENVIRONMENT` are empty
- Applies defaults for all optional fields
- Logs all values at INFO level on startup; replaces secret values with `[REDACTED]`
- `ALLOWED_ORIGINS` is parsed from a comma-separated string

**Acceptance Criteria:**
- `config.Load()` returns an error with a descriptive message if any required env var is missing
- All defaults match the Environment Variables Registry table above
- No `os.Getenv()` call exists anywhere outside `config/config.go`

**Test Requirement:**
```
TestLoad_ShouldReturnConfig_WhenAllRequiredEnvVarsAreSet
TestLoad_ShouldReturnError_WhenOpenAIAPIKeyIsMissing
TestLoad_ShouldReturnError_WhenPineconeAPIKeyIsMissing
TestLoad_ShouldApplyDefaults_WhenOptionalEnvVarsAreAbsent
TestLoad_ShouldParseAllowedOrigins_WhenCommaSeparated
```

---

#### BE-003 — Define domain types and sentinel errors

**Team:** Backend
**Title:** Domain types: Message, Chunk, RecipientDetails, BudgetTier, GiftSuggestion, PersonalityInsight, sentinel errors
**Scope:** `internal/domain/conversation.go`, `internal/domain/recipient.go`, `internal/domain/suggestion.go`, `internal/domain/errors.go`
**Dependencies:** BE-001
**NO third-party imports in this package.**

```go
// conversation.go
type Message struct {
    Index     int
    Sender    string
    Text      string
    Timestamp time.Time
    IsMedia   bool
}

type Chunk struct {
    ID             string
    SessionID      string
    AnonymizedText string
    StartIndex     int
    EndIndex       int
    Metadata       ChunkMetadata
}

type ChunkMetadata struct {
    Topics           []string
    EmotionalMarkers []string
    HasPreference    bool
    HasWish          bool
}

// recipient.go
type BudgetTier string

const (
    BudgetTierBudget   BudgetTier = "BUDGET"
    BudgetTierMidRange BudgetTier = "MID_RANGE"
    BudgetTierPremium  BudgetTier = "PREMIUM"
    BudgetTierLuxury   BudgetTier = "LUXURY"
)

type BudgetRange struct {
    Tier   BudgetTier
    MinINR int
    MaxINR int  // 0 means no upper bound (Luxury tier)
}

var BudgetRanges = map[BudgetTier]BudgetRange{
    BudgetTierBudget:   {BudgetTierBudget,   500,   1000},
    BudgetTierMidRange: {BudgetTierMidRange, 1000,  5000},
    BudgetTierPremium:  {BudgetTierPremium,  5000,  15000},
    BudgetTierLuxury:   {BudgetTierLuxury,   15000, 0},
}

type RecipientDetails struct {
    Name      string
    Relation  string
    Gender    string
    Occasion  string
    Budget    BudgetRange
}

// suggestion.go
type PersonalityInsight struct {
    Insight         string `json:"insight"`
    EvidenceSummary string `json:"evidence_summary"`
}

type ShoppingLinks struct {
    Amazon         string `json:"amazon"`
    Flipkart       string `json:"flipkart"`
    GoogleShopping string `json:"google_shopping"`
}

type GiftSuggestion struct {
    Name              string        `json:"name"`
    Reason            string        `json:"reason"`
    EstimatedPriceINR string        `json:"estimated_price_inr"`
    Category          string        `json:"category"`
    Links             ShoppingLinks `json:"links"`
}

type AnalysisResult struct {
    PersonalityInsights []PersonalityInsight `json:"personality_insights"`
    GiftSuggestions     []GiftSuggestion     `json:"gift_suggestions"`
}

// errors.go
var (
    ErrFileTooLarge        = errors.New("file exceeds maximum allowed size")
    ErrConversationTooShort = errors.New("conversation has too few messages to analyze")
    ErrInvalidBudgetTier   = errors.New("invalid budget tier")
    ErrInvalidSessionID    = errors.New("invalid session ID format")
    ErrInvalidFileType     = errors.New("only .txt files are accepted")
    ErrLLMResponseInvalid  = errors.New("LLM returned invalid or non-conformant JSON")
    ErrRetrievalFailed     = errors.New("retrieval returned no relevant context")
    ErrAllSuggestionsFiltered = errors.New("all suggestions violated budget constraints")
)
```

**Acceptance Criteria:**
- All types compile with no external dependencies
- `BudgetRanges` map contains exactly the 4 tiers with correct INR values
- All sentinel errors are defined and exported

**Test Requirement:**
```
TestBudgetRanges_ShouldHaveFourTiers
TestBudgetRanges_ShouldHaveCorrectINRValues
```

---

#### BE-004 — Define port interfaces

**Team:** Backend
**Title:** Port interfaces: Embedder, LLMClient, VectorStore with MetadataFilter
**Scope:** `internal/port/embedder.go`, `internal/port/llm.go`, `internal/port/vectorstore.go`
**Dependencies:** BE-003
**NO third-party imports. Depends only on `domain` package.**

Implement interfaces exactly as specified in the Go Interface Contracts section above.

**Acceptance Criteria:**
- All three interfaces compile
- `VectorStore` has exactly 3 methods: `Upsert`, `Query`, `DeleteSession`
- `LLMClient.Complete` accepts `CompletionOptions` struct with `MaxTokens` and `JSONMode` fields
- `MetadataFilter` has `HasPreference *bool`, `HasWish *bool`, `Topics []string`

**Test Requirement:** None — interfaces have no logic; compiled and verified by adapter tests.

---

### Phase 2 — Core Pipeline (Parser → Anonymizer → Chunker)

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| BE-005 | Backend | Conversation parser with smart sampling | BE-003, BE-004 | PENDING |
| BE-006 | Backend | Anonymizer (sender-seeded NER + regex pseudonymization) | BE-003, BE-004 | PENDING |
| BE-007 | Backend | Sliding window chunker with metadata enrichment | BE-003, BE-004 | PENDING |

> BE-005, BE-006, and BE-007 can run in **parallel**.

---

#### BE-005 — Conversation parser

**Team:** Backend — BE-Pipeline Agent
**Title:** Parse WhatsApp and plain-text conversations; apply smart sampling (max 400 messages)
**Scope:** `internal/usecase/parse.go`, `internal/usecase/parse_test.go`
**Dependencies:** BE-003, BE-004

**Implementation details:**
- `ParseConversation(text string, maxMessages int) ([]domain.Message, error)`
- WhatsApp format detection: lines matching `[DD/MM/YYYY, HH:MM:SS] Sender: text` regex
  - Multi-line messages: non-matching lines are appended to previous message's text
  - Skip system messages: lines containing "end-to-end encrypted", "Messages and calls are"
  - `IsMedia: true` for lines containing `<Media omitted>` or `image omitted`
- Plain-text detection: look for `You:`, `Friend:`, `Me:` prefixes; if none found, treat whole text as single block
- Smart sampling (applied when `len(messages) > maxMessages`):
  1. Take the most recent 25% of messages (recency bias)
  2. From the remaining, take a spread sample across the conversation to fill up to `maxMessages`
  3. Result: lexically diverse sample reflecting current preferences

**Acceptance Criteria:**
- Parses a sample WhatsApp `.txt` export correctly into typed `[]domain.Message`
- Filters system messages and media lines
- Returns `domain.ErrConversationTooShort` if < 5 parseable messages remain after filtering
- Returns at most `maxMessages` messages after sampling

**Test Requirement:**
```
TestParseConversation_ShouldParseWhatsAppFormat_WhenValidExportProvided
TestParseConversation_ShouldFilterSystemMessages_WhenWhatsAppFormatDetected
TestParseConversation_ShouldSetIsMedia_WhenMediaOmittedLineDetected
TestParseConversation_ShouldReturnError_WhenFewerThanFiveMessages
TestParseConversation_ShouldCapMessages_WhenExceedsMaxMessages
TestParseConversation_ShouldHandlePlainText_WhenNoWhatsAppFormatDetected
```

---

#### BE-006 — Anonymizer

**Team:** Backend — BE-Pipeline Agent
**Title:** Anonymize messages: replace real names/places with stable tokens ([Person_A], [City_1], etc.)
**Scope:** `internal/usecase/anonymize.go`, `internal/usecase/anonymize_test.go`
**Dependencies:** BE-003

**Implementation details:**
- `AnonymizeMessages(messages []domain.Message) ([]domain.Message, map[string]string, error)`
  - Returns anonymized messages AND the token map (only for in-memory use — NEVER log or store)
- Seed entity list from parsed sender names in the message slice
- `[Person_A]` = first unique sender (typically the uploader)
- `[Person_B]` = second unique sender (the recipient)
- `[Person_C]`, `[Person_D]`, ... for additional names found in message bodies
- `[City_1]`, `[City_2]`, ... for place names
- `[Company_1]`, ... for organization names
- Regex for capitalized proper nouns in message bodies (heuristic): `\b[A-Z][a-z]{2,}\b` — filter against a small stop-list of common English capitalized words (I, Monday, Yes, etc.)
- Stable: the same name always maps to the same token within a session
- **NEVER log the token map or any original entity values at any log level**

**Acceptance Criteria:**
- All sender names from the message slice are replaced with `[Person_X]` tokens
- The same name appearing 5 times becomes the same token all 5 times
- Returns a non-nil token map for in-memory use
- Token map is NOT logged at any level (verify in code review)

**Test Requirement:**
```
TestAnonymizeMessages_ShouldReplaceAllSenderNames_WhenMessagesProvided
TestAnonymizeMessages_ShouldUseStableTokens_WhenSameNameAppearsMultipleTimes
TestAnonymizeMessages_ShouldAssignPersonA_ToFirstSender
TestAnonymizeMessages_ShouldAssignPersonB_ToSecondSender
TestAnonymizeMessages_ShouldNotModifyNonPIIText_WhenNoNamesPresent
```

---

#### BE-007 — Chunker

**Team:** Backend — BE-Pipeline Agent
**Title:** Sliding window chunker (configurable window+overlap) with heuristic metadata enrichment
**Scope:** `internal/usecase/chunk.go`, `internal/usecase/chunk_test.go`
**Dependencies:** BE-003

**Implementation details:**
- `ChunkMessages(sessionID string, messages []domain.Message, windowSize int, overlapSize int) ([]domain.Chunk, error)`
- Sliding window: start at index 0, advance by `(windowSize - overlapSize)` each step
- Concatenate messages in a window: `"[Sender]: text\n"` format
- Generate chunk ID: `fmt.Sprintf("%s_chunk_%d", sessionID, chunkIndex)`
- Metadata enrichment (keyword matching against a topic/emotion word list):
  - `Topics`: check for keywords in: `cooking food travel reading craft hobby music sport outdoor fitness creative work job career`
  - `EmotionalMarkers`: check for keywords in: `laugh haha lol love miss excited happy sad worried nervous fun amazing wonderful great`
  - `HasPreference`: true if window contains "want to", "love to", "I like", "I enjoy", "prefer", "favorite", "favourite"
  - `HasWish`: true if window contains "someday", "wish", "dream of", "plan to", "going to", "would love"
- Skip windows where all messages have `IsMedia: true`

**Acceptance Criteria:**
- `ChunkMessages` on 10 messages with window=8, overlap=3 produces at least 2 chunks
- Each chunk has a unique ID with the sessionID prefix
- `HasPreference` is true for a chunk containing "I love hiking"
- `HasWish` is true for a chunk containing "I wish I could travel more"
- No chunk contains media-only messages (IsMedia filter)

**Test Requirement:**
```
TestChunkMessages_ShouldProduceCorrectChunkCount_WhenWindowAndOverlapApplied
TestChunkMessages_ShouldGenerateUniqueChunkIDs_WithSessionIDPrefix
TestChunkMessages_ShouldSetHasPreference_WhenPreferenceKeywordPresent
TestChunkMessages_ShouldSetHasWish_WhenWishKeywordPresent
TestChunkMessages_ShouldSkipMediaOnlyWindows_WhenIsMediaTrue
TestChunkMessages_ShouldEnrichTopics_WhenCookingKeywordPresent
```

---

### Phase 3 — Vector Store + Retrieval

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| BE-008 | Backend | Pinecone VectorStore adapter + in-memory test double | BE-004 | PENDING |
| BE-009 | Backend | Multi-query retrieval constructor + heuristic re-ranker | BE-003, BE-004 | PENDING |

> BE-008 and BE-009 can run in **parallel**.

---

#### BE-008 — Pinecone VectorStore adapter

**Team:** Backend — BE-Adapters Agent
**Title:** Implement `VectorStore` interface: Pinecone adapter (Upsert/Query/DeleteSession) + in-memory test double
**Scope:** `internal/adapter/vectorstore/pinecone.go`, `internal/adapter/vectorstore/pinecone_test.go`, `internal/adapter/vectorstore/memory.go`, `internal/adapter/vectorstore/memory_test.go`
**Dependencies:** BE-004

**Pinecone adapter (`pinecone.go`):**
- Constructor: `NewPineconeStore(apiKey, indexName, environment string, dimensions int) (*PineconeStore, error)`
- Uses `github.com/pinecone-io/go-pinecone` SDK — wrap ALL SDK calls; never import pinecone in usecase
- `Upsert`: batch upsert all chunk vectors into the `sessionID` namespace
  - Each record: `{id: chunk.ID, values: vector, metadata: {has_preference, has_wish, topics, emotional_markers, chunk_index, message_start, message_end}}`
  - **Do NOT include `anonymized_text` in Pinecone metadata** (text stays in Go memory only)
  - Use goroutines + errgroup for concurrent upsert batches
- `Query`: query namespace `sessionID`, filter by `MetadataFilter`, return top-K `domain.Chunk` objects
  - NOTE: Pinecone returns record IDs and metadata; the caller must re-join text from a local chunk map (this is done in the use case, not here)
  - Return `domain.Chunk` objects populated from metadata; set `AnonymizedText: ""` (caller joins it)
- `DeleteSession`: delete entire namespace `sessionID` from the Pinecone index
- Retry logic: exponential backoff for 429/500/502/503/504 HTTP errors (wait 1s, 2s, 4s; max 3 retries)

**In-memory adapter (`memory.go`):**
- `NewMemoryStore() *MemoryStore`
- Thread-safe (uses sync.Mutex) in-memory map: `sessionID -> []storedVector`
- `Upsert`: store chunk IDs + vectors + metadata in the map
- `Query`: compute cosine similarity between queryVector and all stored vectors for the session; return topK
- `DeleteSession`: delete the sessionID key from the map (no-op is also acceptable — never fails)

**Acceptance Criteria (pinecone):**
- Implements `port.VectorStore` interface
- Constructor fails with clear error if apiKey or indexName are empty
- Upsert does NOT include any text fields in Pinecone metadata

**Acceptance Criteria (memory):**
- Implements `port.VectorStore` interface
- `Query` returns correct top-K by cosine similarity
- `DeleteSession` clears all vectors for that session

**Test Requirement:**
```
// memory_test.go (no external API calls)
TestMemoryStore_ShouldReturnTopK_WhenChunksAreUpserted
TestMemoryStore_ShouldIsolateSessionsByID
TestMemoryStore_ShouldReturnEmpty_AfterDeleteSession
TestMemoryStore_ShouldFilterByHasPreference_WhenMetadataFilterApplied

// pinecone_test.go (constructor + config validation only; integration tests deferred)
TestNewPineconeStore_ShouldReturnError_WhenAPIKeyIsEmpty
TestNewPineconeStore_ShouldReturnError_WhenIndexNameIsEmpty
```

---

#### BE-009 — Retrieval strategy and re-ranker

**Team:** Backend — BE-Pipeline Agent
**Title:** Construct 4 multi-retrieval queries from recipient context; embed them concurrently; deduplicate and re-rank retrieved chunks
**Scope:** `internal/usecase/retrieve.go`, `internal/usecase/retrieve_test.go`
**Dependencies:** BE-003, BE-004

**Implementation details:**
- `BuildRetrievalQueries(recipient domain.RecipientDetails, numQueries int) []string`
  - Returns exactly `numQueries` queries (default 4):
    1. "What activities, hobbies, passions, and interests does this person mention wanting to pursue or already enjoying?"
    2. "What personality traits, emotional patterns, and communication style does this person show? What makes them happy or laugh?"
    3. "What things has this person explicitly said they want, wish for, or plan to do someday? Any specific items or experiences mentioned?"
    4. "What is the nature of this relationship and what recurring shared themes or experiences do they have together?"
  - Each query is prefixed with recipient context: `"For a {occasion} gift for {relation}: {base_query}"`
- `RetrieveAndRerank(ctx context.Context, sessionID string, queries []string, chunksByID map[string]domain.Chunk, embedder port.Embedder, store port.VectorStore, topK int) ([]domain.Chunk, error)`
  - Embed all queries concurrently using `errgroup`
  - Query `store` for each embedded query with `topK` results
  - Collect all returned chunk IDs; join with `chunksByID` map to get `AnonymizedText`
  - Deduplicate: if same chunk ID appears in multiple queries, keep once
  - Re-rank: sort by (1) `HasPreference || HasWish` first, then (2) by occurrence count across queries
  - Return at most `numQueries * topK` chunks (deduplicated)

**Acceptance Criteria:**
- `BuildRetrievalQueries` returns exactly `numQueries` non-empty strings
- All queries include recipient relation and occasion context
- `RetrieveAndRerank` returns deduplicated chunks (no duplicate IDs)
- Chunks with `HasPreference=true` appear before those without

**Test Requirement:**
```
TestBuildRetrievalQueries_ShouldReturnCorrectCount_WhenNumQueriesProvided
TestBuildRetrievalQueries_ShouldIncludeRecipientContext_InEveryQuery
TestRetrieveAndRerank_ShouldDeduplicateChunks_WhenSameChunkReturnedByMultipleQueries
TestRetrieveAndRerank_ShouldPrioritizeHasPreferenceChunks_WhenReranking
```

---

### Phase 4 — LLM Integration (Embedder + LLM + Link Generator + Analyze Orchestrator)

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| BE-010 | Backend | OpenAI Embedder adapter | BE-004 | PENDING |
| BE-011 | Backend | OpenAI LLMClient adapter (JSON mode, GPT-4o) | BE-004 | PENDING |
| BE-012 | Backend | Shopping link generator (Amazon/Flipkart/Google URL builder) | BE-003 | PENDING |
| BE-013 | Backend | AnalyzeConversation use case orchestrator (13-step pipeline) | BE-005 thru BE-012 | PENDING |

> BE-010, BE-011, and BE-012 can run in **parallel**. BE-013 depends on all three.

---

#### BE-010 — OpenAI Embedder adapter

**Team:** Backend — BE-Adapters Agent
**Title:** Implement `port.Embedder` using OpenAI `text-embedding-3-small` SDK
**Scope:** `internal/adapter/openai/embedder.go`, `internal/adapter/openai/embedder_test.go`
**Dependencies:** BE-004

**Implementation:**
- Constructor: `NewEmbedder(apiKey, model string, dimensions int) *OpenAIEmbedder`
- `Embed(ctx context.Context, texts []string) ([][]float32, error)`
  - Uses `github.com/openai/openai-go` SDK
  - Batch all texts in a single API call (SDK supports multi-input embedding)
  - Returns error wrapped with `fmt.Errorf("embedding failed: %w", err)` on API failure
- **NEVER log the input texts** — they are anonymized chunks, but logging is still prohibited by privacy policy
- The `openai-go` SDK must ONLY be imported in this file (not in usecase or domain)

**Acceptance Criteria:**
- Implements `port.Embedder` interface
- Constructor fails if apiKey is empty
- Returns `[][]float32` with one vector per input text, each of length `dimensions`

**Test Requirement:**
```
TestNewEmbedder_ShouldReturnError_WhenAPIKeyIsEmpty
// Integration test (skip in CI without real API key):
TestEmbed_ShouldReturnCorrectDimensions_WhenCalledWithRealAPI
```

---

#### BE-011 — OpenAI LLMClient adapter

**Team:** Backend — BE-Adapters Agent
**Title:** Implement `port.LLMClient` using OpenAI chat completions with JSON mode
**Scope:** `internal/adapter/openai/llm.go`, `internal/adapter/openai/llm_test.go`
**Dependencies:** BE-004

**Implementation:**
- Constructor: `NewLLMClient(apiKey, model string, maxTokens int) *OpenAILLMClient`
- `Complete(ctx context.Context, prompt string, opts port.CompletionOptions) (string, error)`
  - Constructs a chat completion request: system prompt = GiftSense persona, user message = `prompt`
  - If `opts.JSONMode == true`: set `ResponseFormat` to JSON object mode
  - Uses `cfg.MaxTokens` from opts if provided; falls back to client default
  - Retry logic: exponential backoff for 429/500/503 (1s, 2s, 4s; max 3 retries)
  - Returns the content string of the first choice's message
  - **NEVER log the prompt content or the response content** (contains anonymized conversation chunks)

**GiftSense system prompt (hardcoded in this adapter):**
```
You are a warm, insightful gift recommendation assistant who reads between the lines
of conversations to understand people deeply.
RULES:
1. Only infer traits and suggest gifts supported by evidence in the provided conversation context.
2. Every gift suggestion MUST have an estimated price within the stated budget range.
3. Respond ONLY with a valid JSON object matching this schema:
{
  "personality_insights": [{"insight": "string", "evidence_summary": "string"}],
  "gift_suggestions": [{"name": "string", "reason": "string", "estimated_price_inr": "string", "category": "string"}]
}
4. Personality insights should be warm and human — written as a perceptive friend, not a corporate analyst.
5. Gift names must be specific enough to search for (e.g., "Pottery starter kit with air-dry clay" not "craft supplies").
```

**Acceptance Criteria:**
- Implements `port.LLMClient` interface
- Constructor fails if apiKey is empty
- Returns raw JSON string from GPT-4o
- Retry logic triggers on 429/500 status codes

**Test Requirement:**
```
TestNewLLMClient_ShouldReturnError_WhenAPIKeyIsEmpty
TestComplete_ShouldRetry_WhenOpenAIReturnsRateLimitError
```

---

#### BE-012 — Shopping link generator

**Team:** Backend — BE-Adapters Agent
**Title:** Generate Amazon India, Flipkart, and Google Shopping URLs for each gift suggestion
**Scope:** `internal/adapter/linkgen/shopping.go`, `internal/adapter/linkgen/shopping_test.go`
**Dependencies:** BE-003

**Implementation:**
- `GenerateLinks(giftName string, budget domain.BudgetRange) domain.ShoppingLinks`
- Pure function — no I/O, no network calls, no errors possible
- Amazon India URL:
  ```
  https://www.amazon.in/s?k={url_encoded_name}&rh=p_36%3A{min_paise}-{max_paise}
  ```
  - Prices in paise: `minPaise = budget.MinINR * 100`, `maxPaise = budget.MaxINR * 100`
  - For `Luxury` tier (MaxINR == 0): omit the max paise suffix → `rh=p_36%3A{min_paise}`
- Flipkart URL:
  ```
  https://www.flipkart.com/search?q={url_encoded_name}&p[]=facets.price_range.from%3D{minINR}&p[]=facets.price_range.to%3D{maxINR}
  ```
  - For `Luxury` tier: omit `price_range.to` parameter
- Google Shopping URL:
  ```
  https://www.google.com/search?q={url_encoded_name}+under+%E2%82%B9{maxINR}&tbm=shop
  ```
  - For `Luxury` tier: `https://www.google.com/search?q={url_encoded_name}+premium+gift&tbm=shop`
- Use `net/url.QueryEscape()` for gift name URL encoding

**Acceptance Criteria:**
- Returns all three URLs for all four budget tiers
- Luxury tier omits upper bound on Amazon and Flipkart
- Gift names with spaces and special characters are correctly URL-encoded
- Amazon prices are in paise (₹500 → 50000)

**Test Requirement:**
```
TestGenerateLinks_ShouldReturnAmazonURL_WithCorrectPaiseValues
TestGenerateLinks_ShouldReturnFlipkartURL_WithCorrectINRValues
TestGenerateLinks_ShouldHandleLuxuryTier_WithNoUpperBound
TestGenerateLinks_ShouldURLEncodeGiftName_WhenSpacesPresent
TestGenerateLinks_ShouldEncodeRupeeSymbol_InGoogleShoppingURL
```

---

#### BE-013 — AnalyzeConversation use case orchestrator

**Team:** Backend — BE-Pipeline Agent
**Title:** Orchestrate the full 13-step RAG pipeline: parse → anonymize → chunk → embed → upsert → query → rerank → prompt → complete → validate → links → delete → return
**Scope:** `internal/usecase/analyze.go`, `internal/usecase/analyze_test.go`
**Dependencies:** BE-005, BE-006, BE-007, BE-008, BE-009, BE-010, BE-011, BE-012

**Struct:**
```go
type AnalyzeService struct {
    embedder   port.Embedder
    llmClient  port.LLMClient
    vectorStore port.VectorStore
    cfg        AnalyzeConfig
}

type AnalyzeConfig struct {
    MaxProcessedMessages int
    ChunkWindowSize      int
    ChunkOverlapSize     int
    TopK                 int
    NumRetrievalQueries  int
    MaxTokens            int
}

func NewAnalyzeService(embedder port.Embedder, llmClient port.LLMClient, store port.VectorStore, cfg AnalyzeConfig) *AnalyzeService

func (s *AnalyzeService) Analyze(ctx context.Context, sessionID string, conversationText string, recipient domain.RecipientDetails) (domain.AnalysisResult, error)
```

**Pipeline steps in `Analyze()`:**
1. `parse.ParseConversation(text, cfg.MaxProcessedMessages)` → `[]domain.Message`
2. `anonymize.AnonymizeMessages(messages)` → anonymized messages + token map (in-memory only)
3. `chunk.ChunkMessages(sessionID, anonMessages, cfg.ChunkWindowSize, cfg.ChunkOverlapSize)` → `[]domain.Chunk`
4. Build `chunksByID map[string]domain.Chunk` for text look-up
5. Extract chunk texts → `[]string`
6. `s.embedder.Embed(ctx, chunkTexts)` → `[][]float32` (concurrent internally via embedder)
7. `s.vectorStore.Upsert(ctx, sessionID, chunks, vectors)`
8. `retrieve.BuildRetrievalQueries(recipient, cfg.NumRetrievalQueries)` → `[]string`
9. `retrieve.RetrieveAndRerank(ctx, sessionID, queries, chunksByID, s.embedder, s.vectorStore, cfg.TopK)` → `[]domain.Chunk`
10. Build GPT-4o prompt string (system prompt in LLM adapter, user message assembled here):
    - Include recipient details (name, relation, gender, occasion, budget tier)
    - Include retrieved anonymized chunks as numbered context blocks
    - Include budget enforcement instruction
11. `s.llmClient.Complete(ctx, prompt, port.CompletionOptions{MaxTokens: cfg.MaxTokens, JSONMode: true})` → JSON string
12. Parse and validate JSON → `domain.AnalysisResult`
    - Unmarshal the JSON into internal struct; validate required fields present
    - Filter gift suggestions with prices outside budget range
    - Return `domain.ErrAllSuggestionsFiltered` if zero suggestions remain
    - Return `domain.ErrLLMResponseInvalid` if JSON is malformed
13. Generate shopping links for each suggestion via `linkgen.GenerateLinks`
14. `s.vectorStore.DeleteSession(ctx, sessionID)` — always called, even if step 12-13 fail (use defer)
15. Return `domain.AnalysisResult`

**Acceptance Criteria:**
- `DeleteSession` is called via `defer` so it runs even on error paths
- All external calls go through interface (no direct SDK imports)
- Budget-violating suggestions are filtered before returning
- Returns `domain.ErrConversationTooShort` when parser returns that error

**Test Requirement (use MemoryStore + mock embedder + mock LLM client):**
```
TestAnalyze_ShouldReturnAnalysisResult_WhenValidInputProvided
TestAnalyze_ShouldCallDeleteSession_WhenPipelineSucceeds
TestAnalyze_ShouldCallDeleteSession_WhenPipelineFails
TestAnalyze_ShouldFilterBudgetViolations_WhenLLMReturnsSuggestionsOutsideRange
TestAnalyze_ShouldReturnError_WhenConversationTooShort
TestAnalyze_ShouldReturnError_WhenLLMReturnsInvalidJSON
```

---

### Phase 5 — HTTP Layer (Gin Routes + Handlers + Middleware + Validation)

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| BE-014 | Backend | DTOs: request/response structs and JSON mappings | BE-003 | PENDING |
| BE-015 | Backend | Gin handlers: POST /api/v1/analyze, GET /health, GET /metrics | BE-013, BE-014 | PENDING |
| BE-016 | Backend | Middleware: CORS, structured logging, rate limiting; Validator | BE-002, BE-014 | PENDING |
| BE-017 | Backend | main.go: wire all deps and start Gin server | BE-002 thru BE-016 | PENDING |

> BE-014 and BE-016 can run in **parallel**. BE-015 depends on BE-013 + BE-014. BE-017 depends on all.

---

#### BE-014 — DTOs

**Team:** Backend — BE-HTTP Agent
**Title:** Request and response DTOs for the analyze endpoint
**Scope:** `internal/delivery/dto/request.go`, `internal/delivery/dto/response.go`
**Dependencies:** BE-003

```go
// request.go — multipart form fields
type AnalyzeRequest struct {
    SessionID  string `form:"session_id"`
    Name       string `form:"name"`
    Relation   string `form:"relation"`
    Gender     string `form:"gender"`
    Occasion   string `form:"occasion"`
    BudgetTier string `form:"budget_tier"`
    // Conversation file is read via c.FormFile("conversation") in handler
}

// response.go
type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**Acceptance Criteria:**
- DTOs compile with no external dependencies beyond standard library
- `AnalyzeRequest` uses `form:` tags for multipart binding

**Test Requirement:** None (data structs; verified by handler tests).

---

#### BE-015 — Gin handlers

**Team:** Backend — BE-HTTP Agent
**Title:** Implement POST /api/v1/analyze, GET /health, GET /metrics handlers
**Scope:** `internal/delivery/http/handler.go`, `internal/delivery/http/handler_test.go`
**Dependencies:** BE-013, BE-014

**Handler struct:**
```go
type Handler struct {
    analyzeService *usecase.AnalyzeService
    cfg            *config.Config
}
func NewHandler(svc *usecase.AnalyzeService, cfg *config.Config) *Handler
```

**POST /api/v1/analyze:**
1. Set `c.Request.Body` size limit to `cfg.MaxFileSizeBytes` via Gin's `MaxMultipartMemory`
2. Read and validate `AnalyzeRequest` from multipart form
3. Read conversation file via `c.FormFile("conversation")`
4. Delegate all validation to `validator.ValidateAnalyzeRequest(req, fileHeader)` → return 400/422 on failure
5. Read file content as string
6. Map `BudgetTier` string → `domain.BudgetRange`; return 400 if invalid
7. Call `analyzeService.Analyze(c.Request.Context(), req.SessionID, text, recipient)`
8. Map errors to HTTP status:
   - `ErrFileTooLarge` → 413
   - `ErrConversationTooShort` → 422
   - `ErrInvalidSessionID` → 400
   - All other errors → 500
9. Return 200 with `SuccessResponse{Data: result, Message: "Analysis complete"}`

**GET /health:** return `200 {"status": "ok"}`

**GET /metrics:** return `200` with in-process counters (total sessions, successful sessions, avg latency)

**Acceptance Criteria:**
- Returns 200 with correct JSON shape on success
- Returns 413 if file > `cfg.MaxFileSizeBytes`
- Returns 422 if conversation too short
- Returns 400 if budget tier is invalid

**Test Requirement (use httptest):**
```
TestAnalyzeHandler_ShouldReturn200_WhenValidRequestProvided
TestAnalyzeHandler_ShouldReturn413_WhenFileTooLarge
TestAnalyzeHandler_ShouldReturn422_WhenConversationTooShort
TestAnalyzeHandler_ShouldReturn400_WhenBudgetTierInvalid
TestAnalyzeHandler_ShouldReturn400_WhenSessionIDMissing
TestHealthHandler_ShouldReturn200_Always
```

---

#### BE-016 — Middleware and Validator

**Team:** Backend — BE-HTTP Agent
**Title:** CORS middleware, structured request logging, rate limiting, and request field validator
**Scope:** `internal/delivery/http/middleware.go`, `internal/delivery/http/validator.go`, `internal/delivery/http/middleware_test.go`, `internal/delivery/http/validator_test.go`
**Dependencies:** BE-002, BE-014

**Middleware:**
- `CORSMiddleware(allowedOrigins []string) gin.HandlerFunc`
  - Allows `Origin` header if it matches any entry in `allowedOrigins`
  - Sets: `Access-Control-Allow-Methods: POST, GET, OPTIONS`, `Access-Control-Allow-Headers: Content-Type`
  - Handles preflight `OPTIONS` requests with 204
- `RequestLogger() gin.HandlerFunc`
  - Logs: method, path, status, latency using `slog` structured logging
  - **NEVER logs request body or any form field values**
- `RateLimiter(requestsPerMinute int) gin.HandlerFunc`
  - In-process token bucket per client IP (`c.ClientIP()`)
  - Returns 429 when limit exceeded

**Validator (`validator.go`):**
- `ValidateAnalyzeRequest(req dto.AnalyzeRequest, fileHeader *multipart.FileHeader, maxFileSizeBytes int64) error`
  - `session_id`: must match UUID v4 regex `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
  - `name`: required, non-empty, max 100 chars
  - `occasion`: required, non-empty
  - `budget_tier`: must be one of `BUDGET`, `MID_RANGE`, `PREMIUM`, `LUXURY`
  - `file`: extension must be `.txt`, size must not exceed `maxFileSizeBytes`

**Acceptance Criteria:**
- CORS middleware returns correct headers for allowed origin
- CORS middleware rejects (no ACAO header) for unlisted origin
- Validator rejects invalid UUID format
- Validator rejects non-.txt file extension
- Request logger does not log form field values (verified by test output capture)

**Test Requirement:**
```
TestCORSMiddleware_ShouldAllowConfiguredOrigin
TestCORSMiddleware_ShouldNotSetHeadersForUnknownOrigin
TestCORSMiddleware_ShouldHandlePreflightRequest
TestValidateAnalyzeRequest_ShouldPass_WhenAllFieldsValid
TestValidateAnalyzeRequest_ShouldFail_WhenSessionIDInvalidFormat
TestValidateAnalyzeRequest_ShouldFail_WhenBudgetTierInvalid
TestValidateAnalyzeRequest_ShouldFail_WhenFileIsNotTxt
TestValidateAnalyzeRequest_ShouldFail_WhenFileExceedsMaxSize
```

---

#### BE-017 — main.go

**Team:** Backend — BE-HTTP Agent
**Title:** Wire all dependencies and start the Gin server
**Scope:** `cmd/server/main.go`, `.env.example`
**Dependencies:** BE-002 thru BE-016

**main.go responsibilities:**
1. `config.Load()` — fail-fast if missing secrets
2. Construct adapters: `NewEmbedder(cfg)`, `NewLLMClient(cfg)`, `NewPineconeStore(cfg)`
3. Construct use case: `NewAnalyzeService(embedder, llmClient, store, analyzeConfig)`
4. Construct handler: `NewHandler(service, cfg)`
5. Set up Gin with middleware: `CORSMiddleware`, `RequestLogger`, `RateLimiter`
6. Register routes:
   - `POST /api/v1/analyze` → `handler.Analyze`
   - `GET /health` → `handler.Health`
   - `GET /metrics` → `handler.Metrics`
7. Start server on `cfg.Port` (default 8080)

**.env.example:**
```
# Required secrets
OPENAI_API_KEY=your_openai_api_key_here
PINECONE_API_KEY=your_pinecone_api_key_here
PINECONE_ENVIRONMENT=us-east-1

# Optional — sensible defaults provided
CHAT_MODEL=gpt-4o
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_DIMENSIONS=1536
MAX_TOKENS=1000
TOP_K=3
NUM_RETRIEVAL_QUERIES=4
PINECONE_INDEX_NAME=giftsense
MAX_FILE_SIZE_BYTES=2097152
MAX_PROCESSED_MESSAGES=400
CHUNK_WINDOW_SIZE=8
CHUNK_OVERLAP_SIZE=3
ALLOWED_ORIGINS=http://localhost:5173
PORT=8080
```

**Acceptance Criteria:**
- `go build ./...` succeeds with no errors
- Application exits immediately with clear error if `OPENAI_API_KEY` is not set
- All env var names in `.env.example` match the Environment Variables Registry table exactly
- `GET /health` returns 200 on a running server

**Test Requirement:** Verified by `go build` + manual smoke test of `/health` endpoint.

---

### Phase 6 — Frontend Foundation (React Scaffold + Session UUID + Routing)

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| FE-001 | Frontend | Vite + React + Tailwind scaffold with project structure | — | PENDING |
| FE-002 | Frontend | Session hook (crypto.randomUUID) + API client | FE-001 | PENDING |
| FE-003 | Frontend | App.jsx: three-screen state machine + useAnalyze hook | FE-001, FE-002 | PENDING |

> FE-001 through FE-003 are sequential.

---

#### FE-001 — Vite + React + Tailwind scaffold

**Team:** Frontend — FE-Scaffold Agent
**Title:** Initialize giftsense-frontend with Vite, React 18, Tailwind CSS, and full directory structure
**Scope:** `giftsense-frontend/` (all config files + empty component stubs)
**Dependencies:** none

**Files to create:**
- `package.json`: deps include `react`, `react-dom`, `vite`, `@vitejs/plugin-react`, `tailwindcss`, `postcss`, `autoprefixer`, `lucide-react`
- `vite.config.js`: plugin: react, server.port: 5173
- `tailwind.config.js`: content: `["./src/**/*.{js,jsx}"]`
- `postcss.config.js`: plugins: tailwindcss, autoprefixer
- `index.html`: root div with id `root`, links to `src/main.jsx`
- `src/main.jsx`: `ReactDOM.createRoot(document.getElementById('root')).render(<App />)`
- `src/App.jsx`: stub returning `<div>GiftSense</div>`
- All component files as empty stubs (correct file names, correct `export default function` signature)
- `.env.example`: `VITE_API_URL=http://localhost:8080`

**Acceptance Criteria:**
- `npm install && npm run build` succeeds with no errors
- `npm run dev` starts dev server on port 5173
- No `localStorage` or `sessionStorage` references anywhere in `src/`

**Test Requirement:** Verified by `npm run build` success.

---

#### FE-002 — Session hook + API client

**Team:** Frontend — FE-Scaffold Agent
**Title:** useSession hook (one UUID per window tab) and giftsense.js API client
**Scope:** `src/hooks/useSession.js`, `src/api/giftsense.js`
**Dependencies:** FE-001

**useSession.js:**
```javascript
import { useRef } from 'react'

export function useSession() {
  const sessionId = useRef(crypto.randomUUID())
  return sessionId.current
}
```
- `useRef` ensures UUID never changes within the tab lifecycle
- `crypto.randomUUID()` is called once on mount
- UUID is NOT stored in state (no re-renders), NOT in localStorage

**giftsense.js:**
```javascript
const API_URL = import.meta.env.VITE_API_URL

export async function analyzeConversation({ sessionId, conversationFile, conversationText, recipientDetails }) {
  const formData = new FormData()
  formData.append('session_id', sessionId)
  formData.append('name', recipientDetails.name)
  formData.append('relation', recipientDetails.relation || '')
  formData.append('gender', recipientDetails.gender || '')
  formData.append('occasion', recipientDetails.occasion)
  formData.append('budget_tier', recipientDetails.budgetTier)

  if (conversationFile) {
    formData.append('conversation', conversationFile)
  } else if (conversationText) {
    const blob = new Blob([conversationText], { type: 'text/plain' })
    formData.append('conversation', blob, 'conversation.txt')
  }

  const response = await fetch(`${API_URL}/api/v1/analyze`, {
    method: 'POST',
    body: formData,
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.message || 'Analysis failed')
  }

  return response.json()
}

export async function pingBackend() {
  try {
    await fetch(`${API_URL}/health`)
  } catch (_) {
    // Silent — ping is best-effort cold start mitigation
  }
}
```

**Acceptance Criteria:**
- `useSession` returns a valid UUID v4 format string
- `useSession` called twice in the same component returns the same UUID
- `giftsense.js` reads API URL from `import.meta.env.VITE_API_URL` (no hardcoded URLs)
- `giftsense.js` attaches `session_id` to every request

**Test Requirement:** Verified by code review (no runtime tests for hooks in this phase).

---

#### FE-003 — App.jsx state machine + useAnalyze hook

**Team:** Frontend — FE-Scaffold Agent
**Title:** App.jsx three-screen router (input/loading/results) and useAnalyze state machine hook
**Scope:** `src/App.jsx`, `src/hooks/useAnalyze.js`
**Dependencies:** FE-001, FE-002

**useAnalyze.js:**
```javascript
// State machine: idle | loading | success | error
export function useAnalyze(sessionId) {
  const [status, setStatus] = useState('idle')  // 'idle'|'loading'|'success'|'error'
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const analyze = async (formData) => {
    setStatus('loading')
    setError(null)
    try {
      const data = await analyzeConversation({ sessionId, ...formData })
      setResult(data.data)
      setStatus('success')
    } catch (err) {
      setError(err.message)
      setStatus('error')
    }
  }
  const reset = () => { setStatus('idle'); setResult(null); setError(null) }
  return { status, result, error, analyze, reset }
}
```

**App.jsx:**
- Import `useSession`, `useAnalyze`, and screen components
- Call `pingBackend()` on mount (cold start mitigation)
- Render based on `status`:
  - `idle` or `error` → `<InputScreen onSubmit={analyze} error={error} />`
  - `loading` → `<LoadingScreen />`
  - `success` → `<ResultsScreen result={result} onReset={reset} />`

**Acceptance Criteria:**
- `pingBackend()` is called in a `useEffect` on mount
- Screen transitions match state machine rules
- `reset()` returns to InputScreen

**Test Requirement:** Verified by visual inspection during integration phase.

---

### Phase 7 — Frontend Upload + Form

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| FE-004 | Frontend | Upload components: UploadZone + TextPaste | FE-001 | PENDING |
| FE-005 | Frontend | Recipient form + BudgetSelector | FE-001 | PENDING |
| FE-006 | Frontend | InputScreen: assemble upload + form + submit | FE-003, FE-004, FE-005 | PENDING |

> FE-004 and FE-005 can run in **parallel**. FE-006 depends on both.

---

#### FE-004 — Upload components

**Team:** Frontend — FE-Input Agent
**Title:** UploadZone (drag-drop desktop + tap mobile) and TextPaste (collapsible textarea)
**Scope:** `src/components/upload/UploadZone.jsx`, `src/components/upload/TextPaste.jsx`
**Dependencies:** FE-001

**UploadZone.jsx:**
- Hidden `<input type="file" accept=".txt">` triggered by button click
- Drag-drop zone (detected via `onDragEnter`/`onDrop` on a div) — hidden on touch devices (`window.matchMedia('(pointer: coarse)')`)
- Show selected file name and size once chosen
- Client-side validation:
  - File > 2MB: inline error "This file is too large. Maximum size is 2MB."
  - File < 500 bytes: inline warning "This file seems too short."
  - File extension not `.txt`: inline error "Only .txt files are accepted."
- Calls `onFileSelect(file)` prop with the `File` object when valid
- Mobile: full-width tap target (min 44px height), label visible above button
- Tailwind classes: mobile-first (`base` → `sm:` → `md:`)

**TextPaste.jsx:**
- Collapsible `<textarea>` (collapsed by default, expand via "Paste text instead" link)
- `autocapitalize="off"` attribute to prevent iOS auto-capitalize
- `spellCheck={false}` to avoid red underlines on chat text
- Calls `onTextChange(text)` prop on change
- Max height with scroll on mobile to avoid layout push

**Acceptance Criteria:**
- No `localStorage` or `sessionStorage` usage
- File size validation happens before any API call
- Touch devices show no drag-drop zone
- All interactive elements are ≥ 44×44px

**Test Requirement:** Visual inspection + code review.

---

#### FE-005 — Recipient form + BudgetSelector

**Team:** Frontend — FE-Input Agent
**Title:** RecipientForm (name, relation, gender, occasion) and BudgetSelector (4-tier card grid)
**Scope:** `src/components/form/RecipientForm.jsx`, `src/components/form/BudgetSelector.jsx`
**Dependencies:** FE-001

**RecipientForm.jsx:**
- Fields: Name (required), Relation (optional, e.g. "best friend", "mom"), Gender (optional, dropdown: Male/Female/Other/Prefer not to say), Occasion (required, dropdown: Birthday/Anniversary/Farewell/Wedding/Festival/Other)
- All inputs have visible `<label>` elements (not just placeholders)
- Error messages linked via `aria-describedby`
- Min font size 16px on inputs (prevents iOS auto-zoom)
- Calls `onChange({ name, relation, gender, occasion })` on any field change

**BudgetSelector.jsx:**
- Four selectable cards: Budget (₹500–₹1000), Mid-Range (₹1000–₹5000), Premium (₹5000–₹15000), Luxury (₹15000+)
- Mobile layout: full-width stacked cards (single column)
- Tablet+ layout: 2×2 grid (`sm:grid-cols-2`)
- Desktop layout: 4-column row (`lg:grid-cols-4`)
- Selected card: highlighted border + checkmark icon (Lucide `Check`)
- Color is not the sole selection indicator (checkmark + border width change)
- Calls `onSelect(tier)` prop with tier string: `"BUDGET"`, `"MID_RANGE"`, `"PREMIUM"`, `"LUXURY"`

**Acceptance Criteria:**
- All form labels are visible and associated with their inputs
- BudgetSelector shows checkmark on selected tier
- BudgetSelector uses exact tier string values matching backend enum
- Responsive layout works on 375px width

**Test Requirement:** Visual inspection + code review.

---

#### FE-006 — InputScreen

**Team:** Frontend — FE-Input Agent
**Title:** Assemble InputScreen with upload, form, budget selector, privacy notice, and submit button
**Scope:** `src/screens/InputScreen.jsx`
**Dependencies:** FE-003, FE-004, FE-005

**InputScreen.jsx:**
- Renders in order (mobile single-column):
  1. GiftSense heading + tagline
  2. `<PrivacyNotice variant="pre-upload" />` — "Your conversation is processed privately in this tab and never stored."
  3. `<UploadZone onFileSelect={...} />` and `<TextPaste onTextChange={...} />`
  4. `<RecipientForm onChange={...} />`
  5. `<BudgetSelector onSelect={...} />`
  6. Submit button (disabled until: file or text provided + name + occasion + budget selected)
  7. Error message if `props.error` is set
- On submit: calls `props.onSubmit({ conversationFile, conversationText, name, relation, gender, occasion, budgetTier })`
- File content is kept in React state only — never stored in localStorage

**Acceptance Criteria:**
- Submit button is disabled until all required fields are filled
- Error prop is displayed inline (not a modal)
- Privacy notice is visible above the upload zone

**Test Requirement:** Visual inspection during integration phase.

---

### Phase 8 — Frontend Results

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| FE-007 | Frontend | InsightCard, GiftCard, ShoppingLinks components | FE-001 | PENDING |
| FE-008 | Frontend | ResultsScreen + shared components (LoadingScreen, ErrorMessage, PrivacyNotice) | FE-003, FE-007 | PENDING |

> FE-007 and FE-008 are sequential.

---

#### FE-007 — Results components

**Team:** Frontend — FE-Results Agent
**Title:** InsightCard, GiftCard, and ShoppingLinks components
**Scope:** `src/components/results/InsightCard.jsx`, `src/components/results/GiftCard.jsx`, `src/components/results/ShoppingLinks.jsx`
**Dependencies:** FE-001

**InsightCard.jsx:**
- Props: `{ insight: string, evidenceSummary: string }`
- Displays: insight text (bold/large) + evidence summary (lighter/smaller)
- Mobile: full-width card, comfortable padding (p-4 minimum)
- Max 3 lines for insight before truncating with "Show more" toggle

**GiftCard.jsx:**
- Props: `{ name, reason, estimatedPriceInr, category, links: { amazon, flipkart, googleShopping } }`
- Displays: gift name (prominent), reason (lighter), price badge (pill-shaped), category badge
- Contains `<ShoppingLinks links={links} />`
- Mobile: full-width card

**ShoppingLinks.jsx:**
- Props: `{ links: { amazon, flipkart, googleShopping } }`
- Three buttons: "Amazon India", "Flipkart", "Google Shopping" (each with platform color + Lucide ExternalLink icon)
- All open in new tab: `target="_blank" rel="noopener noreferrer"`
- Mobile: stacked full-width buttons
- Desktop: horizontal row (`sm:flex-row`)
- Small disclaimer below: "Links open filtered search results. Product availability is not guaranteed."

**Acceptance Criteria:**
- All three buttons are present for each gift card
- Disclaimer text is visible below shopping links
- All links open in new tab with `rel="noopener noreferrer"`

**Test Requirement:** Visual inspection during integration phase.

---

#### FE-008 — ResultsScreen + shared components

**Team:** Frontend — FE-Results Agent
**Title:** ResultsScreen, LoadingScreen, ErrorMessage, PrivacyNotice shared components
**Scope:** `src/screens/ResultsScreen.jsx`, `src/components/shared/LoadingScreen.jsx`, `src/components/shared/ErrorMessage.jsx`, `src/components/shared/PrivacyNotice.jsx`
**Dependencies:** FE-003, FE-007

**ResultsScreen.jsx:**
- Props: `{ result: { personality_insights[], gift_suggestions[] }, onReset: fn }`
- Dismissible post-results `<PrivacyNotice variant="post-results" />` banner at top
- Section: "Personality Insights" — horizontal scroll on mobile (snap scroll), 2-col grid on `md:`
- Section: "Gift Suggestions" — single column on mobile, 2-col grid on `md:`
- "Analyze another conversation" button → calls `props.onReset()`

**LoadingScreen.jsx:**
- Props: none
- Rotating loading text cycle (500ms intervals): "Reading your conversation..." → "Finding personality patterns..." → "Crafting personalized suggestions..." → "Almost there..."
- Centered layout, mobile-friendly (no layout shifts)
- After 8 seconds: show additional text: "Our server is warming up — please wait up to 30 seconds."

**ErrorMessage.jsx:**
- Props: `{ message: string, onRetry: fn }`
- Displays error message + "Try again" button
- Readable on 375px width without truncation

**PrivacyNotice.jsx:**
- Props: `{ variant: 'pre-upload' | 'post-results' }`
- `pre-upload`: "This conversation is processed privately in this tab and never stored." — compact, max 2 lines
- `post-results`: "Your conversation has been permanently deleted from our server." — dismissible (X button)

**Acceptance Criteria:**
- Loading text rotates automatically
- Cold-start message appears after 8 seconds of loading
- Privacy notices match copy specified above
- Results screen is usable on 375px width

**Test Requirement:** Visual inspection during integration phase.

---

### Phase 9 — Integration + E2E

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| DO-001 | DevOps | Backend: .env.example finalize + render.yaml backend service config | BE-017 | PENDING |
| DO-002 | DevOps | Frontend: .env.example + wire VITE_API_URL; CORS smoke test | FE-008, BE-017 | PENDING |

---

#### DO-001 — Backend render.yaml and .env.example

**Team:** DevOps — DO-Config Agent
**Scope:** `giftsense-backend/render.yaml` (or root `render.yaml` — one per service definition), `giftsense-backend/.env.example`

**render.yaml (backend service):**
```yaml
services:
  - type: web
    name: giftsense-backend
    runtime: go
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: OPENAI_API_KEY
        sync: false
      - key: PINECONE_API_KEY
        sync: false
      - key: PINECONE_ENVIRONMENT
        sync: false
      - key: ALLOWED_ORIGINS
        value: https://giftsense-frontend.onrender.com
      - key: CHAT_MODEL
        value: gpt-4o
      - key: EMBEDDING_MODEL
        value: text-embedding-3-small
      - key: EMBEDDING_DIMENSIONS
        value: "1536"
      - key: MAX_TOKENS
        value: "1000"
      - key: TOP_K
        value: "3"
      - key: NUM_RETRIEVAL_QUERIES
        value: "4"
      - key: PINECONE_INDEX_NAME
        value: giftsense
      - key: MAX_FILE_SIZE_BYTES
        value: "2097152"
      - key: MAX_PROCESSED_MESSAGES
        value: "400"
      - key: CHUNK_WINDOW_SIZE
        value: "8"
      - key: CHUNK_OVERLAP_SIZE
        value: "3"
```

**Acceptance Criteria:**
- `render.yaml` defines the backend web service
- All required env vars are present (secrets as `sync: false`)
- `.env.example` documents every variable with a comment

---

#### DO-002 — Frontend integration wiring + CORS validation

**Team:** DevOps — DO-Integration Agent
**Scope:** `giftsense-frontend/.env.example`, `giftsense-frontend/render.yaml` (static site section)

**giftsense-frontend/.env.example:**
```
VITE_API_URL=http://localhost:8080
```

**render.yaml (frontend static site):**
```yaml
  - type: static
    name: giftsense-frontend
    buildCommand: npm install && npm run build
    staticPublishPath: dist
    envVars:
      - key: VITE_API_URL
        value: https://giftsense-backend.onrender.com
```

**Integration smoke test (manual checklist):**
- [ ] Backend starts locally with `.env` file set
- [ ] `GET http://localhost:8080/health` returns `{"status":"ok"}`
- [ ] Frontend dev server starts (`npm run dev`)
- [ ] Frontend sends ping to backend on load (check browser network tab)
- [ ] Submit a sample WhatsApp `.txt` file from frontend → backend logs show session processing → response received
- [ ] Shopping link buttons open correct Amazon/Flipkart/Google URLs
- [ ] CORS: frontend on `localhost:5173` can call backend on `localhost:8080` (check for CORS errors in console)

**Acceptance Criteria:**
- Frontend `.env.example` documents `VITE_API_URL`
- `render.yaml` static site section uses correct `dist` publish path
- Full end-to-end flow works locally before deploying

---

### Phase 10 — Deployment

| ID | Team | Title | Depends On | Status |
|---|---|---|---|---|
| DO-003 | DevOps | Full validation checklist + Render deployment | DO-001, DO-002 | PENDING |

---

#### DO-003 — Full validation and deployment

**Team:** DevOps — DO-Integration Agent
**Title:** Run full validation checklist and confirm Render deployment readiness
**Scope:** Validation runs only — no new files

**Full Validation Checklist:**

**Backend:**
- [ ] `go build ./...` — no compilation errors
- [ ] `go test ./...` — all tests pass
- [ ] `go vet ./...` — no vet warnings
- [ ] `grep -r "os.Getenv" internal/` — must return empty (only config.go uses os.Getenv)
- [ ] `grep -r "github.com" internal/domain internal/usecase` — must return empty (no 3rd-party in domain/usecase)
- [ ] All env vars in `.env.example` match Config struct fields exactly

**Frontend:**
- [ ] `npm run build` — clean build, no errors
- [ ] `grep -r "localStorage\|sessionStorage" src/` — must return empty
- [ ] `grep -r "hardcoded" src/api/` — API URL reads from `import.meta.env.VITE_API_URL` only
- [ ] Bundle size check — warn if gzipped output > 500KB

**Integration:**
- [ ] `GET /health` returns 200
- [ ] `POST /api/v1/analyze` with sample payload returns valid response shape
- [ ] CORS headers present for configured frontend origin
- [ ] File upload > 2MB returns 413
- [ ] Missing `OPENAI_API_KEY` causes clean startup failure with clear error message
- [ ] Session cleanup: `DeleteSession` is called after each request (verify in logs)
- [ ] Response JSON matches API Contract shape in CLAUDE.md exactly

**Acceptance Criteria:**
- All checklist items pass
- Orchestrator prints "GIFTSENSE VALIDATION COMPLETE" summary

---

## Integration Checkpoints

| Checkpoint | After Phase | Go/No-Go Condition |
|---|---|---|
| **CP-1: Interfaces locked** | Phase 1 | All port interfaces compile; no Go errors in domain package; config loads with test env vars |
| **CP-2: Pipeline unit-tested** | Phase 2 | All usecase tests pass; parser, anonymizer, chunker each have green tests |
| **CP-3: Backend API smoke test** | Phase 5 | `go test ./...` passes; `POST /api/v1/analyze` with real API keys returns valid JSON |
| **CP-4: Frontend visual review** | Phase 8 | All three screens render on 375px mobile; no layout breaks; privacy notice visible |
| **CP-5: Full E2E locally** | Phase 9 | Complete flow works: upload → analyze → results with real shopping links |
| **CP-6: Deployment ready** | Phase 10 | All validation checklist items pass; render.yaml complete |

Human confirmation is required at **CP-3** (after Phase 5 — backend complete) and **CP-4** (after Phase 8 — frontend complete) and **CP-6** (before any Render deployment). All other phases proceed automatically.

---

## Risk Register

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| Pinecone SDK breaking changes | Low | High | Pin SDK version in go.mod; test with MemoryStore first |
| OpenAI API rate limits during testing | Medium | Medium | Use `gpt-4o-mini` in dev; retry logic in adapter |
| Render cold start > 30 seconds | Medium | Medium | Ping on page load + user messaging in LoadingScreen |
| JSON mode GPT-4o schema mismatch | Medium | Medium | Go validation layer filters/rejects bad output; retry prompt |
| WhatsApp format changes | Low | Low | Parser is isolated in usecase/parse.go; easy to update |
| Pinecone namespace not deleted on backend crash | Low | Low | Periodic background sweep; vectors are anonymized anyway |
| Bundle size exceeds 500KB | Low | Medium | Use lucide-react tree-shaken imports; no heavy UI libs |
| `EMBEDDING_DIMENSIONS` mismatch with Pinecone index | Low | High | Fail-fast validation in config.go; document in .env.example |
