# Product Requirements Document — upahaar.ai
**Version:** 1.0  
**Date:** 2026-04-07  
**Status:** Current State (as-built)

---

## 1. Overview

### 1.1 Product Summary

**upahaar.ai** (उपहार — Hindi for "gift") is an AI-powered gift recommendation web application. Users provide context about a gift recipient — either by uploading a WhatsApp conversation export or an audio recording — and the system analyzes the content to surface personality insights and suggest thoughtful, budget-aware gift ideas with direct shopping links.

### 1.2 Problem Statement

Choosing a meaningful gift for someone requires knowing what they care about, what they already have, and what fits their personality. Most people lack the time or analytical lens to extract these signals from their existing conversations. Generic gift guides fail to account for individual personality, preferences, and cultural context.

### 1.3 Solution

upahaar.ai acts as an AI "gift advisor." It reads the signals already present in conversations (text or audio) and surfaces personality traits the user may not have consciously noticed, then maps those traits to gift ideas grounded in the recipient's actual interests — not generic bestseller lists.

### 1.4 Target Users

- People buying gifts for close relationships (family, friends, colleagues)
- Users with access to WhatsApp conversations or voice recordings of the recipient
- Indian market primarily (INR pricing, Indian e-commerce links, Hindi-language STT support)

---

## 2. Goals & Success Metrics

| Goal | Metric |
|------|--------|
| Deliver relevant gift suggestions | ≥ 70% of users mark result as "helpful" |
| Minimize time to first result | Median analysis latency < 30 seconds |
| Support non-English users | Sarvam STT accuracy across Hindi, Tamil, Telugu, Kannada, Malayalam |
| Protect user privacy | Zero conversation data persisted after response |
| Drive purchase intent | ≥ 30% of users click at least one shopping link |

---

## 3. User Flows

### 3.1 Text Conversation Flow

```
1. Open app (upahaar.ai)
2. Select "Text export" tab
3. Upload WhatsApp .txt export (drag-drop or browse)
4. Fill recipient form:
   - Name (required)
   - Relation (optional, e.g., "sister")
   - Gender (optional)
   - Occasion (required, e.g., "birthday")
   - Budget tier (required — one of 4 options)
5. Click "Find gift ideas →"
6. Loading screen (animated, ~15–30s)
7. Results screen:
   - Personality insights (traits + supporting evidence from conversation)
   - Gift suggestions (name, reason, price range, shopping links)
8. Optional: feedback widget (appears after scroll or 8-second delay)
```

### 3.2 Audio Conversation Flow

```
1. Open app → select "Audio file" tab
2. Upload audio file (.mp3, .wav, .ogg, .opus, .m4a — max 5 MB, max 60 seconds)
3. Fill recipient form (same as text flow)
4. Click "Transcribe & find gifts →"
5. Backend transcribes audio (Sarvam STT)
6. Backend classifies transcript type:

   → CONVERSATION or MONOLOGUE:
      Skip to results (same as text flow)

   → SONG:
      Show EmotionCardScreen:
        - Extracted emotions (emoji + name + intensity bar)
        - Lyrics snippet
        - Detected language label
        - "Find the perfect gift with this feeling →" button
      User confirms → Loading → Results

   → UNKNOWN:
      Show TranscriptConfirmScreen:
        - Editable textarea with transcribed text
        - "Continue →" button
      User confirms/edits → Loading → Results
```

### 3.3 Feedback Flow

```
After results load:
- After 8 seconds or sufficient scroll, feedback prompt appears
- "Was this helpful?" → Yes / No
- If Yes: optional purchase intent ("Will you buy one of these?")
- If No: optional issue selection (too expensive, not my style, etc.) + free text
- Submit → Thank you message
```

---

## 4. Features

### 4.1 Text Analysis Pipeline

| Feature | Description |
|---------|-------------|
| WhatsApp parser | Parses timestamp-formatted chat exports; falls back to plain text |
| Smart sampling | If conversation > 400 messages, representative sample selected |
| Anonymization | Named Entity Recognition removes/pseudonymizes real names |
| Sliding window chunking | 8-message windows with 3-message overlap |
| Vector embeddings | OpenAI `text-embedding-3-small` (1536 dims) |
| Pinecone storage | Per-session namespace; deleted immediately after analysis |
| Multi-query retrieval | 4 semantic queries → top-3 Pinecone matches each |
| GPT-4o analysis | JSON-mode completion with budget-constrained prompt |
| Shopping link generation | Amazon, Flipkart, Google Shopping URL templates |

### 4.2 Audio Analysis Pipeline

| Feature | Description |
|---------|-------------|
| Audio upload | .mp3, .wav, .ogg, .opus, .m4a — max 5 MB, max 60 seconds |
| Client-side validation | File size + audio duration checked before upload |
| Sarvam batch STT | Indian-language-optimized transcription (multi-language) |
| Input classification | GPT classifies transcript as CONVERSATION / MONOLOGUE / SONG / UNKNOWN |
| Song emotion extraction | 5 emotions with emoji + intensity extracted from song lyrics |
| Lyrics snippet | Representative lyric phrase shown to user |
| Language detection | Language code and label returned (e.g., "Hindi") |
| Transcript confirmation | UNKNOWN type shows editable transcript for user review |
| Emotion confirmation | SONG type shows emotion cards for user confirmation |

### 4.3 Gift Recommendations

| Feature | Description |
|---------|-------------|
| Personality insights | 2–4 trait descriptions with supporting evidence quotes |
| Gift suggestions | 3–5 ideas with name, reason, estimated price, category |
| Budget filtering | All suggestions validated against selected budget tier |
| Shopping links | Deep links with INR price filters to Amazon IN, Flipkart, Google Shopping |
| "Start over" action | Resets app state, returns to input screen |

### 4.4 Feedback & Analytics

| Feature | Description |
|---------|-------------|
| Satisfaction rating | "Helpful" / "Not helpful" post-results prompt |
| Purchase intent | Conditional follow-up (positive path) |
| Issue capture | Issue tag selection + free text (negative path) |
| Event tracking | Shopping link clicks tracked as analytics events |
| PostgreSQL storage | Optional DB backend for feedback and events |
| Rate limiting | 5 requests/minute on `/analyze` endpoint (DB-backed) |

### 4.5 Privacy & Security

| Feature | Description |
|---------|-------------|
| No persistence | Conversation/audio never stored; Pinecone namespace deleted post-analysis |
| Session isolation | `crypto.randomUUID()` per browser tab; never reused |
| Privacy notices | Pre-upload and post-upload notices |
| No localStorage | Session state in React memory only |
| CORS | Configurable allowed origins; backend enforces |
| File size enforcement | 2 MB for text; 5 MB for audio (server + client) |

---

## 5. Budget Tiers

| Tier | Display Label | Min (INR) | Max (INR) |
|------|--------------|-----------|-----------|
| `BUDGET` | Budget | ₹500 | ₹1,000 |
| `MID_RANGE` | Mid-range | ₹1,000 | ₹5,000 |
| `PREMIUM` | Premium | ₹5,000 | ₹15,000 |
| `LUXURY` | Luxury | ₹15,000 | — |

---

## 6. API Contract

### 6.1 POST /api/v1/analyze
**Content-Type:** multipart/form-data

**Request fields:**
```
session_id     string   UUID (required)
conversation   file     .txt only, max 2 MB (required)
name           string   Recipient name (required)
relation       string   Relation to recipient (optional)
gender         string   Recipient gender (optional)
occasion       string   Gift occasion (required)
budget_tier    string   BUDGET|MID_RANGE|PREMIUM|LUXURY (required)
```

**Response 200:**
```json
{
  "data": {
    "personality_insights": [
      { "insight": "...", "evidence_summary": "..." }
    ],
    "gift_suggestions": [
      {
        "name": "...",
        "reason": "...",
        "estimated_price_inr": "₹1,000–₹2,000",
        "category": "...",
        "links": {
          "amazon": "https://...",
          "flipkart": "https://...",
          "google_shopping": "https://..."
        }
      }
    ]
  },
  "message": "Analysis complete"
}
```

---

### 6.2 POST /api/v1/analyze-audio
**Content-Type:** multipart/form-data

**Request fields:**
```
session_id     string   UUID (required)
audio          file     .mp3/.wav/.ogg/.opus/.m4a, max 5 MB (required)
name           string   (required)
relation       string   (optional)
gender         string   (optional)
occasion       string   (required)
budget_tier    string   BUDGET|MID_RANGE|PREMIUM|LUXURY (required)
```

**Response 200 — CONVERSATION or MONOLOGUE:**
```json
{
  "data": { /* AnalysisResult */ },
  "audio_analysis": { "input_type": "CONVERSATION", "transcript": "...", "language_code": "en", "language_label": "English" },
  "message": "Analysis complete"
}
```

**Response 200 — SONG:**
```json
{
  "audio_analysis": {
    "input_type": "SONG",
    "emotions": [
      { "name": "Deep Love", "emoji": "❤️", "intensity": 0.9 }
    ],
    "lyrics_snippet": "...",
    "language_code": "hi",
    "language_label": "Hindi"
  },
  "message": "Song detected — confirm emotions to continue"
}
```

**Response 200 — UNKNOWN:**
```json
{
  "audio_analysis": {
    "input_type": "UNKNOWN",
    "transcript": "..."
  },
  "message": "Please review the transcript"
}
```

---

### 6.3 POST /api/v1/analyze-from-transcript
**Content-Type:** application/json

**Request:**
```json
{
  "session_id": "...",
  "transcript": "...",
  "name": "...",
  "relation": "...",
  "gender": "...",
  "occasion": "...",
  "budget_tier": "MID_RANGE",
  "confirmed_emotions": [
    { "name": "Joy", "emoji": "😊", "intensity": 0.8 }
  ]
}
```

**Response:** Same as `/api/v1/analyze` (200 with AnalysisResult)

---

### 6.4 POST /api/v1/feedback
**Content-Type:** application/json

```json
{
  "session_id": "...",
  "satisfaction": "helpful",
  "purchase_intent": "definitely",
  "issues": [],
  "free_text": "",
  "budget_tier": "MID_RANGE",
  "suggestion_count": 4
}
```

---

### 6.5 POST /api/v1/events
**Content-Type:** application/json (fire-and-forget)

```json
{
  "session_id": "...",
  "event_type": "link_click",
  "target": "amazon",
  "metadata": {}
}
```

---

### 6.6 GET /health
Returns HTTP 200. Used by Render health checks.

---

## 7. Technical Architecture

### 7.1 Backend

| Layer | Package | Responsibility |
|-------|---------|---------------|
| Domain | `internal/domain/` | Business entities, no external dependencies |
| Ports | `internal/port/` | Interface definitions (Embedder, LLMClient, VectorStore, Transcriber) |
| Use Cases | `internal/usecase/` | Orchestration of domain logic |
| Adapters | `internal/adapter/` | OpenAI, Pinecone, Sarvam, PostgreSQL implementations |
| Delivery | `internal/delivery/http/` | Gin HTTP handlers, middleware, DTOs |
| Config | `config/` | Environment variable loading (single point of `os.Getenv`) |

**Pattern:** Clean Architecture + Ports & Adapters. No third-party imports in `domain/` or `usecase/`. All external calls go through interfaces.

### 7.2 Frontend

| Layer | Location | Responsibility |
|-------|----------|---------------|
| Screens | `src/screens/` | Top-level views (Input, Loading, Results, EmotionCard, TranscriptConfirm) |
| Components | `src/components/` | Reusable UI (cards, forms, upload zones, feedback) |
| Hooks | `src/hooks/` | State machines (useAnalyze, useAudioAnalyze, useFeedback, useSession) |
| API | `src/api/upahaar.js` | Fetch wrapper; attaches session_id to all requests |

**Pattern:** Functional React + hook-based state machines. No external state library.

### 7.3 Infrastructure

| Service | Provider | Purpose |
|---------|----------|---------|
| Backend API | Render.com (Web Service) | Go binary serving HTTP |
| Frontend | Render.com (Static Site) | Vite-built React SPA |
| Vector DB | Pinecone (Serverless) | Per-session embedding storage |
| LLM + Embeddings | OpenAI API | GPT-4o + text-embedding-3-small |
| STT | Sarvam API | Batch audio transcription |
| Database | Neon PostgreSQL (optional) | Feedback, events, rate limiting |
| Analytics | Vercel Analytics | Page views + frontend events |

---

## 8. Environment Variables

### Backend

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | YES | — | OpenAI API key |
| `PINECONE_API_KEY` | YES | — | Pinecone API key |
| `PINECONE_ENVIRONMENT` | YES | — | Pinecone environment (e.g., `us-east-1`) |
| `SARVAM_API_KEY` | NO | — | Sarvam STT key; audio disabled if absent |
| `DATABASE_URL` | NO | — | PostgreSQL URL; feedback/rate limiting disabled if absent |
| `CHAT_MODEL` | NO | `gpt-4o` | OpenAI chat model |
| `EMBEDDING_MODEL` | NO | `text-embedding-3-small` | OpenAI embedding model |
| `EMBEDDING_DIMENSIONS` | NO | `1536` | Embedding vector dimensions |
| `MAX_TOKENS` | NO | `1000` | Max LLM response tokens |
| `TOP_K` | NO | `3` | Top K Pinecone results per query |
| `NUM_RETRIEVAL_QUERIES` | NO | `4` | Number of semantic retrieval queries |
| `PINECONE_INDEX_NAME` | NO | `upahaar` | Pinecone index name |
| `MAX_FILE_SIZE_BYTES` | NO | `2097152` | Text upload size limit (bytes) |
| `AUDIO_MAX_FILE_SIZE_BYTES` | NO | `5242880` | Audio upload size limit (bytes) |
| `MAX_PROCESSED_MESSAGES` | NO | `400` | Message sampling limit |
| `CHUNK_WINDOW_SIZE` | NO | `8` | Chunking window (messages) |
| `CHUNK_OVERLAP_SIZE` | NO | `3` | Chunk overlap (messages) |
| `PORT` | NO | `8080` | HTTP server port |
| `ALLOWED_ORIGINS` | NO | `http://localhost:5173` | CORS origins (comma-separated) |
| `RATE_LIMIT_PER_MINUTE` | NO | `5` | Rate limit for /analyze |

### Frontend

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_URL` | YES | `http://localhost:8080` | Backend base URL |

---

## 9. Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Latency | Text analysis: < 30 seconds end-to-end (P90) |
| Audio latency | Transcription + analysis: < 60 seconds (P90) |
| File limits | Text: 2 MB; Audio: 5 MB, 60 seconds max |
| Privacy | No conversation data stored beyond request lifetime |
| Rate limiting | 5 analyze requests/minute per IP (when DB configured) |
| Mobile support | Fully responsive; tap-to-upload on mobile |
| Graceful degradation | Audio disabled if `SARVAM_API_KEY` absent; feedback/rate limiting disabled if `DATABASE_URL` absent |
| Session isolation | Each browser tab gets an independent UUID; no cross-tab data sharing |

---

## 10. Screens Reference

| Screen | Component File | Trigger |
|--------|---------------|---------|
| Input | `InputScreen.jsx` | App load |
| Loading | `LoadingScreen.jsx` | Form submit |
| Results | `ResultsScreen.jsx` | Analysis complete |
| Emotion Card | `EmotionCardScreen.jsx` | Audio SONG classification |
| Transcript Confirm | `TranscriptConfirmScreen.jsx` | Audio UNKNOWN classification |

---

## 11. Out of Scope

The following are explicitly not supported in the current version:

- User accounts / authentication / login
- Saving or sharing results
- Gift purchasing within the app (all shopping is external links)
- Batch analysis (multiple recipients in one session)
- Conversation input from non-WhatsApp sources (iMessage, Telegram exports)
- Audio files longer than 60 seconds
- Real-time audio recording (mic input)
- Multi-language UI (English only)
- Admin dashboard for feedback analytics

---

## 12. Dependencies & Third-Party Services

| Service | Purpose | Fallback if unavailable |
|---------|---------|------------------------|
| OpenAI API | Embeddings + GPT-4o | Hard failure (required) |
| Pinecone | Vector storage & retrieval | Hard failure (required) |
| Sarvam API | Audio transcription | Audio tab disabled |
| Neon PostgreSQL | Feedback, events, rate limiting | Features silently disabled |
| Vercel Analytics | Frontend usage tracking | Silently absent |
| Amazon IN | Shopping link destination | Static URL (no API call) |
| Flipkart | Shopping link destination | Static URL (no API call) |
| Google Shopping | Shopping link destination | Static URL (no API call) |

---

## 13. Glossary

| Term | Definition |
|------|------------|
| Session ID | UUID generated per browser tab via `crypto.randomUUID()` |
| Budget Tier | One of four fixed price bands (BUDGET, MID_RANGE, PREMIUM, LUXURY) |
| Chunk | A sliding window of 8 messages with 3-message overlap; the unit of vector storage |
| Pinecone Namespace | Per-session vector storage bucket; deleted after analysis |
| AudioInputType | Classification label assigned to transcribed audio: SONG, CONVERSATION, MONOLOGUE, UNKNOWN |
| EmotionSignal | A detected emotion from song lyrics: `{ name, emoji, intensity }` |
| Multi-query retrieval | Generating 4 semantic queries from recipient details and running each against Pinecone |
| Sarvam STT | Sarvam batch Speech-to-Text API; optimized for Indian languages |
