# upahaar.ai — Product Requirements Document

**Version:** 1.0  
**Date:** 2026-04-17  
**Status:** Current Implementation  
**Brand:** upahaar.ai (Sanskrit/Hindi: उपहार — gift)

---

## 1. Product Overview

### 1.1 What It Is

upahaar.ai is an AI-powered, privacy-first gift recommendation engine that analyzes the emotional context of a relationship to suggest deeply personalized gifts. It accepts two input modalities — a WhatsApp chat export or a Spotify song — and returns personality-derived gift suggestions, shopping links, and a custom greeting card.

### 1.2 Core Value Proposition

Generic gift guides fail because they ignore the specific person. upahaar.ai reads the actual language and emotional texture of a relationship (or the song someone associates with the recipient) to surface gifts that feel chosen, not searched.

### 1.3 Target Users

- Anyone who wants to give a meaningful, personalized gift
- Users who have WhatsApp conversation history with the recipient
- Users who associate a specific song with the recipient or occasion
- Primary geography: India (INR pricing, Amazon.in / Flipkart links)

---

## 2. User Flows

### 2.1 Flow A — Chat-Based Analysis

```
Step 1: Choose "Text" path
         └─ Upload WhatsApp .txt export or paste plain-text conversation

Step 2: Enter recipient details
         └─ Name (required), Relation (optional), Gender (optional),
            Occasion (required), Budget Tier (required)

Step 3: Results
         └─ Personalized greeting card (SVG + PDF download)
         └─ Personality insights (3–5 cards with evidence)
         └─ Gift suggestions (3–6 items with Amazon/Flipkart/Google links)
         └─ Feedback widget
```

### 2.2 Flow B — Song-Based Analysis

```
Step 1: Choose "Audio/Song" path
         └─ Search Spotify for a song associated with the recipient

Step 2: Enter recipient details
         └─ Same fields as Flow A

Step 3: Review detected emotions
         └─ LLM-extracted emotions from song + audio features
         └─ User can confirm or adjust

Step 4: Results
         └─ Same as Flow A results + song context header
```

### 2.3 Session Lifecycle

- Each browser window generates a new `crypto.randomUUID()` session ID on mount.
- All API calls carry this session ID.
- Backend maps session ID to a private Pinecone namespace (Flow A only).
- Pinecone namespace is deleted immediately after the response is assembled.
- No conversation text, session IDs, or personal data are logged or persisted.

---

## 3. Functional Requirements

### 3.1 Input Handling

| Requirement | Detail |
|-------------|--------|
| File upload | WhatsApp `.txt` export, max 2 MB |
| Audio upload | Audio file for transcription, max 5 MB |
| Text detection | Auto-detect WhatsApp format (`[DATE, TIME] Sender:`) vs. plain `Sender: text` |
| Message sampling | Keeps up to 400 messages: 25% most recent + evenly spaced older messages |
| Client-side validation | File size checked before upload; shows inline error if over limit |
| Spotify search | Free-text search → Spotify track results with album art, artist, preview |
| Song preview | 30-second Spotify preview clip playable in browser |

### 3.2 Recipient Form

| Field | Type | Required | Options |
|-------|------|----------|---------|
| Name | Text | Yes | Free text |
| Relation | Select | No | Friend, Partner, Parent, Sibling, Colleague, Other |
| Gender | Select | No | Male, Female, Non-binary, Prefer not to say |
| Occasion | Select | Yes | Birthday, Anniversary, Mother's Day, Friendship Day, Wedding, Graduation, Congratulations, Thank You, Just Because |
| Budget Tier | Button group | Yes | BUDGET, MID_RANGE, PREMIUM, LUXURY |

**Budget Tier Ranges (INR):**

| Tier | Min | Max |
|------|-----|-----|
| BUDGET | ₹500 | ₹1,000 |
| MID_RANGE | ₹1,000 | ₹5,000 |
| PREMIUM | ₹5,000 | ₹15,000 |
| LUXURY | ₹15,000 | — |

### 3.3 AI Analysis Pipeline (Flow A)

1. **Parse** — detect format, filter system messages, sample up to 400 messages
2. **Anonymize** — regex-strip names, phone numbers, emails from all messages
3. **Chunk** — sliding window (size=8, overlap=3), assign metadata (topics, emotional markers, preference/wish flags)
4. **Embed** — OpenAI `text-embedding-3-small` (1536 dims) for all chunks
5. **Upsert** — chunks + vectors pushed to Pinecone namespace `session_id`
6. **Multi-query retrieval** — 4 queries (hobbies, personality, wishes, relationship dynamics) embedded and queried (top-K=3 each)
7. **Rerank** — deduplicate; rank chunks by preference/wish signals, then frequency across queries
8. **Emotion extraction** — LLM extracts up to 5 emotions (constrained to 15-word vocabulary) from sampled chunks
9. **Insights + gifts** — GPT-4o generates personality insights (3–5) and gift suggestions (3–6) in a single prompt, JSON-mode
10. **Budget validation** — gift prices validated against selected tier range
11. **Card generation** — LLM writes card message; SVG + PDF rendered
12. **Cleanup** — Pinecone namespace deleted
13. **Respond** — assembled JSON response returned to client

### 3.4 AI Analysis Pipeline (Flow B — Song)

1. **Song selection** — Spotify search + track selection
2. **Emotion analysis** — fetch audio features (valence, energy, danceability, tempo); LLM extracts emotions (cached by track ID)
3. **User confirmation** — UI shows detected emotions; user can confirm
4. **Direct LLM generation** — insights + gifts generated from song context + emotions alone (no RAG)
5. **Card generation** — same as Flow A
6. **Respond** — assembled response returned

### 3.5 Greeting Card

| Attribute | Detail |
|-----------|--------|
| Format | SVG (inline in response) + PDF (base64 in response) |
| Fonts | Playfair Display (headings), Inter (body) |
| Occasions | Birthday, Mother's Day, Anniversary, Friendship, Wedding, Graduation, Congratulations, Thank You, Generic |
| Motifs | Confetti (birthday), Floral (mothers), Rings (anniversary), Stars (friendship), Sunburst (generic) |
| Palettes | Occasion-specific color themes (background, primary, accent, ink, muted) |
| Content | Headline, body paragraph, closing, signature, recipient name |
| Download | PDF download triggered client-side from base64 blob |

### 3.6 Shopping Links

All links are generated with budget filters applied:

| Retailer | URL Pattern |
|----------|-------------|
| Amazon India | `amazon.in/s?k={name}&rh=p_36%3A{min_paise}-{max_paise}` |
| Flipkart | `flipkart.com/search?q={name}&price_range.from={min}&to={max}` |
| Google Shopping | `google.com/search?q={name}+under+₹{max}&tbm=shop` |

### 3.7 Feedback System

- **Satisfaction**: helpful / not helpful (binary)
- **Purchase intent**: Yes / Maybe / No
- **Issues**: Multi-select (irrelevant suggestions, wrong budget, poor insights, other)
- **Free text**: Optional comments
- **Rate limiting**: 5 feedback submissions per session per minute
- **Storage**: PostgreSQL via GORM

### 3.8 Analytics Events

Events tracked (fire-and-forget via `sendBeacon`):

- Shopping link clicks (platform, gift name, session)
- Feedback submissions
- Path selection (text vs. audio)

---

## 4. Non-Functional Requirements

### 4.1 Privacy

| Requirement | Implementation |
|-------------|---------------|
| No conversation storage | Text never written to disk or database |
| No logging of personal content | Conversation text excluded from all log levels |
| Session isolation | Pinecone namespace = session ID; deleted post-response |
| Anonymization before embedding | Names, phones, emails stripped before chunking |
| Client-side session ID | `crypto.randomUUID()` per window; never persisted to localStorage |
| Privacy notice | Displayed on results screen: "Your data was not stored" |

### 4.2 Performance

| Requirement | Target |
|-------------|--------|
| Analysis latency (P50) | < 15 seconds end-to-end |
| Max conversation file size | 2 MB |
| Max audio file size | 5 MB |
| Pinecone top-K per query | 3 |
| Parallel LLM calls | Insights and gifts generated concurrently |
| Song emotion cache | Cached by Spotify track ID to avoid re-analysis |

### 4.3 Rate Limiting

| Scope | Limit |
|-------|-------|
| Analysis requests | 5 per session per minute (configurable) |
| Feedback submissions | Per-session enforcement |
| Implementation | DB-backed token bucket per session ID |

### 4.4 Security

- CORS: Validated against `ALLOWED_ORIGINS` environment variable
- File type enforcement: Only `.txt` accepted for conversation upload
- File size enforcement: Server-side limit (Gin middleware) + client-side pre-check
- No SQL injection surface: GORM parameterized queries only
- No XSS surface: SVG rendered from server-controlled template; card content sanitized by Go template

---

## 5. API Contract

### 5.1 Analyze Conversation

```
POST /api/v1/analyze
Content-Type: multipart/form-data

Fields:
  session_id    string  UUID (required)
  conversation  file    .txt, max 2MB (required)
  name          string  (required)
  relation      string  (optional)
  gender        string  (optional)
  occasion      string  (required)
  budget_tier   string  BUDGET | MID_RANGE | PREMIUM | LUXURY (required)

Response 200:
{
  "data": {
    "personality_insights": [
      { "insight": "string", "evidence_summary": "string" }
    ],
    "gift_suggestions": [
      {
        "name": "string",
        "reason": "string",
        "estimated_price_inr": "₹1000-₹2000",
        "category": "string",
        "links": {
          "amazon": "https://...",
          "flipkart": "https://...",
          "google_shopping": "https://..."
        }
      }
    ],
    "card": {
      "svg": "string",
      "pdf_base64": "string",
      "theme_id": "string",
      "content": { ... }
    }
  },
  "message": "Analysis complete"
}
```

### 5.2 Analyze from Song

```
POST /api/v1/analyze-from-song
Content-Type: application/json

{
  "session_id": "uuid",
  "track": { "id": "...", "name": "...", "artist": "..." },
  "emotions": [{ "name": "Joy", "emoji": "😊", "intensity": 0.9 }],
  "name": "string",
  "relation": "string",
  "occasion": "string",
  "budget_tier": "string"
}
```

### 5.3 Spotify Search

```
GET /api/v1/spotify/search?q={query}

Response 200:
{
  "tracks": [
    {
      "id": "string",
      "name": "string",
      "artist": "string",
      "album_art": "https://...",
      "preview_url": "https://..."
    }
  ]
}
```

### 5.4 Feedback

```
POST /api/v1/feedback
Content-Type: application/json

{
  "session_id": "uuid",
  "satisfaction": "helpful" | "not_helpful",
  "purchase_intent": "yes" | "maybe" | "no",
  "issues": ["irrelevant_suggestions"],
  "free_text": "string"
}
```

### 5.5 Error Responses

```json
{
  "error": "error_code",
  "message": "Human-readable message",
  "details": { }
}
```

Common error codes: `file_too_large`, `invalid_budget_tier`, `llm_error`, `retrieval_failed`, `rate_limit_exceeded`

### 5.6 Health Check

```
GET /health → 200 OK
```

---

## 6. Frontend Architecture

### 6.1 State Machine

The app is a multi-step wizard driven by a single React context (`StepperContext`):

| State Key | Type | Purpose |
|-----------|------|---------|
| `sessionId` | UUID string | Immutable per window mount |
| `currentPath` | `'text' \| 'audio' \| null` | Input modality selected |
| `currentStep` | number | 0-indexed step within path |
| `formData` | object | All user inputs accumulated across steps |
| `analysisState` | `'idle' \| 'loading' \| 'success' \| 'error'` | API call lifecycle |
| `result` | object | Full API response |
| `audioAnalysis` | object | Transcription + emotion classification |
| `songEmotions` | array | Detected emotions from Spotify song |
| `error` | string | Error message to display |

### 6.2 Step Routing

| Path | Step 0 | Step 1 | Step 2 | Step 3 |
|------|--------|--------|--------|--------|
| Text | InputStep | RecipientStep | ResultsStep | — |
| Audio | InputStep | RecipientStep | EmotionStep | ResultsStep |

### 6.3 Component Tree

```
App
└── StepperContext (state, dispatch)
    └── ProgressBar
    └── StepTransition (animated)
        ├── InputStep
        │   ├── UploadZone (drag-drop .txt)
        │   └── SpotifySongPicker (search + preview)
        ├── RecipientStep
        │   ├── RecipientForm
        │   └── BudgetSelector
        ├── EmotionStep (audio path only)
        │   └── SongEmotionSummary
        └── ResultsStep
            ├── CardHero (SVG render)
            ├── CardActions (PDF download, Start Over)
            ├── InsightCard × N
            ├── GiftCard × N
            │   └── ShoppingLinks
            └── FeedbackWidget
```

### 6.4 Design Principles

- **Mobile-first**: Tailwind base → `sm:` → `md:` → `lg:` ordering
- **No heavy UI library**: Tailwind + Lucide React only
- **No persistence**: Zero `localStorage` / `sessionStorage` usage
- **Privacy by design**: Session UUID in memory only; cleared on window close

---

## 7. Backend Architecture

### 7.1 Clean Architecture Layers

```
delivery/http        → Gin handlers, DTOs, middleware
usecase/             → Business logic (no framework imports)
port/                → Go interfaces (Embedder, LLMClient, VectorStore, etc.)
adapter/             → Concrete implementations (OpenAI, Pinecone, Spotify, Sarvam)
domain/              → Pure types and sentinel errors
```

### 7.2 Key Interfaces

```go
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type LLMClient interface {
    Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, error)
}

type VectorStore interface {
    Upsert(ctx context.Context, sessionID string, chunks []domain.Chunk, vectors [][]float32) error
    Query(ctx context.Context, sessionID string, queryVector []float32, topK int, filter MetadataFilter) ([]domain.Chunk, error)
    DeleteSession(ctx context.Context, sessionID string) error
}

type SpotifyClient interface { ... }
type Transcriber interface { ... }       // Sarvam AI
type FeedbackStore interface { ... }
type SongEmotionCache interface { ... }
type RateLimiter interface { ... }
```

### 7.3 Environment Variables

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `OPENAI_API_KEY` | Yes | — | OpenAI API authentication |
| `PINECONE_API_KEY` | Yes | — | Pinecone authentication |
| `PINECONE_ENVIRONMENT` | Yes | — | Pinecone environment region |
| `PINECONE_INDEX_NAME` | No | upahaar | Pinecone index name |
| `CHAT_MODEL` | No | gpt-4o | LLM model for generation |
| `EMBEDDING_MODEL` | No | text-embedding-3-small | Embedding model |
| `EMBEDDING_DIMENSIONS` | No | 1536 | Embedding vector dimensions |
| `MAX_TOKENS` | No | 1000 | LLM max output tokens |
| `TOP_K` | No | 3 | Pinecone results per query |
| `NUM_RETRIEVAL_QUERIES` | No | 4 | Multi-query count |
| `MAX_FILE_SIZE_BYTES` | No | 2097152 | Max conversation upload size |
| `AUDIO_MAX_FILE_SIZE_BYTES` | No | 5242880 | Max audio upload size |
| `MAX_PROCESSED_MESSAGES` | No | 400 | Message sampling cap |
| `CHUNK_WINDOW_SIZE` | No | 8 | Sliding window size |
| `CHUNK_OVERLAP_SIZE` | No | 3 | Chunk overlap |
| `PORT` | No | 8080 | HTTP listen port |
| `ALLOWED_ORIGINS` | No | localhost:5173 | CORS allowed origins (comma-separated) |
| `RATE_LIMIT_PER_MINUTE` | No | 5 | Requests per session per minute |
| `DATABASE_URL` | No | — | PostgreSQL (enables feedback + caching) |
| `SARVAM_API_KEY` | No | — | Audio transcription (Sarvam AI) |
| `SPOTIFY_CLIENT_ID` | No | — | Spotify OAuth client ID |
| `SPOTIFY_CLIENT_SECRET` | No | — | Spotify OAuth client secret |

---

## 8. External Dependencies

| Service | Purpose | Required |
|---------|---------|---------|
| OpenAI API | Embedding + chat generation (GPT-4o) | Yes |
| Pinecone | Session-namespaced vector storage | Yes |
| Spotify Web API | Track search, audio features | No (disables audio path) |
| Sarvam AI | Hindi/Indic audio transcription | No (disables audio upload) |
| PostgreSQL | Feedback storage, song emotion cache, rate limiting | No (fallback: in-memory) |

**Frontend:**

| Package | Version | Purpose |
|---------|---------|---------|
| React | 19.2.4 | UI framework |
| Vite | 8.0.0 | Build tool |
| Tailwind CSS | 3.4.19 | Styling |
| Lucide React | 0.577.0 | Icons |
| @vercel/analytics | 2.0.1 | Page analytics |
| @vercel/speed-insights | 2.0.0 | Performance monitoring |

**Backend:**

| Package | Purpose |
|---------|---------|
| gin-gonic/gin v1.9.1 | HTTP routing + middleware |
| openai/openai-go v1.12.0 | OpenAI API client |
| pinecone-io/go-pinecone v1.1.1 | Vector DB client |
| signintech/gopdf v0.36.0 | PDF rendering |
| gorm.io/gorm | ORM for PostgreSQL |
| joho/godotenv | .env loading |
| stretchr/testify v1.11.1 | Testing assertions |

---

## 9. Deployment

### 9.1 Docker Compose (Local / Self-Hosted)

```
docker-compose up
  → backend:  port 8080 (Go binary, multi-stage build)
  → frontend: port 3000 (nginx serving Vite build)
```

nginx routes `/api/` requests to `backend:8080`. All other routes serve `index.html` (SPA routing).

### 9.2 Cloud (Render)

- **Backend**: Render Web Service — Go binary, auto-detects `PORT`
- **Frontend**: Render Static Site — `npm run build`, `dist/` directory
- **Database**: Render PostgreSQL (optional)

### 9.3 Font Assets

Fonts (`Playfair_Display/`, `Inter/`) must be present in the repository root or mounted as a volume for PDF generation to function. The Go PDF renderer loads TTF files at startup.

---

## 10. Emotion Vocabulary

The following 15 emotions form the constrained vocabulary used by the LLM for both chat emotion extraction and song emotion analysis. This ensures consistent, displayable emotion labels:

> Deep Love, Heartbreak, Longing, Joy, Nostalgia, Passion, Melancholy, Hope, Devotion, Playfulness, Yearning, Tenderness, Celebration, Grief, Warmth

---

## 11. Greeting Card Occasion Matrix

| Occasion Key | Greeting | Motif | Palette Theme |
|-------------|---------|-------|--------------|
| `birthday` | Happy Birthday | Confetti | Warm orange/gold |
| `mothers-day` | Happy Mother's Day | Floral | Soft pink/green |
| `anniversary` | Happy Anniversary | Rings | Deep rose/purple |
| `friendship` | Happy Friendship Day | Stars | Sky blue/yellow |
| `wedding` | Congratulations | Rings | Cream/gold |
| `graduation` | Congratulations | Sunburst | Navy/gold |
| `congratulations` | Congratulations | Sunburst | Green/gold |
| `thank-you` | Thank You | Stars | Teal/warm |
| `just-because` | Thinking of You | Sunburst | Lavender/warm |

---

## 12. Out of Scope (Current Implementation)

The following are **not** implemented and are explicitly excluded from the current scope:

- User accounts or authentication
- Saved gift lists or history
- Email or push notifications
- Direct e-commerce checkout or affiliate integration
- Multi-language UI (English only)
- WhatsApp cloud API integration (file upload only, no live chat)
- Image or video input analysis
- Collaborative gift pooling
- Price comparison or real-time stock availability

---

## 13. Known Constraints

| Constraint | Detail |
|-----------|--------|
| Pinecone cold start | First request after idle period may take 3–5 extra seconds |
| GPT-4o latency | End-to-end P95 can exceed 20s for large conversations |
| Spotify preview availability | Not all tracks have 30-second preview URLs |
| PDF font dependency | PDF generation fails if font files are absent from filesystem |
| Song path has no RAG | Song-based gifts rely purely on prompt context, not retrieved chunks |
| Rate limit is session-scoped | Different browser windows bypass rate limiting (by design) |
| Audio transcription language | Sarvam AI is optimized for Hindi/Indic languages; English accuracy may vary |
