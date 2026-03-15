# GiftSense — System Architecture & Learning Guide
### A Teaching-First, Building-Second RAG System Design Document

> **How to use this document:** Each module represents a real architectural layer of the GiftSense system. Read them in order. Every module teaches a foundational concept *and* slots into the larger system. By the end, you will understand not just what to build — but why every decision was made, what was traded away, and how the pieces connect.

---

## Table of Contents

1. [Module 1 — What Is RAG and Why Does It Exist?](#module-1)
2. [Module 2 — System Overview & Data Flow Architecture](#module-2)
3. [Module 3 — Conversation Upload & Input Processing](#module-3)
4. [Module 4 — Anonymization & Privacy-Safe PII Handling](#module-4)
5. [Module 5 — Chunking Strategy for Conversational Data](#module-5)
6. [Module 6 — Embeddings & Semantic Vector Space](#module-6)
7. [Module 7 — Vector Store Design with Pinecone Free Tier](#module-7)
8. [Module 8 — Retrieval Strategy & Query Construction](#module-8)
9. [Module 9 — Prompt Engineering for RAG — Personality Insights & Gift Suggestions](#module-9)
10. [Module 10 — Budget-Aware Shopping Link Generation](#module-10)
11. [Module 11 — Session-Scoped Ephemeral Pipeline (Privacy by Design)](#module-11)
12. [Module 12 — Go Backend Architecture — Clean Architecture with Gin](#module-12)
13. [Module 13 — React Frontend Architecture](#module-13)
14. [Module 14 — OpenAI Cost Model & Optimization Strategy](#module-14)
15. [Module 15 — Render Free Tier Deployment Architecture](#module-15)
16. [Module 16 — Observability, Failure Modes & Production Hardening](#module-16)
17. [Module 17 — Phased Build Plan](#module-17)
18. [Module 18 — Application Configuration via Environment Variables](#module-18)

---

<a name="module-1"></a>
## Module 1 — What Is RAG and Why Does It Exist?

### Objective
Establish a clear mental model of Retrieval-Augmented Generation before touching any implementation. Every architectural decision downstream flows from truly understanding *why* RAG exists as a pattern.

---

### Concepts to Learn
- The knowledge problem in large language models
- Fine-tuning vs. prompt stuffing vs. RAG — the three paradigms
- The core RAG loop: Retrieve → Augment → Generate
- When RAG is the right tool and when it is not

---

### Detailed Explanation

**The Knowledge Problem**

Large language models like GPT-4o are trained on a fixed snapshot of the world. Once trained, their internal knowledge is frozen. They cannot know about your specific conversation with your mom. They cannot read the file you uploaded today. They have no access to private, personal, or real-time data.

This creates a fundamental gap: the model is *capable* of sophisticated reasoning, but it is *blind* to the specific context it needs to reason over.

**Three Paradigms for Closing the Gap**

*Fine-tuning* means retraining or partially retraining the model on your own data so the knowledge is baked into the weights. For GiftSense, this is nonsensical — every user has different conversations, and you cannot retrain a model per session. Fine-tuning is for teaching the model new *skills* or *styles*, not new *facts*.

*Prompt stuffing* (also called "context stuffing" or "naive RAG") means simply pasting the entire source document directly into the LLM prompt. If a user uploads a 200-message WhatsApp conversation, you stuff all 200 messages into the prompt and ask GPT-4o to reason over it. This works for small documents but breaks down for several reasons: (a) token limits are finite and expensive; (b) you expose raw PII directly to OpenAI; (c) the LLM's attention degrades over very long contexts ("lost in the middle" problem); and (d) it gives you no control over *what* the model focuses on.

*RAG — Retrieval-Augmented Generation* is the third paradigm. Instead of stuffing everything in, you build a retrieval layer that identifies only the most relevant pieces of information and passes *those* to the LLM. The pipeline has three distinct phases:

1. **Index time (happens once per session):** The uploaded conversation is processed, cleaned, anonymized, split into chunks, and each chunk is converted into a vector embedding. These embeddings are stored in a vector store.

2. **Retrieve time (happens per query):** When you need to generate gift suggestions, you construct a semantic query (e.g., "what does this person enjoy, what are their hobbies, what emotional patterns exist in this conversation"). This query is embedded into the same vector space. The vector store returns the most semantically similar chunks.

3. **Generate time:** Only the retrieved chunks — not the entire conversation — are injected into the GPT-4o prompt as context. The model generates insights and suggestions grounded in this retrieved evidence.

**Why RAG Is the Right Choice for GiftSense**

- The source data (a conversation) is private, per-session, and never the same twice. Fine-tuning is impossible.
- The conversation can be arbitrarily long. Prompt stuffing would blow the token budget and expose full PII.
- You need selective context injection — only the emotional signals, preference cues, and personality markers matter. RAG lets you extract exactly those.
- You want to keep raw personal content away from OpenAI. RAG lets you anonymize before embedding.

**When RAG Fails**

RAG is not magic. It fails when:
- Chunks are too large (retrieval becomes imprecise) or too small (chunks lack sufficient context)
- The embedding model does not capture the semantic content of the domain well
- The retrieved chunks do not actually contain the answer (retrieval precision failure)
- The LLM hallucinates beyond what the retrieved context supports (generation failure)
- The query used to retrieve context is poorly constructed

Understanding these failure modes is as important as understanding the happy path.

---

### Design Decisions
- RAG is chosen over fine-tuning and prompt stuffing for all the reasons above.
- OpenAI's `text-embedding-3-small` is chosen for embeddings because it is cost-efficient, high-quality, and keeps the embedding provider identical to the generation provider, simplifying the pipeline.

---

### Alternative Approaches
- **Pure prompt stuffing with GPT-4o's 128k context window:** For a short conversation (< 50 messages), this could theoretically work. It is faster to implement but exposes raw PII to OpenAI, is expensive per call, and scales poorly.
- **Fine-tuning a small local model on gift recommendation patterns:** Far too complex for a learning project, requires significant training data, and does not solve the per-session personalization problem.
- **Keyword-based (BM25) retrieval instead of semantic embeddings:** Faster and requires no OpenAI calls at index time, but misses semantic similarity — "she loves adventure" would not retrieve chunks about "hiking" or "travel" unless those exact words appear.

---

### Trade-offs / Consequences

| Approach | Cost | Privacy | Personalization | Complexity |
|---|---|---|---|---|
| RAG (chosen) | Medium | High (with anonymization) | High | Medium |
| Prompt stuffing | High | Low (raw PII to OpenAI) | Medium | Low |
| Fine-tuning | Very High | N/A | Low | Very High |

---

### How This Module Connects to the Overall System
Every subsequent module — chunking, embedding, retrieval, prompt engineering — is an implementation of a specific phase of this RAG loop. Understanding the full loop first makes every downstream decision legible.

---

<a name="module-2"></a>
## Module 2 — System Overview & Data Flow Architecture

### Objective
Map the complete end-to-end data flow of GiftSense — from the moment a user uploads a conversation to the moment they see gift suggestions with shopping links. Understand which systems are involved at each step, what data crosses each boundary, and where OpenAI is invoked.

---

### Concepts to Learn
- System boundary diagrams and data flow design
- Identifying trust boundaries in a data pipeline
- The principle of data minimization
- Stateless vs. stateful service design
- Multi-window session isolation
- External vector store as an architectural boundary
- Application-wide configurability via environment variables

---

### Detailed Explanation

**The Complete Data Flow**

```
[User Browser — Window 1]   [User Browser — Window 2]   [User Browser — Window N]
         │                           │                           │
         │  (1) HTTPS POST /api/analyze                         │
         │      { recipient_details, conversation_text,         │
         │        session_id (UUID generated client-side) }     │
         ▼                           ▼                           ▼
[React Frontend — Render Static Site]
         │
         │  (2) POST /api/analyze  (each window sends its own independent request)
         ▼
[Go Gin Backend — Render Web Service]
         │
         ├──► (3) Config Loader
         │         Reads all runtime parameters from environment:
         │         CHAT_MODEL, EMBEDDING_MODEL, MAX_TOKENS, TOP_K,
         │         MAX_FILE_SIZE_BYTES, PINECONE_API_KEY, etc.
         │
         ├──► (4) Input Validation & Sanitization
         │         File size check: rejects uploads > MAX_FILE_SIZE_BYTES (2MB)
         │         Budget, occasion, relation field validation
         │
         ├──► (5) Conversation Parser
         │         Detects format (WhatsApp export / plain text)
         │         Normalizes into message[]{ sender, text, timestamp }
         │
         ├──► (6) Anonymizer
         │         NER scan → replace names/places/identifiers with tokens
         │         [Person_A], [Person_B], [City_1], etc.
         │         Raw conversation NEVER leaves this stage
         │
         ├──► (7) Chunker
         │         Sliding window chunking on anonymized messages
         │         Metadata enrichment: Topics, EmotionalMarkers, HasPreference, HasWish
         │
         ├──► (8) Embedder  ────────────────────────────────────────────────────┐
         │         Calls OpenAI {EMBEDDING_MODEL} (configurable)                │
         │         Input: anonymized chunk text                                 │
         │         Output: N-dimensional float vector                           │
         │         ◄── OpenAI sees: anonymized text only ──────────────────────►│
         │                                                                       │
         ├──► (9) Pinecone Free Tier (session-scoped namespace)  ◄──────────────┘
         │         Index name: giftsense (shared, single index)
         │         Namespace: {session_id} (one per request, isolates windows)
         │         Upserts: { chunk_id, vector, metadata }
         │         ◄── Pinecone sees: anonymized vectors + metadata only ──────►
         │                                                                       │
         ├──► (10) Query Constructor                                            │
         │          Builds 4 targeted retrieval queries from recipient context  │
         │          Embeds each query via OpenAI {EMBEDDING_MODEL}              │
         │                                                                       │
         ├──► (11) Retriever                                                    │
         │          Queries Pinecone namespace for this session                 │
         │          top_k={TOP_K} (configurable, default 3 per query)          │
         │          Metadata filtering: HasPreference, Topics, etc.            │
         │          Returns top-K most relevant anonymized chunks               │
         │                                                                       │
         ├──► (12) Prompt Builder                                               │
         │          Assembles: system prompt + retrieved chunks + recipient    │
         │          details + budget constraints                                │
         │          ◄── OpenAI sees: anonymized chunks + recipient metadata ──► │
         │                                                                       │
         ├──► (13) GPT-{CHAT_MODEL} Completion Call (configurable model)        │
         │          max_tokens={MAX_TOKENS} (configurable)                      │
         │          Output: structured JSON                                     │
         │          { personality_insights[], gift_suggestions[] }             │
         │                                                                       │
         ├──► (14) Link Generator                                               │
         │          Constructs Amazon India, Flipkart, Google Shopping URLs    │
         │                                                                       │
         ├──► (15) Pinecone Namespace Cleanup                                   │
         │          DELETE namespace: {session_id}                             │
         │          All vectors for this session are permanently removed       │
         │                                                                       │
         └──► (16) Response to Frontend
                   { personality_insights[], gift_suggestions[{ ...links }] }
```

**Multi-Window Session Isolation**

A critical architectural change from the original design: users can open multiple browser windows or tabs simultaneously, each uploading a different conversation for a different recipient. Each window operates as a completely independent session.

Session isolation is achieved through a **client-generated session UUID**. When the React app loads in any window, it generates a `crypto.randomUUID()` and attaches it to every API request as a header or request field. On the backend, this UUID becomes the **Pinecone namespace** for that session's vectors. Since Pinecone namespaces are isolated within a single index, two concurrent windows cannot read or contaminate each other's vectors.

This is the key architectural reason Pinecone is a natural fit for multi-window support: namespace isolation is a first-class Pinecone feature, whereas with an in-memory store you would need to manually manage session-keyed maps with locking.

**Trust Boundaries**

There are now four trust boundaries in this system:

1. **Browser → Backend:** HTTPS only. Each window sends its own session UUID and conversation text. The backend treats each request independently.

2. **Backend → OpenAI API:** Only anonymized content crosses this boundary. Raw names, personal identifiers, and the original conversation text *never* cross this boundary.

3. **Backend → Pinecone:** Only anonymized vectors and metadata cross this boundary. Pinecone receives: vector arrays, chunk metadata (topics, emotional markers, booleans), and the session namespace UUID. It never receives raw conversation text, real names, or personal identifiers.

4. **Backend Cleanup → Pinecone:** After the response is assembled and returned, the backend issues a namespace delete request to Pinecone. This removes all vectors for the completed session.

**What OpenAI Never Sees**
- The original, un-anonymized conversation
- Real names of the user or recipient
- Specific locations, phone numbers, or identifiers
- The user's identity, session ID, or account information

**What OpenAI Does See**
- Anonymized conversation chunks (e.g., "[Person_A] mentioned they want to visit [City_1] someday")
- Recipient metadata: relation, occasion, budget range, gender
- Retrieval query text (constructed from metadata, not raw conversation)
- The assembled prompt with retrieved anonymized context

**What Pinecone Does See**
- Anonymized embedding vectors (floating point numbers — no text recoverable from vectors alone)
- Chunk metadata: topics array, emotional marker booleans, preference flags, chunk index
- The session namespace UUID (a random identifier — not linked to user identity)

---

### Design Decisions
- The vector store is **Pinecone free tier with session-scoped namespaces** — each browser window gets an isolated namespace, and the namespace is deleted after the response is returned to the client.
- The anonymization step is placed **before any OpenAI or Pinecone call** — embeddings, completions, and vector upserts only ever handle anonymized content.
- **Client-side session UUID generation** for multi-window isolation — the client generates the session identity, not the server. This keeps the backend stateless between requests.
- **All runtime parameters are configurable via environment variables** — model names, token limits, Top-K values, file size caps — so no code change is required to tune the system.

---

### Alternative Approaches
- **In-process in-memory store:** Simpler and requires no external dependency. However, it does not support multi-window isolation cleanly (shared Go process memory must be explicitly keyed and locked per session), and Render may serve different windows from different process restarts. Pinecone's namespace model handles this more cleanly.
- **Storing sessions in Redis:** Would allow resumable sessions, but violates the "no data stored" privacy promise. Not appropriate for this use case.
- **Streaming responses:** GPT-4o supports streaming, which would improve perceived latency. Adds implementation complexity — a reasonable Phase 2 enhancement.

---

### Trade-offs / Consequences
- **Pinecone free tier** means anonymized vectors leave your server and transit Pinecone's infrastructure. This is acceptable because: (a) only anonymized content is stored, and (b) the namespace is explicitly deleted after each session. The privacy story is weaker than fully in-memory, but stronger than storing raw conversations.
- **Session namespace lifecycle depends on the cleanup call.** If the Go backend crashes between returning the response and issuing the delete, vectors for that session remain in Pinecone until TTL or manual cleanup. Mitigation: implement a TTL-based cleanup strategy and a periodic background sweep (discussed in Module 7).
- **Multi-window support** introduces the need for session UUID deduplication logic — an extremely unlikely but theoretically possible UUID collision between two simultaneous windows. `crypto.randomUUID()` makes this negligible in practice.

---

### How This Module Connects to the Overall System
This module is the map. Every subsequent module zooms into one segment of this data flow. Whenever you feel lost in a downstream module, return to this diagram and locate where you are.

---

<a name="module-3"></a>
## Module 3 — Conversation Upload & Input Processing

### Objective
Design the input ingestion layer — the first stage of the pipeline. Understand how raw, messy, multi-format conversation data is parsed and normalized into a structured form the rest of the pipeline can work with reliably. This module also covers the 2MB file size allowance and how it shapes validation and cost expectations.

---

### Concepts to Learn
- Input normalization as a pipeline stage
- Format detection and parsing strategies
- Defensive input validation
- Structured representation of conversational data
- File size limits as a system design lever (cost, privacy, performance)
- Multi-window upload: each window is an independent upload context

---

### Detailed Explanation

**The Problem with Raw Conversation Input**

Users will upload conversations in at least three formats:
1. **WhatsApp text export** — a `.txt` file with a specific timestamp-based format: `[DD/MM/YYYY, HH:MM:SS] Name: message text`
2. **Plain text paste** — free-form text pasted into a textarea, potentially with no structure at all
3. **Screenshots** — images of conversations (requires OCR, discussed below)

Each format requires different parsing logic, but the rest of the pipeline should receive the same normalized structure regardless of input format. This is the Parser's job.

**Multi-Window Upload Context**

Each browser window or tab is a completely independent upload context. Window 1 may be uploading a conversation with the user's mom for a birthday gift, while Window 2 is uploading a conversation with a colleague for a farewell gift — simultaneously, without any interference. The backend is stateless between requests; each window's upload triggers a fully independent pipeline execution with its own session UUID and its own Pinecone namespace.

The frontend does not need to be aware of other windows. Each React app instance manages its own state independently. The session UUID is generated at page load and is unique per window.

**Normalized Message Structure**

The parser's output should be a list of message objects with a consistent shape:

```
Message {
    Index:     int          // position in conversation (0-based)
    Sender:    string       // raw sender label (before anonymization)
    Text:      string       // message body
    Timestamp: time.Time    // parsed if available, zero value if not
    IsMedia:   bool         // true if "<Media omitted>" or similar
}
```

**WhatsApp Export Parsing**

WhatsApp exports follow a recognizable pattern. The parser uses a line-by-line scan, detects lines that match the timestamp-sender pattern, and accumulates multi-line messages (since a single message can span multiple lines). System messages ("Messages and calls are end-to-end encrypted") are filtered out. `<Media omitted>` lines set `IsMedia: true` and contribute no text content to embeddings.

The key insight here is that the parser should be tolerant — it should not fail on lines that don't match the expected pattern. Instead, it appends them to the previous message's text as a continuation.

**Plain Text Parsing**

Plain text has no structure guarantees. The parser uses a heuristic approach:
- If the text contains recognizable sender prefixes (e.g., "You:", "Friend:", "Me:"), it attempts to split on those.
- If no structure is detectable, the entire text is treated as a single large document and chunked directly without per-message attribution.

Losing message attribution is acceptable for the core RAG goal — the emotional signals and preference markers still exist in the text even without knowing which specific line came from whom.

**Screenshot OCR**

OCR (Optical Character Recognition) is the most complex input path. For the MVP, screenshot support is deferred. The recommended approach when added is to use a third-party OCR API (such as Google Cloud Vision or AWS Textract's free tier) rather than running a local OCR model — local models are too heavy for Render free tier.

**The 2MB File Size Limit — Design Rationale**

The maximum accepted file upload size is **2MB** (configurable via the `MAX_FILE_SIZE_BYTES` environment variable). This is a deliberate increase from the MVP's ~50KB text assumption, chosen to accommodate:
- Long-running WhatsApp conversations spanning months or years
- Group chat exports (which can be very dense)
- Users with very active messaging relationships

**What 2MB means in practice:**

A typical WhatsApp `.txt` export is approximately 1KB per 10 messages (text only, no media). 2MB ≈ 20,000 messages. This is a very long conversation — years of chat history.

However, processing 20,000 messages through the full pipeline (anonymization, chunking, embedding) would:
- Generate ~2,000 chunks at a 10-message sliding window
- Require ~2,000 embedding API calls
- Cost significantly more per session (~$0.04 for embeddings alone, not counting completions)

**The resolution: accept 2MB files but process a representative sample.** After parsing, the pipeline samples the conversation intelligently before chunking:
- **Recency bias:** Weight the most recent 25% of messages higher (they reflect current preferences)
- **Density sampling:** From the full conversation, select the top 40-60 message windows that have the highest lexical diversity (most varied vocabulary = most useful signal)
- **Hard cap:** Feed a maximum of 400 messages (~40 chunks) into the chunking and embedding stages regardless of conversation length

This means the 2MB limit is an *input* limit, not a *processing* limit. Users get the benefit of uploading their full conversation history, while the system processes a curated, cost-efficient sample.

**Input Validation Constraints**

Before any processing begins, the input layer enforces:
- **Maximum file size: 2MB** (enforced at the HTTP layer via `MAX_FILE_SIZE_BYTES` env var — Gin's body size limit middleware is set to this value)
- **Minimum conversation length:** 5 parseable messages. Below this, there is not enough signal to generate meaningful insights.
- **File type validation:** Only `.txt` accepted in MVP (MIME type check + extension check)
- **Budget range:** Must be one of the defined tiers — it is not a free-text field
- **Occasion:** Must be from a predefined list
- **Session UUID:** Must be a valid UUID format (generated client-side by the frontend)

**Why Input Validation Matters for RAG**

RAG quality is directly dependent on input quality. Garbage in, garbage out applies especially strongly here because the chunker and embedder operate on whatever the parser produces. A malformed parse that merges multiple messages into one giant string will produce semantically confused chunks and poor embeddings. Investing in a robust parser pays dividends throughout the entire pipeline.

---

### Design Decisions
- The parser outputs a **normalized Message slice**, not raw text. This gives every downstream stage a consistent, typed interface to work with.
- Screenshot/OCR support is **deferred to a later phase** — it adds significant complexity and cost with marginal benefit for the MVP.
- **File size limit is 2MB**, enforced at the HTTP middleware layer before the request body reaches application code. This is configured via `MAX_FILE_SIZE_BYTES` env var.
- **Intelligent sampling** is applied after parsing for large conversations — accept up to 2MB but process a curated maximum of ~400 messages through the RAG pipeline.
- **Session UUID is generated client-side** and attached to every request, enabling multi-window isolation at the Pinecone namespace level.

---

### Alternative Approaches
- **Send the raw text to GPT-4o and ask it to parse the conversation:** This would work but costs tokens, is slow, and conflates parsing with reasoning. Keep parsing deterministic and cheap.
- **Accept the full 2MB and process all of it:** Would work for embedding cost (~$0.04) but produces hundreds of chunks, most of which are low-signal noise. Smart sampling produces better retrieval results with fewer chunks.
- **Accept only plain text, no file upload:** Simpler implementation but significantly worse UX. Most WhatsApp users export as `.txt`.

---

### Trade-offs / Consequences
- **Smart sampling means very old conversations are under-represented.** A 3-year chat export will have its oldest messages sampled less. Recency bias is intentional — recent conversations reflect current preferences better.
- **2MB limit** is generous enough for real-world conversation exports while preventing abuse (e.g., uploading a 50MB log file to consume API credits). The configurable env var means this can be tuned without redeployment.
- **Each window is independent.** If a user opens 5 windows and submits all 5 simultaneously, the backend receives 5 independent requests and processes them in parallel. Each session has its own Pinecone namespace and its own OpenAI calls. There is no shared state to conflict.

---

### How This Module Connects to the Overall System
The parser's output (normalized Message slice) is the input to the Anonymizer (Module 4). The quality of parsing directly determines the quality of anonymization and chunking.

---

<a name="module-4"></a>
## Module 4 — Anonymization & Privacy-Safe PII Handling

### Objective
Design the anonymization layer — the most critical privacy component in the system. This is where the architectural promise "we don't send your raw conversation to OpenAI" is actually enforced technically, not just promised in a privacy policy.

---

### Concepts to Learn
- PII (Personally Identifiable Information) categories and detection
- Named Entity Recognition (NER) as a technique
- Pseudonymization vs. anonymization
- Privacy by design as an architectural pattern
- What "data minimization" means in practice

---

### Detailed Explanation

**Why This Layer Exists**

OpenAI's API data handling policy (as of the time of writing) states that API inputs are not used to train models by default, but the data does transit their systems. Even with that policy, sending a raw conversation like:

> "Priya: Hey Riya, remember when we went to Bangalore for Aunt Meena's wedding? Lol you left your passport at the Marriott"

...to a third-party API means real names, real relationships, a real city, a real hotel brand, and a real event are leaving your system. Users have not consented to this. It is a trust violation even if it is not technically a policy violation.

The anonymizer exists to transform this before it ever reaches the embedding or completion API calls.

**What PII Looks Like in Conversational Data**

In a WhatsApp-style conversation, PII appears as:
- **Person names:** First names, nicknames, family titles used as names ("Didi", "Bhai")
- **Place names:** Cities, neighborhoods, specific venues, restaurants
- **Institution names:** School names, company names, hospital names
- **Dates tied to identity:** Birthdays, anniversaries (though dates alone are usually acceptable)
- **Contact information:** Phone numbers, email addresses (unlikely in WhatsApp but possible)
- **Relationship descriptions:** "my husband", "our boss" — not names, but identifying context

**The Anonymization Strategy: Pseudonymization with Stable Tokens**

The chosen approach is *pseudonymization* — replacing identified entities with stable, consistent placeholder tokens throughout the conversation. "Stable" is important: if "Priya" appears 40 times in the conversation, all 40 instances are replaced with the same token (e.g., `[Person_A]`). This preserves the conversational structure and relationship patterns even after anonymization.

The token assignment:
- `[Person_A]` = the conversation uploader (identified by the sender field)
- `[Person_B]` = the primary recipient (the gift recipient)
- `[Person_C]`, `[Person_D]`, etc. for other named individuals
- `[City_1]`, `[City_2]` for places
- `[Company_1]` for organizations

After anonymization, the earlier example becomes:
> "[Person_A]: Hey [Person_B], remember when we went to [City_1] for [Person_C]'s wedding? Lol you left your passport at [Company_1]"

This is now safe to embed and pass to OpenAI. The emotional content — excitement, shared memory, humor, relationship warmth — is entirely preserved. The PII is not.

**Implementing NER in Go**

Go does not have mature native NLP libraries the way Python does (spaCy, NLTK). The options are:

1. **Regex-based heuristics:** Fast, no external dependencies, works well for capitalized proper nouns in English. Limited accuracy for edge cases and non-English names common in Indian contexts.

2. **Lightweight API-based NER:** Call a small, cheap NER API (e.g., a dedicated NER model via Hugging Face Inference API free tier) for named entity detection. More accurate, but adds a network call and dependency.

3. **Use the sender field from parsing as the primary name source:** Since WhatsApp exports include sender names, you already *know* the names of all participants. Use the parsed sender names as the primary NER input, and supplement with regex for names that appear mid-message. This is the most practical approach for GiftSense.

The recommended strategy is **Option 3 + Option 1**: use parsed sender names as the seed entity list, then run a capitalized proper noun regex scan to catch other names mentioned in message bodies. This handles the 80% case efficiently without external dependencies.

**Before/After Anonymization — Concrete Example**

*Before anonymization (what the parser produces):*
```
[Sender: Riya] "Mom I want to do pottery classes so badly, 
                 Nisha was telling me about this place in Bandra"
[Sender: Mom]  "Haha my Riya and her hobbies 😄 first guitar, 
                 then watercolors, now pottery!"
[Sender: Riya] "This time I'm serious! Also can we go to that 
                 Persian restaurant in Colaba for my birthday?"
```

*After anonymization (what OpenAI receives):*
```
[Person_A] expressed strong interest in pottery classes, 
            mentioning a recommendation from [Person_C] 
            about a place in [City_1].
[Person_B] responded warmly, referencing [Person_A]'s 
            pattern of enthusiastically pursuing creative hobbies 
            (guitar, watercolors, now pottery).
[Person_A] reaffirmed commitment to this hobby and requested 
            a celebration dinner at a specific restaurant type 
            ([City_2] area).
```

Notice two things: (a) all PII is replaced, and (b) this example shows the output *after* anonymization *and* before chunking — in practice, it may make sense to do light summarization of message clusters during anonymization to further reduce raw text exposure.

**What Emotional Signal is Preserved**

The anonymized text still clearly signals:
- Person A is creative and enthusiastic about hands-on crafts
- There is a pattern of trying multiple creative hobbies
- Person B (the gift recipient — the mom) teases warmly, suggesting a playful relationship
- Person A has a preference for experiential activities and specific dining preferences

All of this is retained after anonymization. The RAG pipeline has everything it needs.

---

### Design Decisions
- **Pseudonymization (not full anonymization):** Replacing names with stable tokens rather than deleting them entirely preserves relationship structure, which is crucial for personality insight generation.
- **NER via sender-field seeding + regex:** Practical balance between accuracy and implementation complexity in Go without Python NLP dependencies.
- **Anonymization happens before any OpenAI call, including the embedding call.** The embedding model embeds the pseudonymized text. This is non-negotiable.

---

### Alternative Approaches
- **Full redaction (delete names entirely):** Simpler but loses relationship context. "Nisha was telling me" becomes "was telling me" — the social circle signal disappears.
- **Generalization (replace names with categories):** Replace "Riya" with "[Daughter]", "Mom" with "[Mother]". Richer than tokens but requires relationship inference which is complex and error-prone.
- **Send raw text to OpenAI and rely on their data policy:** Not acceptable. Privacy should be a technical guarantee, not a policy promise.
- **Client-side anonymization (in the browser before upload):** Interesting approach — the raw conversation never leaves the user's device. Complex to implement reliably in JavaScript, but worth exploring as a future enhancement for maximum privacy.

---

### Trade-offs / Consequences
- **Regex-based NER misses edge cases:** Indian names that are not capitalized in chat contexts (e.g., "bhai", "didi" used as names) may not be caught. Mitigation: the sender field captures the major participants; regex handles in-body names for common cases.
- **Pseudonymization is reversible in theory:** The mapping from "Riya" to "[Person_A]" exists in memory during the session. If that mapping were logged or persisted, PII could be recovered. The mapping must *never* be logged, stored, or included in any API call payload.

---

### How This Module Connects to the Overall System
The anonymized, normalized message data flows into the Chunker (Module 5). The anonymization token map is held in memory for the session duration and discarded with the session.

---

<a name="module-5"></a>
## Module 5 — Chunking Strategy for Conversational Data

### Objective
Design the chunking layer — the stage that splits the anonymized conversation into embedding-ready units. Understand why chunking strategy is one of the highest-leverage decisions in a RAG pipeline, and why conversational data requires different thinking than document chunking.

---

### Concepts to Learn
- Why chunking exists in RAG
- Fixed-size vs. semantic vs. sentence-window chunking
- The chunk size dilemma: specificity vs. context
- Chunk overlap and why it exists
- Metadata enrichment at chunk creation time
- Conversational data as a unique chunking challenge

---

### Detailed Explanation

**Why You Can't Embed the Whole Conversation**

An embedding model converts text into a fixed-size vector. That vector is a compressed representation of the semantic meaning of the input. If you embed an entire 200-message conversation as one vector, you get one point in vector space that represents the "average" meaning of the whole conversation. This is almost useless for retrieval — you can never find the specific cluster of messages about pottery classes or the recurring theme of food preferences, because they are all averaged into one blurry point.

Chunking solves this by creating many small, semantically focused units, each with its own embedding vector. Now each chunk occupies a distinct point in the vector space that accurately represents its specific semantic content.

**The Chunk Size Dilemma**

Chunk size is a fundamental tension in RAG:

- **Too small (e.g., single messages):** Each chunk lacks context. A single message like "[Person_A]: Yes!" is meaningless without the surrounding exchange. The embedding will be imprecise. Retrieval will return many contextually empty chunks.

- **Too large (e.g., 50 messages at once):** Each chunk contains too many topics. A chunk about pottery, birthday plans, and a work complaint all mixed together will produce an embedding that is "about everything" and "specifically about nothing." Retrieval precision degrades.

The sweet spot for conversational data is **semantically coherent exchanges** — clusters of messages that revolve around a single topic, emotion, or theme.

**Chunking Strategy for GiftSense: Sliding Window with Semantic Boundaries**

The recommended approach is a hybrid:

1. **Sliding window:** Group messages into windows of N consecutive messages (e.g., N=8). This gives each chunk enough context to be coherent.

2. **Overlap:** Consecutive windows overlap by K messages (e.g., K=3). This ensures that a topic that spans a window boundary is represented in at least one complete chunk.

3. **Soft semantic boundaries:** Optionally, detect topic shifts using simple heuristics (long time gap between messages, change in subject keywords, shift in emotional tone) and use those as hard chunk boundaries regardless of window size.

**For GiftSense specifically**, a window of 6-10 messages with 3-message overlap tends to work well because:
- WhatsApp conversations are often short-message, rapid-fire exchanges
- Emotional context (humor, excitement, affection) spans multiple turns
- Gift-relevant signals (hobbies, wishes, preferences) often emerge over 4-6 messages

**What a Chunk Looks Like**

After chunking, each chunk is a structured object:

```
Chunk {
    ID:             string     // unique ID for this chunk
    SessionID:      string     // ties this chunk to the session
    AnonymizedText: string     // the anonymized messages in this window
    StartIndex:     int        // first message index in this chunk
    EndIndex:       int        // last message index in this chunk
    Metadata: {
        Topics:          []string   // detected topic keywords (hobby, food, travel, etc.)
        EmotionalMarkers: []string  // detected signals (humor, warmth, excitement, nostalgia)
        HasPreference:   bool       // does this chunk contain a stated preference?
        HasWish:         bool       // does this chunk contain an expressed wish?
    }
}
```

**Metadata Enrichment — Why It Matters**

The metadata fields on each chunk are computed at chunk creation time using simple keyword matching and pattern detection. These are not perfect — they are lightweight heuristics. Their value is in enabling **metadata-filtered retrieval** later. When constructing a retrieval query for "birthday gift for mom who loves creative hobbies," you can filter the vector store to prioritize chunks where `HasPreference: true` or `Topics: ["hobby", "craft"]`. This improves retrieval precision without increasing embedding cost.

**A Concrete Example**

Given this anonymized input:
```
[Person_A]: I've been watching so many cooking videos lately
[Person_B]: Haha you always say that and then order Zomato 😂
[Person_A]: No this time seriously! I want to learn to make pasta from scratch
[Person_B]: Sure sure 😄 what else is new
[Person_A]: I also really want to read more, I haven't finished a book in ages
[Person_B]: You have like 30 books on your shelf you've never opened
```

This becomes one chunk with metadata:
- Topics: `["cooking", "food", "reading", "books"]`
- EmotionalMarkers: `["humor", "aspiration", "self-awareness"]`
- HasPreference: `true` (pasta-making, reading)
- HasWish: `true` (expressed desire to learn cooking, read more)

This chunk will be highly retrievable when the query is "what are this person's hobbies or interests?"

---

### Design Decisions
- **Sliding window with overlap** over semantic boundary detection. Semantic boundary detection requires an additional LLM or NLP call — too expensive and slow for a free-tier pipeline. Sliding window with overlap is a known-good approximation.
- **Metadata enrichment via heuristics** rather than LLM-based classification. Keeps the pipeline cheap and fast. Quality is sufficient for the filtering task.
- **Chunk size of 6-10 messages.** Validated as appropriate for WhatsApp-style short-message conversations.

---

### Alternative Approaches
- **LLM-based semantic chunking:** Send the conversation to GPT-4o and ask it to identify thematically coherent segments. Very high quality chunks, but expensive and slow. Not viable for free tier.
- **Fixed character-count chunking:** Simple to implement but ignores message boundaries, often splitting a message mid-sentence. Poor quality for conversational data.
- **Single-sentence chunking:** Each sentence is a chunk. Very precise but loses inter-sentence context. Works well for formal documents, poorly for casual chat.
- **Summary chunks:** Instead of embedding raw message text, generate a one-sentence summary of each 10-message window and embed the summary. Reduces noise, improves embedding quality, but adds an LLM call per chunk — expensive.

---

### Trade-offs / Consequences
- **Overlap increases the total number of chunks**, which means more embedding API calls. With N=8 and K=3, a 100-message conversation produces roughly 20-25 chunks. At ~0.0001 USD per chunk (text-embedding-3-small), this is negligible.
- **Heuristic metadata is imperfect.** A chunk about "I need to see a doctor" might incorrectly be tagged with `HasWish: true`. This is acceptable — retrieval is probabilistic, not deterministic. One noisy chunk among 25 does not ruin the output.

---

### How This Module Connects to the Overall System
Each chunk object (with its text and metadata) is passed to the Embedder (Module 6), which calls OpenAI to convert the text to a vector. The chunk's metadata travels alongside the vector into the vector store, enabling filtered retrieval.

---

<a name="module-6"></a>
## Module 6 — Embeddings & Semantic Vector Space

### Objective
Understand what embeddings are, how they work, and why they are the foundation of semantic retrieval. Move beyond "embeddings are vectors" to a genuine intuition for what they represent and why the choice of embedding model matters.

---

### Concepts to Learn
- What an embedding is (conceptually and mathematically)
- The vector space model and semantic similarity
- Cosine similarity as a distance metric
- Why dimensionality matters
- OpenAI's embedding model options and their trade-offs
- The embedding API call and what it costs

---

### Detailed Explanation

**What Is an Embedding?**

An embedding is a function that takes a piece of text and outputs a list of floating point numbers (a vector). The vector is designed so that texts with similar meanings produce vectors that are close together in the high-dimensional space.

"I love hiking in the mountains" and "She enjoys outdoor trekking" will produce vectors that are geometrically close to each other, because the embedding model was trained to encode semantic meaning — not just surface-level word matching.

"She enjoys outdoor trekking" and "The quarterly earnings report exceeded expectations" will produce vectors far apart, because their semantic content is unrelated.

This geometric relationship is what makes retrieval possible: instead of searching for exact keyword matches, you search for vectors close to your query vector.

**The Vector Space Intuition**

Imagine a 3-dimensional space (for intuition — real embeddings use 1536 dimensions). Each axis represents some dimension of meaning. Texts about outdoor activities cluster together in one region. Texts about food and cooking cluster in another. Texts about relationships and emotions occupy yet another region.

When you embed a query like "what are her hobbies and passions?", you get a vector in this space. The nearest neighbors of that vector are the chunks of the conversation that discuss hobbies and passions — even if those chunks never use the word "hobby."

This is the core power of semantic retrieval: **vocabulary-independent meaning matching.**

**Cosine Similarity**

The standard distance metric for embedding retrieval is cosine similarity. It measures the angle between two vectors rather than the absolute distance. This is preferable to Euclidean distance because it is invariant to vector magnitude — a short chunk and a long chunk that discuss the same topic will have similar angles even if their vectors have different lengths.

Cosine similarity produces a score between -1 and 1:
- 1.0 = identical semantic direction (perfect match)
- 0.0 = orthogonal (completely unrelated)
- -1.0 = semantically opposite

In practice, for good retrieval you want scores above 0.75-0.80. A retrieval hit with cosine similarity below 0.6 is often a noisy match.

**OpenAI's Embedding Models**

OpenAI offers two primary embedding models relevant to this project:

| Model | Dimensions | Cost per 1M tokens | Best For |
|---|---|---|---|
| `text-embedding-3-small` | 1536 | ~$0.02 | Cost-efficient general use, high quality |
| `text-embedding-3-large` | 3072 | ~$0.13 | Maximum quality, higher cost |
| `text-embedding-ada-002` | 1536 | ~$0.10 | Legacy, being phased out |

For GiftSense, `text-embedding-3-small` is the clear choice. The quality difference between small and large is marginal for conversational text retrieval. The cost difference is 6.5x. Given the free-tier context and the goal of keeping the per-session cost low, `text-embedding-3-small` is optimal.

**What the Embedding Call Looks Like (Conceptually)**

The embedder takes each anonymized chunk's text and makes one API call to OpenAI's embedding endpoint. The input is the text string. The output is an array of 1536 floating point numbers.

This is done for every chunk at index time (when the conversation is first processed), and again for the query at retrieval time.

**The Embedding Call Boundary**

At this stage, the data that crosses the OpenAI boundary is:
- The anonymized chunk text (e.g., "[Person_A] expressed interest in pottery and creative hobbies")
- Nothing else — no names, no metadata, no session IDs

OpenAI returns only a vector. No text is returned. No content is persisted by OpenAI beyond their stated data retention window (zero retention on the API by default, unless zero-data-retention is explicitly configured — worth checking OpenAI's current policy).

---

### Design Decisions
- **`text-embedding-3-small`** for all embeddings — both index-time (chunk embedding) and query-time (query embedding). Using the same model for both is non-negotiable: similarity scores only make sense when both vectors exist in the same embedding space.
- **No local embedding models.** Running a local embedding model (e.g., via GGUF quantization) on Render's free tier with 512MB RAM is not viable. OpenAI's API is the only reasonable option given the constraints.

---

### Alternative Approaches
- **`text-embedding-3-large`:** Higher quality but 6.5x more expensive. Not justified for conversational text where `small` already performs well.
- **Sentence Transformers via API (Hugging Face):** Free, high-quality open models available via Hugging Face Inference API. Tempting on cost, but introduces a second vendor dependency and a different embedding space, making the pipeline more complex. Mixing embedding providers is a source of subtle, hard-to-debug bugs.
- **TF-IDF or BM25 (non-neural retrieval):** No embedding API calls needed, purely keyword-based. Fast and cheap. Missing the core advantage of semantic retrieval — would not find "outdoor adventure" chunks in response to a query about "travel and exploration."

---

### Trade-offs / Consequences
- **Vendor dependency on OpenAI for embeddings:** If OpenAI raises embedding prices or the API is unavailable, the embedding stage fails. The Go interface wrapper (discussed in Module 12) mitigates this by making the provider swappable without rewriting business logic.
- **Embedding quality is fixed:** Once chunks are embedded with `text-embedding-3-small`, that is the representation used for the entire session. There is no re-embedding with a different model mid-session. This is not a practical concern for GiftSense but is relevant at scale.

---

### How This Module Connects to the Overall System
Each chunk's embedding vector is stored in the Pinecone index (Module 7) under the session's namespace, alongside the chunk metadata. The chunk text is held in Go process memory keyed by chunk ID. The retrieval query is embedded using the same model, and the Pinecone query API performs cosine similarity search to find the most relevant chunks.

---

<a name="module-7"></a>
## Module 7 — Vector Store Design with Pinecone Free Tier

### Objective
Understand what a vector store is, how Pinecone works, and why Pinecone's namespace model is the correct choice for GiftSense given the requirements of multi-window isolation, ephemeral session data, and free-tier infrastructure. Learn how to design a session-scoped vector lifecycle — create, populate, query, delete — within a single request.

---

### Concepts to Learn
- What a vector store is and how it differs from a relational database
- Pinecone's data model: indexes, namespaces, and records
- Session-scoped namespace design for ephemeral data
- Approximate Nearest Neighbor (ANN) algorithms (HNSW, IVF)
- Metadata filtering in Pinecone
- The privacy implications of an external vector store
- Cleanup strategies for ephemeral namespaces (delete-on-complete vs. TTL sweep)

---

### Detailed Explanation

**What Is a Vector Store?**

A vector store is a database optimized for storing and querying high-dimensional vectors. Unlike a relational database where you query by exact field values ("WHERE user_id = 42"), a vector store answers queries of the form: "give me the K vectors most similar to this query vector." The core operation is **nearest neighbor search** — finding the K stored vectors geometrically closest (by cosine similarity) to a query vector.

At scale, computing this naively (compare query against every stored vector one-by-one) is prohibitively slow. Production vector stores use **Approximate Nearest Neighbor (ANN) algorithms** such as HNSW (Hierarchical Navigable Small World graphs) that trade a small recall loss for massive speed gains. Pinecone uses HNSW internally for its serverless tier.

**Pinecone's Data Model**

Pinecone organizes data in three levels:

1. **Project:** Your Pinecone account — one project for GiftSense.

2. **Index:** A named vector space with a fixed dimensionality. For GiftSense, you create one index: `giftsense`, configured for 1536 dimensions (matching `text-embedding-3-small`'s output) with cosine similarity as the distance metric. On the free tier, you are limited to 1 index.

3. **Namespace:** A logical partition within an index. Vectors in different namespaces are completely isolated — a query against namespace `abc-123` will never return vectors from namespace `def-456`. **This is the key mechanism for multi-window session isolation.** Each browser window session gets its own namespace, named after the session UUID.

4. **Records:** Individual vector entries within a namespace. Each record has an ID (the chunk ID), a vector (the embedding), and a metadata object (chunk topics, emotional markers, preference flags).

**Session-Scoped Namespace Lifecycle**

Every GiftSense request follows this exact Pinecone lifecycle:

```
Request Start
    │
    ├── Namespace: {session_uuid}  ← created implicitly on first upsert
    │
    ├── UPSERT: ~25-40 chunk vectors into namespace {session_uuid}
    │   Each record: { id: chunk_id, values: []float32, metadata: { ... } }
    │
    ├── QUERY: 4 × top-{TOP_K} similarity searches against namespace {session_uuid}
    │   Filter: metadata conditions (HasPreference, Topics, etc.)
    │
    ├── (GPT-4o completion and link generation happen here)
    │
    └── DELETE: namespace {session_uuid}  ← all vectors permanently removed

Request End
```

This lifecycle means that vectors exist in Pinecone only for the duration of a single request. Once the response is returned to the frontend, the namespace deletion is issued. No historical conversation data accumulates in Pinecone.

**Pinecone Free Tier Constraints**

| Constraint | Free Tier Limit | GiftSense Usage |
|---|---|---|
| Indexes | 1 index | 1 index (`giftsense`) |
| Storage | ~100K vectors (serverless) | ~40 vectors per session × concurrent sessions |
| Namespaces | Unlimited within index | 1 per active session |
| API calls | Generous free allowance | ~30 upserts + 4 queries + 1 delete per session |
| Dimensions | Up to 1536 | 1536 |
| Region | Specific free region | Use the closest free region (e.g., us-east-1) |

GiftSense's active vector footprint at any given moment is tiny: `active_concurrent_sessions × ~40 vectors`. Even at 50 concurrent sessions, that is ~2,000 vectors — a fraction of the free tier limit.

**Metadata Filtering in Pinecone**

Pinecone supports metadata filtering as part of a query. You can combine vector similarity with metadata conditions:

```
Query: {
    vector: [query_embedding],
    topK: 3,
    namespace: "session-uuid-here",
    filter: {
        "has_preference": { "$eq": true }
    }
}
```

This means for the "gift interests" retrieval query, you filter to only chunks where `has_preference` is true, and Pinecone returns the top-3 most similar chunks that also satisfy that filter. This dramatically improves retrieval precision: you are not just finding semantically similar chunks, you are finding semantically similar chunks that actually contain expressed preferences.

Available metadata filter operators in Pinecone: `$eq`, `$ne`, `$in`, `$nin`, `$gt`, `$gte`, `$lt`, `$lte`. For GiftSense's boolean and array metadata, `$eq` and `$in` cover all filter cases.

**Metadata Schema for Pinecone Records**

```
PineconeRecord {
    id:     "{session_id}_{chunk_index}"    // unique within the index
    values: []float32                        // 1536-dimensional embedding vector
    metadata: {
        "has_preference":  bool
        "has_wish":        bool
        "topics":          []string          // e.g., ["cooking", "travel"]
        "emotional_markers": []string        // e.g., ["humor", "warmth"]
        "chunk_index":     int
        "message_start":   int
        "message_end":     int
    }
}
```

Note: The `anonymized_text` of the chunk is **not stored in Pinecone metadata**. Storing text in Pinecone metadata is possible but unnecessary — it costs storage, and you already have the text in Go process memory during the request. The retrieval step returns matching record IDs, which the Go backend uses to look up the corresponding chunk text from a local `map[string]Chunk` built at upsert time.

This is an important design subtlety: **Pinecone stores vectors and metadata; Go memory stores the text.** The two are joined by chunk ID during retrieval. No conversation text ever reaches Pinecone.

**Cleanup: Delete-on-Complete vs. Periodic Sweep**

The primary cleanup strategy is **delete-on-complete**: after assembling the response, the Go handler issues a Pinecone namespace delete before returning. This covers the happy path.

For failure scenarios (Go backend crashes between response and delete, or the delete call itself fails due to a Pinecone API error), a **periodic background sweep** is needed:

- A Go goroutine runs every 30 minutes as a background task
- It calls the Pinecone List Namespaces API (or tracks active namespaces in a short-lived in-memory set)
- Any namespace older than 1 hour is deleted
- This sweep is best-effort: it runs while the Render instance is alive, but if the instance is spun down, the sweep does not run

Additionally, Pinecone's serverless tier eventually compacts unused namespaces, providing a long-term safety net even if both cleanup paths fail.

**The Go Client for Pinecone**

Pinecone provides an official Go client library (`github.com/pinecone-io/go-pinecone`). The client is initialized once at application startup (in `main.go`) with the API key from the environment variable `PINECONE_API_KEY`, and injected into the application via the `VectorStore` interface — exactly as described in Module 12. The business logic never imports the Pinecone client directly.

---

### Design Decisions
- **Single shared index with per-session namespaces**, not one index per session. Creating a new Pinecone index per session would be extremely slow (index creation takes seconds to minutes) and is limited on the free tier. Namespaces are instant and unlimited.
- **Chunk text stored in Go memory, not Pinecone metadata.** Only vectors and boolean/string metadata are stored in Pinecone. This keeps Pinecone from ever receiving readable conversation content.
- **Delete-on-complete + periodic sweep** as the two-layer cleanup strategy. The primary path handles the happy path; the sweep handles failures.

---

### Alternative Approaches
- **In-process in-memory store:** No external dependency, maximum privacy, slightly lower complexity. However, it does not provide clean multi-window isolation (separate goroutines share the same Go process memory and need explicit keying), does not survive Render's occasional instance recycling, and offers no ANN capabilities at scale. See original v1 document for full analysis.
- **Redis with vector search (Redis Stack):** Excellent in-memory vector store. Requires a separate Redis service (another free-tier slot), and Redis free tier on Render is limited. The namespace isolation model is not as natural as Pinecone's.
- **Weaviate free tier:** More powerful metadata filtering, multi-tenancy support. Heavier SDK, more complex setup. Overkill for GiftSense's use case.
- **pgvector on Render's free PostgreSQL:** Keeps everything in one service. But Render's free PostgreSQL has storage limits, no guaranteed persistence on free tier, and mixing conversation vectors with a relational DB feels architecturally misaligned.

---

### Trade-offs / Consequences
- **Pinecone introduces a new external data boundary.** Anonymized vectors leave your Go server and live in Pinecone's infrastructure (AWS us-east-1 on the free tier). The privacy argument: vectors are floating-point numbers — you cannot reconstruct the original text from a vector alone. The anonymized text itself never reaches Pinecone.
- **Network latency for upsert and query.** Each Pinecone operation adds ~50-150ms of network latency (Render US East → Pinecone US East). With concurrent goroutines for upserts, this is minimized. Query calls are sequential per query (4 queries × 4 parallel goroutines = effectively 1-2 round trips of latency).
- **Pinecone free tier availability.** Free tier services can have occasional API instability. The Go error handling layer must handle `503` responses from Pinecone gracefully with retry logic (exponential backoff, max 3 retries).

---

### How This Module Connects to the Overall System
The Pinecone namespace acts as the session's vector storage. Chunk embeddings from Module 6 are upserted here. Retrieval queries from Module 8 are executed against this namespace. After the pipeline completes (Module 9-10), the namespace is deleted, closing the ephemeral lifecycle.

---

<a name="module-8"></a>
## Module 8 — Retrieval Strategy & Query Construction

### Objective
Design the retrieval layer — the stage that translates the user's intent (suggest a birthday gift for this person) into an effective semantic search over the vector store. Understand how to construct retrieval queries that surface the right chunks, and why retrieval quality is the single most important factor in RAG output quality.

---

### Concepts to Learn
- Dense vs. sparse vs. hybrid retrieval
- Query construction strategies for RAG
- Top-K retrieval and choosing the right K
- Re-ranking retrieved chunks
- The "retrieval is not search" mindset
- How recipient metadata shapes retrieval

---

### Detailed Explanation

**Retrieval Quality Is the Limiting Factor**

A common mistake in RAG system design is to obsess over the LLM prompt while neglecting retrieval. But consider: if the retrieved chunks do not contain the relevant information, no amount of prompt engineering can help GPT-4o generate a grounded, accurate response. The model can only reason over what it is given. Garbage retrieved → garbage generated.

Conversely, if the retrieved chunks are highly relevant and information-dense, even a simple prompt will produce excellent output.

**Dense vs. Sparse Retrieval**

*Dense retrieval* (what GiftSense uses) works with embedding vectors. The query is embedded into the same vector space as the chunks. Nearest neighbor search finds semantically similar chunks, even without keyword overlap.

*Sparse retrieval* (BM25, TF-IDF) works on keyword frequency. The query and documents are represented as sparse vectors of term weights. This is excellent for exact-match scenarios but misses semantic relationships.

*Hybrid retrieval* combines both: rank chunks by both semantic similarity and keyword frequency, then merge the ranked lists. Hybrid retrieval generally outperforms either alone. For GiftSense, with small corpora (≤30 chunks) and rich semantic signal, dense-only retrieval is sufficient. Hybrid retrieval is a meaningful enhancement for a Phase 3 improvement.

**Query Construction for GiftSense**

The retrieval query is not simply "gift for mom birthday ₹2000." That query is technically correct but semantically shallow. A well-constructed query asks for the information the LLM actually needs to generate a good gift suggestion.

GiftSense should issue **multiple targeted retrieval queries** per session, each designed to surface different types of relevant context:

*Query 1 — Interests and hobbies:*
"What activities, hobbies, passions, and interests does this person mention enjoying or wanting to pursue?"

*Query 2 — Personality and emotional patterns:*
"What personality traits, emotional patterns, and communication style characterize this person? What makes them happy? What do they find funny?"

*Query 3 — Expressed wishes and desires:*
"What things has this person explicitly said they want, wish for, or plan to do someday?"

*Query 4 — Relationship dynamics:*
"What is the nature of the relationship between the two people in this conversation? What are the recurring themes and shared experiences?"

Each query is embedded separately and retrieves the top 3-5 most relevant chunks. The union of these retrieved chunk sets (deduplicated) becomes the context for the GPT-4o prompt. This multi-query retrieval strategy significantly improves coverage — a single query often misses signals that a different framing of the same question would have caught.

**Incorporating Recipient Metadata into Queries**

The recipient metadata (relation, occasion, budget, gender) should be woven into the query text, not treated separately:

"What hobbies and interests has [Person_B] mentioned that could inform a birthday gift for a mother from her child, with a budget of ₹1000-₹5000?"

Including this context in the query vector means the embedding captures the occasion-and-relation framing, biasing retrieval toward chunks that are relevant not just to hobbies in general, but to the kind of hobbies that translate into giftable items.

**Top-K and Chunk Deduplication**

Each query retrieves top-3 chunks. With 4 queries, the retrieved set is at most 12 chunks (likely fewer due to deduplication, since the most central chunks will appear in multiple query results). 

12 chunks × average 400 tokens per chunk ≈ 4,800 tokens of context. This is well within GPT-4o's context window and cost-effective.

**Re-Ranking**

After retrieval, a simple re-ranking step improves quality. Re-ranking reorders the retrieved chunks to maximize the relevance and diversity of the context:

1. **Deduplicate** overlapping chunks (chunks that share > 50% of their messages)
2. **Prioritize chunks with high metadata signals** (HasPreference, HasWish) regardless of similarity score
3. **Ensure diversity** — avoid passing 5 chunks all about the same topic (e.g., all about cooking) when other topics exist

For GiftSense with ≤12 retrieved chunks, this re-ranking is a lightweight pass, not a full neural re-ranking model (which would be overkill and expensive).

---

### Design Decisions
- **Multi-query retrieval (4 targeted queries)** over a single composite query. Each targeted query surfaces different signal types that a composite query would blend into mediocrity.
- **Dense retrieval only** for the MVP. Hybrid retrieval is reserved for a later optimization phase.
- **Lightweight heuristic re-ranking** over neural re-ranking models. Neural re-rankers (e.g., Cohere Rerank) add cost, latency, and another API dependency. Heuristic re-ranking is sufficient for 12-chunk sets.

---

### Alternative Approaches
- **Single composite retrieval query:** Simpler but lower quality. The embedding of a long, multi-aspect query tends to pull toward the "average" of all the aspects, which is often the center of the vector space (matching nothing specifically well).
- **HyDE (Hypothetical Document Embeddings):** Instead of embedding the query directly, ask GPT-4o to generate a *hypothetical answer* to the query and embed that instead. The hypothetical answer's embedding often aligns better with the chunks. Clever technique, but adds an LLM call just for retrieval — expensive for this use case.
- **Query expansion:** Use GPT-4o to rewrite the query in multiple ways, retrieve for each rewriting, and merge results. Similar to multi-query but LLM-generated. More expensive, better quality.

---

### Trade-offs / Consequences
- **4 embedding API calls for retrieval queries** adds to the per-session cost (discussed in detail in Module 14). Each is cheap individually; the concern is latency (4 sequential API calls) — these should be issued concurrently using Go goroutines.
- **Heuristic re-ranking can miss relevance signals** that a neural re-ranker would catch. Acceptable for this project's quality bar.

---

### How This Module Connects to the Overall System
The re-ranked, deduplicated retrieved chunk set is passed to the Prompt Builder (Module 9), which assembles the context into a GPT-4o completion request.

---

<a name="module-9"></a>
## Module 9 — Prompt Engineering for RAG — Personality Insights & Gift Suggestions

### Objective
Design the prompt engineering layer — how the retrieved context, recipient metadata, and output requirements are assembled into a prompt that reliably produces structured, grounded, creative outputs from GPT-4o.

---

### Concepts to Learn
- Anatomy of a RAG prompt (system prompt, context injection, user instruction)
- Grounding constraints — preventing hallucination
- Structured output via JSON mode
- Token budget management
- GPT-4o vs. GPT-4o-mini: the cost-quality trade-off
- Persona and tone instruction in LLM prompts
- Budget compliance enforcement in prompts

---

### Detailed Explanation

**The Anatomy of a RAG Prompt**

A well-structured RAG prompt has three distinct sections:

1. **System Prompt:** Sets the model's persona, capabilities, constraints, and output format. This is your "standing instructions" — things that are true for every request.

2. **Context Block:** The retrieved chunks. This is the evidence the model should reason over. Clearly delineated so the model knows this is grounded evidence, not the query.

3. **User Instruction:** The specific task for this request — generate personality insights and gift suggestions for these specific recipient details.

**The System Prompt for GiftSense**

The system prompt establishes:
- **Role:** "You are a warm, insightful gift recommendation assistant who reads between the lines of conversations to understand people deeply."
- **Grounding rule:** "You MUST only infer personality traits and suggest gifts that are directly supported by evidence in the provided conversation context. Do NOT invent traits or interests not evidenced in the context."
- **Budget rule:** "Every gift suggestion MUST have an estimated price within the stated budget range. If a gift would likely cost more than the upper bound, do not suggest it."
- **Output format:** "Respond ONLY with a valid JSON object matching the schema defined below. No preamble, no explanation outside the JSON."
- **Tone for personality insights:** "Personality insights should be warm, fun, and human — written as if you're the person's most perceptive friend, not a corporate analyst. Use light humor where appropriate."

**The Grounding Constraint — Preventing Hallucination**

Hallucination in RAG occurs when the LLM generates content that is not supported by the retrieved context. In GiftSense, this would look like: "She seems like she would enjoy photography" when nothing in the conversation mentioned photography.

Grounding constraints in the prompt are your primary defense:
- Explicitly instructing the model to cite-or-omit ("only state what the conversation supports")
- Providing a clearly delimited context block ("the following conversation excerpts are your ONLY source of evidence")
- Asking the model to include a brief reason for each suggestion that references the context ("explain why this gift fits, grounded in what you read")

The reason requirement is particularly powerful: by asking the model to justify each suggestion, you force it to have evidence. If it cannot cite a reason from the context, it (ideally) will not make the suggestion.

**Budget Compliance Enforcement**

Budget enforcement requires both prompt-level and output-validation-level enforcement:

*Prompt-level:* Explicitly state the budget range in the system prompt and repeat it in the user instruction. "The user's budget is ₹1000-₹5000. No gift suggestion may have an estimated price outside this range. If you are uncertain about a gift's price, err toward a lower estimate."

*Output-validation-level:* After GPT-4o returns the structured JSON, the Go backend validates each suggestion's `price_range` field against the stated budget. Any suggestion that falls outside the range is filtered out before being returned to the frontend. This defense-in-depth means that even if the LLM violates the budget constraint (which does happen with lower-capability models), the user never sees it.

**Structured Output via JSON Mode**

GPT-4o supports a "JSON mode" that forces the model to respond with valid JSON. The system prompt defines the expected schema explicitly:

```
Output Schema:
{
  "personality_insights": [
    {
      "insight": "string (the observation, written warmly)",
      "evidence_summary": "string (brief paraphrase of what in the conversation supports this)"
    }
  ],
  "gift_suggestions": [
    {
      "name": "string (gift name, specific enough to search for)",
      "reason": "string (why this fits, grounded in the conversation)",
      "estimated_price_inr": "string (e.g., '₹800-₹1200')",
      "category": "string (e.g., 'experience', 'hobby', 'consumable', 'utility')"
    }
  ]
}
```

The `name` field in `gift_suggestions` is particularly important — it is the input to the link generator (Module 10). It should be specific enough to produce useful search results: "Pottery starter kit with air-dry clay" is better than "craft supplies."

**GPT-4o vs. GPT-4o-mini**

| Model | Quality | Cost (input/output per 1M tokens) | Latency |
|---|---|---|---|
| GPT-4o | Excellent | ~$5 / ~$15 | ~2-4s |
| GPT-4o-mini | Very Good | ~$0.15 / ~$0.60 | ~1-2s |

For GiftSense's final completion call, **GPT-4o is recommended** because:
- The quality difference in nuanced personality insight generation is meaningful
- Users are making a significant decision (choosing a gift) and expect quality
- The completion is the only GPT-4o call — all intermediate steps can use mini

**GPT-4o-mini can be used for:** Any preprocessing steps that involve LLM calls (e.g., query expansion, summarization of long chunks before embedding). These are lower-stakes classification tasks where mini's quality is sufficient.

**Token Budget Management**

The assembled prompt has several components with different token costs:
- System prompt: ~300-500 tokens (fixed)
- Retrieved context (12 chunks × ~400 tokens): ~4,800 tokens
- Recipient metadata and user instruction: ~100 tokens
- Expected output (3 insights + 5 suggestions): ~800 tokens

Total per call: approximately 6,000-7,000 tokens. At GPT-4o pricing, this is roughly $0.03-0.04 per session for the completion call alone.

If token costs need to be reduced:
- Reduce top-K to 8 chunks instead of 12: saves ~1,600 tokens
- Truncate long chunks to 300 tokens max: saves variable amount
- Switch to GPT-4o-mini: reduces cost by ~20-30x at some quality cost

---

### Design Decisions
- **GPT-4o for the primary completion call.** Quality matters for the user-facing output.
- **JSON mode** for structured output. Much more reliable than regex parsing of free-text LLM output.
- **Budget enforcement both in prompt and in Go validation layer.** Defense in depth.
- **Reason required per suggestion** to force grounding and enable frontend display.

---

### Alternative Approaches
- **Two separate LLM calls:** One for personality insights, one for gift suggestions. More expensive but allows each call to be independently optimized and retried on failure.
- **Chain-of-thought prompting before final output:** Ask the model to reason step-by-step before generating the JSON. Improves quality but significantly increases token usage and latency.
- **Function calling instead of JSON mode:** OpenAI's function calling feature allows you to define a precise TypeScript-like schema that the model must conform to. More precise than JSON mode, but more complex to implement. A good Phase 2 improvement.

---

### Trade-offs / Consequences
- **A single completion call means a single point of failure.** If the call fails or returns malformed JSON, the user sees an error. Mitigation: retry logic with exponential backoff in the Go HTTP client.
- **JSON mode does not guarantee schema conformance** — it guarantees valid JSON. The model may still produce JSON that does not match the expected schema (extra fields, missing fields). The Go validation layer must handle this gracefully.

---

### How This Module Connects to the Overall System
The GPT-4o response (validated JSON) flows into the Link Generator (Module 10), which enriches each gift suggestion with shopping links before the final response is sent to the frontend.

---

<a name="module-10"></a>
## Module 10 — Budget-Aware Shopping Link Generation

### Objective
Design the link generation layer that transforms GPT-4o's gift suggestion names into actionable, budget-filtered, real shopping links across Amazon India, Flipkart, and Google Shopping — without any web scraping, product APIs, or hardcoded URLs.

---

### Concepts to Learn
- URL construction as a programmatic pattern
- Platform-specific search URL schemas
- Budget range URL encoding for supported platforms
- URL encoding of Indian product names in Go
- Honest disclosure design (links as pre-filtered searches, not guaranteed products)

---

### Detailed Explanation

**The Core Insight**

E-commerce platforms like Amazon and Flipkart expose their search functionality via URL parameters. A search that you perform manually in your browser is entirely encoded in the URL. This means you can construct search URLs programmatically — no API key, no web scraping, no affiliate agreement required. You are simply linking to a pre-configured search results page.

**Amazon India URL Structure**

Amazon India's search URL with price filtering:

```
Base: https://www.amazon.in/s
Parameters:
  k   = search query (URL-encoded)
  rh  = refinement filters
        n:976419031 = all categories
        p_36:{min}00-{max}00 = price range in paise (₹ × 100)
  ref = refinement reference
```

Example for "pottery starter kit" with budget ₹1000-₹5000:
```
https://www.amazon.in/s?k=pottery+starter+kit&rh=p_36%3A100000-500000
```

Note: Amazon prices in their URL filter are in paise (100 paise = ₹1). So ₹1000 = 100000 and ₹5000 = 500000 in the URL parameter.

The Go backend constructs this URL by:
1. Taking the gift name string
2. URL-encoding it for the `k` parameter (handling spaces, special characters, Devanagari if applicable)
3. Calculating the paise values from the budget range
4. Assembling the URL string

**Flipkart URL Structure**

Flipkart's search URL:
```
Base: https://www.flipkart.com/search
Parameters:
  q           = search query (URL-encoded)
  p[]=facets.price_range.from= {min}
  p[]=facets.price_range.to= {max}
```

Example:
```
https://www.flipkart.com/search?q=pottery+starter+kit&p[]=facets.price_range.from%3D1000&p[]=facets.price_range.to%3D5000
```

Flipkart's price range uses actual rupee values, not paise. The URL encoding of the `p[]` array parameters requires careful URL escaping.

**Google Shopping URL Structure**

Google Shopping does not support URL-level price filtering as reliably as Amazon or Flipkart. However, a Google Shopping search URL that includes both the product name and budget context in the query string is the practical solution:

```
Base: https://www.google.com/search
Parameters:
  q   = "{gift_name} under ₹{max_budget}" (URL-encoded)
  tbm = shop
  tbs = vw:l (list view, optional)
```

Example:
```
https://www.google.com/search?q=pottery+starter+kit+under+%E2%82%B95000&tbm=shop
```

The rupee symbol (₹) URL-encodes to `%E2%82%B9`. Google Shopping will surface relevant results and users can apply their own price filter from Google's sidebar, but the query string provides intent context.

**The Budget Range Mapping**

GiftSense uses defined budget tiers (not arbitrary user input), which simplifies URL construction:

| Tier Label | Min (₹) | Max (₹) | Amazon Min (paise) | Amazon Max (paise) |
|---|---|---|---|---|
| Budget | 500 | 1000 | 50000 | 100000 |
| Mid-Range | 1000 | 5000 | 100000 | 500000 |
| Premium | 5000 | 15000 | 500000 | 1500000 |
| Luxury | 15000 | ∞ | 1500000 | (omitted) |

For the Luxury tier (no upper bound), the price filter upper limit is omitted from URLs, allowing Amazon and Flipkart to show all results above the minimum.

**URL Encoding for Indian Product Names**

Product names may include non-ASCII characters (Hindi product names, brand names with special characters). Go's standard library `net/url` package handles this correctly via `url.QueryEscape()` or `url.Values.Encode()`. The key edge cases to handle:
- Spaces → `+` or `%20` (Amazon prefers `+`, Flipkart handles both)
- Rupee symbol (₹ = `%E2%82%B9`)
- Devanagari characters (encoded as multi-byte UTF-8 percent-encoding)

**Honest Disclosure Design**

A critical UX and trust design decision: the shopping links should be presented to the user as "search results links" — not as direct product links. The UI should include clear language: "Tap to see search results on Amazon India filtered to your budget" — not "Buy this on Amazon."

This framing is honest (you cannot guarantee specific product availability), sets correct expectations, and actually provides a better user experience (the user can browse and compare within their budget range).

**Where in the Pipeline Link Generation Happens**

Link generation happens **after** the GPT-4o completion call, as a pure transformation step in Go. The process is:
1. Parse the validated JSON from GPT-4o
2. For each gift suggestion, call the link generator with `(giftName string, budgetTier BudgetRange)`
3. The link generator returns three URLs: Amazon, Flipkart, Google Shopping
4. These URLs are attached to the gift suggestion object
5. The enriched response is sent to the frontend

This is a synchronous, pure-function operation with no API calls, no network I/O. It is essentially string formatting. It is fast and cannot fail (in a meaningful way).

---

### Design Decisions
- **URL construction, no product APIs.** No Amazon Product Advertising API, no Flipkart Affiliate API. These require registration, approval, and have rate limits. URL construction requires nothing and is infinitely available.
- **Google Shopping via query string with budget intent** rather than direct price parameter filtering. Google's URL price filter parameters are less stable and documented than Amazon/Flipkart.
- **Defined budget tiers (not user-entered ranges).** Simplifies URL construction, prevents edge cases (e.g., min > max, non-numeric inputs), and guides users toward sensible ranges.

---

### Alternative Approaches
- **Amazon Product Advertising API:** Returns real product data with prices and availability. Much better user experience (you can show actual products with images). Requires Amazon Associates membership, API approval, and adds latency. A Phase 3+ enhancement.
- **Price comparison APIs (like PriceSpy or JustWatch-equivalent for products):** Not available for Indian e-commerce at a useful free tier.
- **Hardcoded product links per gift category:** Maintains a database of pre-vetted products. Curated and reliable, but requires ongoing maintenance as product listings change. Not viable for a system with thousands of possible gift names.

---

### Trade-offs / Consequences
- **Amazon URL price filter format may change.** Amazon does not officially document their search URL schema. It works in practice but is not guaranteed. Mitigation: the link generation logic should be in a single, easily modifiable module, and the format can be verified and updated if it breaks.
- **Google Shopping link quality depends on query specificity.** A vague gift name ("art supplies") produces a useful search. A very specific name with Indian brand words may produce poor results. The prompt engineering in Module 9 is specifically designed to produce gift names that are "specific enough to search for."

---

### How This Module Connects to the Overall System
The link-enriched gift suggestions are assembled into the final API response object and returned to the React frontend (Module 13), which renders them with clickable links and budget badges.

---

<a name="module-11"></a>
## Module 11 — Session-Scoped Ephemeral Pipeline (Privacy by Design)

### Objective
Design the session management architecture that makes GiftSense's privacy promise a technical guarantee: all conversation data lives only for the duration of a single API request and is irreversibly gone when the request completes.

---

### Concepts to Learn
- Session-scoped vs. user-scoped vs. persistent data models
- Ephemeral architecture as a privacy pattern
- In-memory session management in Go (for non-vector data)
- Pinecone namespace as the session boundary for vectors
- The difference between "we delete your data" (policy) and "we cannot store your data" (architecture)
- What "privacy by design" means at the infrastructure level
- Render free tier's statefulness implications

---

### Detailed Explanation

**The Privacy Guarantee Hierarchy**

There are three levels at which a system can make a privacy claim:

1. **Policy level:** "We have a privacy policy that says we delete your data." This is the weakest guarantee. Policies can be changed, violated, selectively enforced, or superseded by a legal request.

2. **Process level:** "We have automated deletion jobs that run every X hours." Better, but deletion can fail. A crashed deletion job means data persists. Auditing deletion is hard.

3. **Architecture level:** "The system is designed so that data retention beyond the session is technically very difficult and explicitly countered by automatic cleanup." GiftSense operates as close to this level as possible given that Pinecone is an external service.

**How Sessions Work with Pinecone and Multiple Windows**

Each browser window generates a `crypto.randomUUID()` on load. This UUID is the session identity for that window. It is sent with every API call from that window as a field in the request body.

On the backend, this UUID becomes the Pinecone namespace for all vectors produced in that session. Because Pinecone namespaces are isolated within an index, two concurrent windows — even on the same browser, same IP, same user — have completely separate vector spaces. They cannot interfere with each other.

The session lifecycle:

```
Window Opens → UUID generated (client-side, no server involvement)
     │
User submits form
     │
POST /api/analyze { session_id: "abc-123", ... }
     │
Backend creates Pinecone namespace "abc-123" (implicitly via first upsert)
     │
Pipeline runs: parse → anonymize → chunk → embed → upsert → query → complete
     │
Response assembled and returned to frontend
     │
Backend issues DELETE namespace "abc-123" to Pinecone
     │
Request ends — all Go local variables go out of scope and are GC'd
     │
Nothing remains (in Go process memory or in Pinecone)
```

**What "Session-Scoped" Means for Go Process Memory**

The Go process still handles session data in request-scoped memory for most of the pipeline. The anonymized message slice, the chunk objects (with their text), the token map, the retrieval results — all of these live in Go heap memory scoped to the request handler function. Only the embedding vectors and metadata move to Pinecone. When the handler function returns, Go's garbage collector reclaims all request-scoped allocations.

The combination of Go's request-scoped memory + Pinecone namespace deletion means:
- **Conversation text:** Never leaves Go process memory (anonymized form only); GC'd on request end
- **Anonymization token map:** Lives in Go memory; GC'd on request end
- **Embedding vectors:** Live in Pinecone namespace; deleted on request end (or periodic sweep)
- **Chunk metadata (topics, flags):** Lives in Pinecone namespace; deleted on request end

**Multi-Window Data Isolation**

With the in-memory design from the original document, multi-window isolation required explicit session-keyed maps with mutex locking in the Go process. With Pinecone namespaces, isolation is inherent: each window's vectors are in a different namespace and cannot be queried from another session's context.

The Go process itself is still stateless — there is no in-process state shared between requests. Each request's handler function is independent, running in its own goroutine with its own stack.

**What Data Is Not Ephemeral (Be Precise)**

- **Application logs:** Gin's access logs record HTTP metadata (timestamp, endpoint, status, IP, latency). The session UUID will appear in logs (as a request field). The conversation content is NOT in logs. Logs persist per Render's platform behavior (typically a rolling window of recent logs).
- **OpenAI API transit:** Anonymized content transits OpenAI's infrastructure. Under standard API usage policy, OpenAI retains inputs for a limited window for abuse monitoring. See Module 14 for details.
- **Pinecone vectors:** Live in Pinecone infrastructure from first upsert until namespace deletion. Retention window is the duration of the request plus the time for the delete call to complete. In failure scenarios, the periodic sweep handles stragglers within an hour.
- **Frontend browser memory:** The rendered results (insights and gift suggestions) live in the user's browser tab until they close it. Not the backend's concern.

**The Frontend Session Notice (Updated for Multi-Window)**

The React frontend should display a notice that reflects the multi-window reality: "Each conversation you upload is processed privately in its own session and never stored. Closing this tab permanently erases your data." This is slightly more precise than the original — each tab/window is a session, not the entire browser.

---

### Design Decisions
- **Pinecone namespace as the session boundary** for vectors, rather than in-process memory. Enables clean multi-window isolation without Go-level session maps or locking.
- **Go request-scoped memory** for all non-vector session state (conversation text, anonymization map, chunk objects). No database writes for conversation data.
- **Delete-on-complete as the primary cleanup**, with a periodic background sweep as the safety net.
- **Explicit session notice in frontend UI** — the privacy model is communicated per-window, not for the whole browser.

---

### Alternative Approaches
- **Fully in-process session store (original design):** Maximum privacy — nothing leaves Go memory. Multi-window isolation requires explicit mutex-protected session maps. Does not survive instance recycling. Appropriate if Pinecone is unavailable or if privacy requirements are even stricter.
- **Explicit session IDs with server-side state:** Would allow resumable sessions (user closes tab, reopens, session resumes). But requires a session store (Redis, DB), making true ephemerality impossible.
- **Client-side-only processing:** Run anonymizer and chunker in the browser via WebAssembly. Only embeddings and completions go to OpenAI. Maximum privacy — raw conversation never leaves the browser. Complex to implement, and calling OpenAI from the browser exposes the API key (critical security problem). Not viable without a backend proxy.

---

### Trade-offs / Consequences
- **Pinecone as external boundary:** Vectors live outside Go process memory. The privacy story requires trust that Pinecone does not log or retain vector data beyond the namespace's existence. Mitigation: only anonymized data ever reaches Pinecone; vectors cannot be reverse-engineered to recover conversation text.
- **Cleanup can fail.** If Render spins down the instance before the namespace delete call completes, that namespace persists in Pinecone until the periodic sweep runs. The sweep handles this within an hour. This is an acceptable residual risk for a session-based consumer product.
- **Multi-window is supported, but concurrent heavy usage adds Pinecone API call volume.** 50 concurrent sessions = 50 sets of namespace operations. The free tier's API call limits are generous enough that this is not a concern at the expected scale of a learning/portfolio project.

---

### How This Module Connects to the Overall System
The session-scoped architecture wraps the entire pipeline from Module 3 through Module 10. It is the "outer container" that governs data lifetime. All pipeline stages in this system operate within the session's lifetime.

---

<a name="module-12"></a>
## Module 12 — Go Backend Architecture — Clean Architecture with Gin

### Objective
Design the structural architecture of the Go backend — how code is organized, how dependencies flow, how the OpenAI client is wrapped for testability, and how idiomatic Go patterns apply to a RAG system.

---

### Concepts to Learn
- Clean Architecture (or Ports & Adapters) in Go
- Dependency inversion via interfaces
- The Go interface for abstraction without inheritance
- Wrapping third-party clients for testability and swappability
- Gin middleware design for cross-cutting concerns
- Error handling as a first-class concern in Go

---

### Detailed Explanation

**The Core Structural Principle: Dependency Inversion**

The key design rule for the Go backend: **business logic (the RAG pipeline) should not depend on infrastructure (OpenAI, Gin, file system)**. Infrastructure should depend on business logic interfaces.

This is achieved via Go interfaces. Instead of the Chunker directly calling the OpenAI embedding API, it calls a `Embedder` interface. The production implementation of that interface calls OpenAI. A test implementation returns hardcoded vectors. The Chunker never knows which implementation it is talking to.

**Project Structure**

```
giftsense-backend/
├── cmd/
│   └── server/
│       └── main.go              // Entry point: wires everything together
│                                // Reads config, constructs all adapters,
│                                // injects into use cases, starts Gin router
│
├── internal/
│   ├── domain/                  // Pure domain types — no dependencies
│   │   ├── conversation.go      // Message, Chunk, Session types
│   │   ├── recipient.go         // RecipientDetails, BudgetRange types
│   │   └── suggestion.go        // GiftSuggestion, PersonalityInsight types
│   │
│   ├── usecase/                 // Business logic — depends only on domain + ports
│   │   ├── analyze.go           // AnalyzeConversation use case (orchestrates pipeline)
│   │   ├── parse.go             // Conversation parsing logic
│   │   ├── anonymize.go         // Anonymization logic
│   │   ├── chunk.go             // Chunking logic
│   │   └── retrieve.go          // Retrieval and query construction logic
│   │
│   ├── port/                    // Interface definitions (ports)
│   │   ├── embedder.go          // Embedder interface
│   │   ├── llm.go               // LLMClient interface
│   │   └── vectorstore.go       // VectorStore interface (Add, Query, DeleteSession)
│   │
│   ├── adapter/                 // Implementations of ports (adapters)
│   │   ├── openai/
│   │   │   ├── embedder.go      // OpenAI implementation of Embedder
│   │   │   │                    // Uses cfg.EmbeddingModel, cfg.EmbeddingDimensions
│   │   │   └── llm.go           // OpenAI implementation of LLMClient
│   │   │                        // Uses cfg.ChatModel, cfg.MaxTokens
│   │   ├── vectorstore/
│   │   │   └── pinecone.go      // Pinecone implementation of VectorStore
│   │   │                        // Uses cfg.PineconeAPIKey, cfg.PineconeIndexName
│   │   │                        // Handles: Upsert, Query (with metadata filter), DeleteNamespace
│   │   └── linkgen/
│   │       └── shopping.go      // Shopping link generation
│   │
│   └── delivery/                // HTTP layer — depends on usecases via interfaces
│       ├── http/
│       │   ├── handler.go       // Gin handlers
│       │   │                    // Sets Gin's MaxMultipartMemory to cfg.MaxFileSizeBytes
│       │   ├── middleware.go    // CORS, request logging, rate limiting
│       │   └── validator.go     // Request validation (file size, UUID format, fields)
│       └── dto/
│           ├── request.go       // Request DTOs (AnalyzeRequest — includes SessionID)
│           └── response.go      // Response DTOs (AnalysisResponse)
│
└── config/
    └── config.go                // Reads and validates ALL env vars at startup
                                 // Fails fast if required vars are missing
                                 // Exposes a typed Config struct used throughout
```

**The Config Struct — Central Nervous System**

The `config.Config` struct is loaded once at startup in `main.go` and injected into every adapter and use case that needs it. No component reads `os.Getenv()` directly — only `config.go` does. This creates a single, auditable place where all configuration is validated at startup.

```
Config {
    // OpenAI
    OpenAIAPIKey:        string     // OPENAI_API_KEY (required)
    ChatModel:           string     // CHAT_MODEL (default: "gpt-4o")
    EmbeddingModel:      string     // EMBEDDING_MODEL (default: "text-embedding-3-small")
    EmbeddingDimensions: int        // EMBEDDING_DIMENSIONS (default: 1536)
    MaxTokens:           int        // MAX_TOKENS (default: 1000)

    // Retrieval
    TopK:                int        // TOP_K (default: 3) — results per query
    NumRetrievalQueries: int        // NUM_RETRIEVAL_QUERIES (default: 4)

    // Pinecone
    PineconeAPIKey:      string     // PINECONE_API_KEY (required)
    PineconeIndexName:   string     // PINECONE_INDEX_NAME (default: "giftsense")
    PineconeEnvironment: string     // PINECONE_ENVIRONMENT (e.g., "us-east-1")

    // Pipeline
    MaxFileSizeBytes:    int64      // MAX_FILE_SIZE_BYTES (default: 2097152 = 2MB)
    MaxProcessedMessages: int       // MAX_PROCESSED_MESSAGES (default: 400)
    ChunkWindowSize:     int        // CHUNK_WINDOW_SIZE (default: 8)
    ChunkOverlapSize:    int        // CHUNK_OVERLAP_SIZE (default: 3)

    // Server
    Port:                string     // PORT (set by Render automatically)
    AllowedOrigins:      []string   // ALLOWED_ORIGINS (comma-separated frontend URLs)
}
```

**The VectorStore Interface — Updated for Pinecone**

The `VectorStore` interface gains a `DeleteSession` method to accommodate the Pinecone namespace lifecycle:

```
VectorStore interface {
    Upsert(sessionID string, chunks []Chunk, vectors [][]float32) error
    Query(sessionID string, queryVector []float32, topK int, filter MetadataFilter) ([]Chunk, error)
    DeleteSession(sessionID string) error
}
```

The Pinecone adapter implements all three. The original in-memory adapter (kept for testing) implements `DeleteSession` as a no-op. This means the business logic in the use case can call `DeleteSession` unconditionally without knowing which implementation is in use.

**The AnalyzeConversation Use Case**

The `AnalyzeConversation` use case in `usecase/analyze.go` is the orchestrator. It coordinates the entire pipeline:
1. Parse the conversation (calls the parser)
2. Anonymize (calls the anonymizer, receives the anonymized messages + token map)
3. Chunk (calls the chunker with anonymized messages)
4. Embed chunks (calls the Embedder interface → OpenAI)
5. Populate the VectorStore (calls the VectorStore interface)
6. Construct retrieval queries
7. Embed queries (calls the Embedder interface again)
8. Retrieve top-K chunks (calls the VectorStore interface)
9. Build the GPT-4o prompt
10. Call GPT-4o (calls the LLMClient interface → OpenAI)
11. Validate and parse the JSON response
12. Generate shopping links (calls the link generator)
13. Return the final AnalysisResult

This is a long sequential pipeline. In Go, it is implemented as a function that calls each stage and checks errors explicitly after each call. There are no hidden exceptions — if any stage fails, the error is returned to the Gin handler, which returns an appropriate HTTP error to the frontend.

**Concurrent Embedding with Goroutines**

Embedding 20-30 chunks sequentially (one API call at a time) would take 20-30 API round trips. These can be parallelized. Go's goroutines and `sync.WaitGroup` (or an `errgroup`) allow all chunk embeddings to be issued concurrently, with the results collected once all goroutines complete. In practice, this reduces embedding latency from ~5-10 seconds to ~0.5-1 second.

**Gin Middleware**

Gin middleware handles cross-cutting concerns that should not live in business logic:
- **CORS middleware:** Allows the React frontend (different origin on Render) to call the backend API
- **Request logging:** Logs request method, path, status, and latency — NOT request body (never log conversation content)
- **Rate limiting:** Prevents abuse on the free tier. A simple token bucket limiter per IP address, implemented in-process (no Redis needed for simple rate limiting)
- **API key injection:** The OpenAI API key is read from the environment variable at startup and injected into the OpenAI adapter constructor — never passed in a request header or hardcoded in source

**Error Handling Philosophy in Go**

Go handles errors explicitly — every function that can fail returns an error value. The RAG pipeline has many failure points (parsing failures, anonymization failures, OpenAI API errors, malformed JSON from GPT-4o). Each error should be:
1. Checked immediately after the call that produced it
2. Wrapped with context: `fmt.Errorf("chunking stage failed: %w", err)`
3. Propagated up to the Gin handler, which maps it to an appropriate HTTP status
4. Logged at the appropriate level (debug for expected errors, error for unexpected ones)

Never log the conversation content, never log the API key, never log PII. The error context should be structural (stage name, error type) not content (what was in the conversation).

---

### Design Decisions
- **Clean Architecture layers** (domain, usecase, port, adapter, delivery). This structure makes each layer independently testable and replaceable.
- **Interfaces for all external dependencies** (OpenAI, vector store). This is standard Go practice and essential for a maintainable, testable codebase.
- **No global state for session data.** Everything is passed through function parameters.

---

### Alternative Approaches
- **Flat structure (everything in one package):** Faster to start but becomes unmaintainable as the system grows. Mixing HTTP handling, business logic, and OpenAI calls in one file is a common beginner pattern to avoid.
- **gRPC instead of REST/Gin:** More performant for internal microservice communication, but overkill for a monolithic backend. Gin with REST is appropriate here.
- **Dependency injection framework (like Wire or fx):** Automates the wiring of dependencies. Useful at larger scales, adds complexity for a small project. Manual dependency injection in `main.go` is simpler and more legible for this size.

---

### Trade-offs / Consequences
- **Clean Architecture adds initial setup overhead.** Creating interface files before implementing them feels like extra work upfront. The payoff comes when you want to swap the OpenAI embedder for a different provider, or when you write tests using mock implementations.
- **Explicit error propagation is verbose** but makes failure modes visible. In a RAG pipeline with 10+ stages, knowing exactly which stage failed and why is invaluable for debugging.

---

### How This Module Connects to the Overall System
This module defines the structural skeleton that every other module's code lives inside. Modules 3-10 describe *what* each pipeline stage does; this module describes *where it lives* and *how stages communicate* in the Go codebase.

---

<a name="module-13"></a>
## Module 13 — React Frontend Architecture

### Objective
Design the React frontend — a lean, performant, mobile-first single-page application that handles conversation upload, recipient details collection, and result display — with support for multiple independent browser windows, clear privacy communication, and a UI that works as well on a phone as on a desktop.

---

### Concepts to Learn
- Single-page application (SPA) state management for a multi-step flow
- Mobile-first responsive design principles
- Client-side session UUID generation for multi-window isolation
- File upload UX patterns on mobile and desktop
- Optimistic UI and loading state management
- Communicating privacy guarantees in UI design
- Lean React without heavy frameworks for performance on free-tier hosting

---

### Detailed Explanation

**Application Flow and State**

The frontend has a simple three-screen flow:

1. **Input Screen:** Recipient details form + conversation upload
2. **Loading Screen:** Processing animation while the backend works
3. **Results Screen:** Personality insights + gift suggestions with shopping links

The entire application state can be managed with React's built-in `useState` and `useReducer` hooks — no external state management library (Redux, Zustand) is needed. The state object has three fields: `formData` (recipient details), `uploadedConversation` (the text content), and `results` (the API response or null).

**Session UUID Generation (Multi-Window Support)**

When the React app mounts for the first time in a tab or window, it generates a session UUID using the browser's built-in `crypto.randomUUID()`. This UUID is stored in a `useRef` (not state — it should never change during the session and should not trigger re-renders). It is attached to every API call in the request body.

Each window gets its own UUID because each window runs its own React app instance with its own independent component lifecycle. Opening 3 tabs = 3 independent UUIDs = 3 independent Pinecone namespaces on the backend. There is zero coordination required between tabs.

The frontend should not display this UUID to users. It is an internal implementation detail, not a user-visible concept.

**Mobile-First Responsive Design**

GiftSense's primary audience is likely mobile — users who receive WhatsApp conversations on their phones and want to look up gift ideas. The UI must be fully functional on a 375px-wide phone screen without horizontal scrolling.

Mobile-first design principles applied throughout:
- **Layout:** Single-column layout on mobile, optional two-column on tablet/desktop (input form + upload side-by-side at ≥768px breakpoint)
- **Touch targets:** All interactive elements (buttons, selector cards, upload zone) must be at least 44×44px — Apple's minimum tap target guideline
- **Font sizes:** Minimum 16px for body text on mobile to prevent iOS auto-zoom on input focus
- **Budget selector cards:** Stack vertically on mobile (full-width), 2-column grid on tablet, 4-column row on desktop
- **Gift suggestion cards:** Full-width on mobile, 2-column grid on tablet/desktop
- **Shopping link buttons:** Stack vertically on mobile (full-width buttons), horizontal row on desktop

CSS approach: Tailwind CSS utility classes, which are tree-shaken at build time. Use `sm:`, `md:`, `lg:` breakpoint prefixes exclusively for layout adjustments. Core styling (colors, spacing, typography) uses Tailwind's default scale.

**No Heavy UI Frameworks**

Render's free tier serves the React app as a static site (pre-built files). The concern is the JavaScript bundle size — large bundles mean slow initial load for users on mobile connections with variable data speeds.

Avoid heavy component libraries (Material UI, Ant Design, Chakra) — each adds 100-300KB to the bundle. Use Tailwind CSS (utility classes, tree-shaken at build time = small CSS output). Use only the specific Lucide icons you need (tree-shaken by the bundler) rather than importing an entire icon library.

**The Upload Component (Mobile-First)**

The conversation upload component must work well for both mobile and desktop users:

*Desktop path:*
- Drag-and-drop zone: large, visually inviting drop target
- Click-to-upload: triggers hidden `<input type="file">` (accepts `.txt` only, `max 2MB`)
- Text paste area: collapsible `<textarea>` for pasting conversation text

*Mobile path:*
- Large tap target that opens the device file picker (no drag-and-drop on mobile — feature-detect and hide the drop zone on touch devices)
- The file picker on iOS/Android will show the Files app where users can navigate to their `.txt` export
- Text paste area: full-width textarea that works well with mobile keyboard, `autocapitalize="off"` to prevent iOS from capitalizing chat message text

**File size display and validation feedback:**
- Show the selected file name and size before submission
- If file exceeds 2MB, display an inline error immediately (before API call): "This file is too large. Maximum size is 2MB."
- If file is too small (< 500 bytes), display: "This file seems too short. Please export a full conversation."

No file should be stored in the browser's local storage or IndexedDB — the file content lives only in React state (in memory) for the duration of the session.

**The Budget Range Selector**

Budget is a defined-tier selector (not a free-text input). On mobile, the tiers stack as full-width selectable cards. On desktop, they display as a 4-column card row. The selected tier gets a visual highlight (border, background change). Tapping a different tier deselects the current one. This maps directly to the backend's budget tier enum.

**Privacy Notice Design**

Two privacy notices are required, both sized for mobile readability:

1. **Pre-upload notice** (on the Input Screen): A calm, visible banner stating "This conversation is processed privately in this tab and never stored." Displayed below the upload component, not as a modal. On mobile, this should be compact — 2 lines maximum.

2. **Post-results notice** (on the Results Screen): A smaller, dismissible notice at the top of results: "Your conversation has been processed and permanently deleted from our server." Dismissible to maximize screen real estate on mobile.

**Results Display (Mobile-Optimized)**

Personality insights are displayed as horizontally scrollable cards on mobile (a swipeable carousel) and as a wrapping grid on desktop. Each insight card is compact — the observation text and evidence summary should fit on a phone screen without scrolling within the card.

Each gift suggestion is a card with:
- Gift name (prominent, 18-20px)
- Reason why it fits (lighter style, 14-16px)
- Budget badge (pill-shaped, shows estimated price range)
- Three shopping link buttons:
  - On mobile: full-width stacked buttons (Amazon, Flipkart, Google Shopping), each with platform icon + label
  - On desktop: horizontal row of buttons
  - Each opens in a new tab (`target="_blank" rel="noopener noreferrer"`)
  - Small disclaimer below: "These links open filtered search results. Product availability is not guaranteed."

**Loading State Design**

The backend pipeline takes 3-8 seconds (embedding calls + GPT-4o completion). The loading screen should:
- Show a meaningful progress animation sized for mobile (a centered animation, not a full-screen overlay that obscures the form)
- Display rotating text: "Reading your conversation...", "Finding patterns...", "Crafting suggestions..."
- On mobile, ensure the loading animation does not cause layout shifts when it appears

**Error Handling in the UI**

Possible error states the frontend must handle:
- **Network error / cold start:** "Our server is warming up — please try again in a moment." With a retry button.
- **File too large (client-side):** Immediate inline validation before any API call
- **Validation error (from backend):** Inline field errors with clear copy
- **Backend processing error:** "Something went wrong — please try again"
- **Timeout:** "This is taking longer than usual — you can wait or try again"

All error messages should be readable on a mobile screen without truncation.

**Accessibility on Mobile**

- All form inputs must have visible labels (not just placeholders — placeholders disappear on focus)
- Error messages must be associated with their inputs via `aria-describedby` for screen reader support
- Color is not the only indicator of state (selected budget tier uses both color AND a checkmark icon)
- Sufficient color contrast for outdoor mobile use (WCAG AA minimum)

---

### Design Decisions
- **Built-in React state management** (no Redux/Zustand). The state is simple enough — three screens, a form, and an API response.
- **No heavy component library.** Performance > convenience for a free-tier deployment with mobile users on variable data connections.
- **Mobile-first layout via Tailwind CSS responsive utilities.** Write the mobile layout first; layer desktop enhancements on top via `sm:` and `md:` breakpoints.
- **`crypto.randomUUID()` for session identity.** Client-side UUID generation is immediate, requires no server round-trip, and is unique per window/tab.
- **Privacy notice as a first-class, mobile-compact UI element**, not legal copy hidden in the footer.

---

### Alternative Approaches
- **Next.js instead of plain React:** Server-side rendering would improve initial load and SEO. But GiftSense is not SEO-driven, and Next.js on Render free tier requires a Node.js server process (another service slot). Plain React deployed as a static site is simpler and faster.
- **Progressive Web App (PWA) with Add to Home Screen:** Would give mobile users an app-like experience. Service workers add complexity without clear benefit for a stateless, session-only app.
- **React Native / Expo for a native mobile app:** Maximum mobile UX. Overkill for a learning project, adds significant build and deployment complexity.

---

### Trade-offs / Consequences
- **No progress events** from the backend means the loading animation is cosmetic. Users with slow connections or on cold-start servers may feel uncertain. A WebSocket or Server-Sent Events (SSE) progress stream would be a meaningful Phase 3 improvement.
- **No result persistence in the browser** means refreshing the results page loses everything. Acceptable given the privacy model; users on mobile can screenshot or share results before leaving.
- **Multi-window is transparent to the user.** Users do not see session IDs or namespace information. From their perspective, each tab is just an independent instance of the app — which is the correct mental model.

---

### How This Module Connects to the Overall System
The React frontend is the user-facing skin over the entire backend pipeline. It collects input (feeding Module 3), triggers the analysis (the entire pipeline), and renders the output (from Module 9 and Module 10).

---

<a name="module-14"></a>
## Module 14 — OpenAI Cost Model & Optimization Strategy

### Objective
Build a concrete, numeric understanding of what one GiftSense user session costs in OpenAI API fees — broken down by stage — and identify where optimization delivers the highest return.

---

### Concepts to Learn
- Token-based pricing and how to calculate token counts
- Embedding cost vs. completion cost relative magnitudes
- The GPT-4o vs. GPT-4o-mini cost trade-off in context
- Cost per session vs. cost at scale (100 users/day)
- Caching strategies for cost reduction
- What OpenAI's data retention policy means for this project's privacy model

---

### Detailed Explanation

**OpenAI Pricing Reference (approximate, verify current pricing at platform.openai.com)**

| Model | Input per 1M tokens | Output per 1M tokens |
|---|---|---|
| GPT-4o | ~$5.00 | ~$15.00 |
| GPT-4o-mini | ~$0.15 | ~$0.60 |
| text-embedding-3-small | ~$0.02 per 1M tokens | — |

**Cost Breakdown per User Session**

*Stage 1 — Chunk Embedding (Index Time):*
- ~25 chunks × ~400 tokens/chunk = ~10,000 tokens
- Cost: 10,000 / 1,000,000 × $0.02 = **$0.0002**

*Stage 2 — Query Embedding (Retrieval Time):*
- 4 retrieval queries × ~50 tokens/query = ~200 tokens
- Cost: 200 / 1,000,000 × $0.02 = **<$0.000005** (negligible)

*Stage 3 — GPT-4o Completion:*
- Input tokens: system prompt (~400) + retrieved context (~4,800) + user instruction (~100) = ~5,300 tokens
- Output tokens: 3 insights + 5 suggestions + reasons = ~600 tokens
- Input cost: 5,300 / 1,000,000 × $5.00 = **$0.0265**
- Output cost: 600 / 1,000,000 × $15.00 = **$0.009**
- Total completion cost: **~$0.036**

*Total per session: ~$0.036 (dominated by GPT-4o completion)*

**Cost at Scale**

| Daily users | Monthly cost (GPT-4o) | Monthly cost (GPT-4o-mini) |
|---|---|---|
| 10 | ~$11 | ~$0.50 |
| 100 | ~$108 | ~$5 |
| 1,000 | ~$1,080 | ~$50 |

**Where Is the Money Going?**

96% of the per-session cost is in the GPT-4o completion call. Embeddings are almost free by comparison. This means the primary cost lever is: **which model do you use for completion?**

If GPT-4o-mini is used for the completion call instead of GPT-4o, cost drops from ~$0.036 to ~$0.0014 per session — a 25x reduction. Quality drops too, but for many sessions the quality difference is acceptable.

**The GPT-4o vs. GPT-4o-mini Decision Framework**

Use GPT-4o when:
- The conversation is rich and complex (long, multi-topic)
- The output quality directly affects the user's gift-buying decision
- The personality insights need to be nuanced and specific (not generic)

Use GPT-4o-mini when:
- Any intermediate classification or preprocessing LLM calls are added (e.g., query expansion, chunk summarization, intent detection) — these are lower-stakes tasks
- The conversation is short and simple (mini performs well on simple extractions)
- You are in a development/testing environment and do not need production quality

**A Tiered Model Strategy**

A production-quality cost optimization: use a two-tier model:
1. Run a fast GPT-4o-mini call first to assess conversation complexity and richness
2. Based on the assessment (simple vs. complex), route to GPT-4o or GPT-4o-mini for the full completion
3. Simple conversations (short exchanges, minimal signal) go to mini; rich, complex conversations go to 4o

This is a meaningful Phase 4 optimization — do not implement it in Phase 1.

**Caching Strategies**

*Embedding cache:* Once a chunk is embedded, its vector can theoretically be cached by chunk content hash. In practice, every session has unique content, so embedding cache hit rates are near zero. Not worth implementing.

*Completion cache:* Similarly, every session produces unique context. Completion caching is not applicable.

*What can be cached:* If certain gift queries are common (e.g., "birthday gift for mom, ₹1000-₹5000"), you could cache the top-10 gift suggestions for that exact context. But this undermines the core value proposition of personalization. Not recommended.

The honest conclusion: there is no meaningful caching opportunity for GiftSense. Each session is unique by design. Cost optimization comes from model selection, token reduction (smaller chunks, fewer retrieved chunks), and conversation length limits.

**OpenAI Data Retention Policy**

OpenAI's API (as of recent policy) states that inputs to the API are not used to train models by default. They retain inputs and outputs for a limited time for abuse monitoring (typically 30 days for standard API access), but this can be reduced to zero under a Zero Data Retention (ZDR) agreement, which is available to enterprise customers.

For GiftSense on a standard API key: OpenAI retains the API inputs (anonymized chunks, prompts) for up to 30 days for abuse monitoring. This is why anonymization is non-negotiable — even under this 30-day retention window, no PII exists in what OpenAI retains.

For the project's privacy disclosure, the honest statement is: "Your conversation is anonymized before any processing. Anonymized text excerpts are processed by OpenAI's API and may be retained by OpenAI for up to 30 days per their data policy. No personally identifiable information is included in what OpenAI receives."

---

### Design Decisions
- **GPT-4o for the primary completion call** in production. The quality of personality insights and gift suggestions is the core value proposition — this is not where to cut costs at launch.
- **GPT-4o-mini for any intermediate LLM steps** that are added in later phases.
- **Hard conversation length cap** as the primary cost control, not model downgrading.

---

### Alternative Approaches
- **Prompt compression:** Use a tool like LLMLingua to compress the retrieved context before sending to GPT-4o, reducing token count. Reduces cost by 30-50%, may reduce quality slightly. A valid Phase 3 optimization.
- **Batching multiple sessions:** If multiple users submit sessions around the same time, their embedding calls could be batched into a single OpenAI embedding request (which supports multiple inputs). Reduces API call overhead but adds latency complexity. Not worth it at early scale.

---

### Trade-offs / Consequences
- **GPT-4o cost at scale is real.** At 1,000 users/day, the monthly OpenAI bill is ~$1,080. This is sustainable for a paid product but not for a free one. The free-tier deployment on Render does not change the OpenAI cost.
- **No caching opportunity** means cost scales linearly with users. There is no "economy of scale" in the LLM call cost for GiftSense.

---

### How This Module Connects to the Overall System
The cost model informs which models to use (Modules 6, 9), how many chunks to retrieve (Module 8), and how to set conversation length limits (Module 3). It is a constraint that propagates through the entire pipeline design.

---

<a name="module-15"></a>
## Module 15 — Render Free Tier Deployment Architecture

### Objective
Design the deployment configuration for GiftSense on Render's free tier — understanding the constraints, designing around them rather than against them, and producing a reliable deployment that handles cold starts gracefully.

---

### Concepts to Learn
- Render's service types (Web Service vs. Static Site)
- Free tier constraints (spin-down, RAM, CPU, egress limits)
- Cold start latency and mitigation strategies
- Environment variable management for secrets
- Static site deployment for React
- Go binary deployment and build configuration

---

### Detailed Explanation

**Render Free Tier Architecture**

GiftSense uses two Render services:

*Service 1 — Backend: Render Web Service (Free Tier)*
- Runtime: Go binary
- Spin-down behavior: spins down after 15 minutes of inactivity
- RAM: 512MB
- CPU: shared, throttled
- Egress: no hard limit but throttled under high load
- Cost: $0

*Service 2 — Frontend: Render Static Site (Free Tier)*
- Hosting: CDN-served static files
- No spin-down (static files are always available)
- Unlimited bandwidth for typical loads
- Cost: $0

**The Cold Start Problem**

When the Render Web Service (Go backend) has been idle for 15+ minutes, it spins down. The next request wakes it up, but startup takes 5-30 seconds (loading the Go binary, initializing the Gin router, connecting to the OpenAI client). During this time, the frontend shows a loading state.

Without mitigation, a user who lands on GiftSense after a period of inactivity will submit their form and wait up to 30 seconds before anything happens — with no feedback that anything is working.

**Cold Start Mitigation Strategies**

*Strategy 1 — Frontend ping mechanism:* The React frontend, on initial page load, sends a lightweight "ping" request to the backend health check endpoint (`GET /health`). If the backend is asleep, this ping wakes it up. The form is interactive while the backend warms up in the background. By the time the user fills out the form (typically 30-60 seconds), the backend is warmed up. This is a free mitigation that works well.

*Strategy 2 — User-facing messaging:* The loading screen includes handling for slow responses. If the API call has not responded in 8 seconds, the frontend shows: "Our server is waking up — this may take up to 30 seconds on first use. Thank you for your patience." Honest communication is better than an apparent hang.

*Strategy 3 — UptimeRobot free monitoring:* UptimeRobot (free tier) pings a URL every 5 minutes. Configure it to ping the backend health endpoint. This keeps the service alive during off-hours — at the cost of ~288 pings/day of compute time on Render. Render's free tier may not allow this (their terms may prohibit keeping services artificially awake). Check current Render terms before implementing.

**Go Binary Optimization for Render**

Render builds the Go binary during deployment using the Go toolchain. The build configuration:
- Use `CGO_ENABLED=0` — disable CGO for a fully static binary with no glibc dependency. This ensures the binary runs on any Linux environment.
- Use `-ldflags="-s -w"` — strip debug symbols from the binary, reducing binary size by 30-50%.
- Set `GOOS=linux GOARCH=amd64` — explicitly target the Render host architecture.

The resulting binary starts in < 1 second. The cold start latency is Render's infrastructure spin-up time, not Go startup time.

**Environment Variable Management**

All runtime configuration is managed through environment variables — never hardcoded. Render provides an environment variables management interface in the dashboard. The full set of variables GiftSense reads at startup:

**Required — application fails to start without these:**
- `OPENAI_API_KEY` — the OpenAI API key. Set as a secret env var in Render dashboard.
- `PINECONE_API_KEY` — the Pinecone API key. Set as a secret env var in Render dashboard.

**Configurable — have sensible defaults but should be explicitly set:**
- `CHAT_MODEL` — OpenAI chat completion model. Default: `gpt-4o`. Set to `gpt-4o-mini` for cost reduction.
- `EMBEDDING_MODEL` — OpenAI embedding model. Default: `text-embedding-3-small`.
- `EMBEDDING_DIMENSIONS` — vector dimensions. Default: `1536` (matches `text-embedding-3-small`). Must match Pinecone index configuration.
- `MAX_TOKENS` — maximum output tokens for chat completions. Default: `1000`.
- `TOP_K` — number of nearest-neighbor results per retrieval query. Default: `3`.
- `NUM_RETRIEVAL_QUERIES` — number of parallel retrieval queries issued per session. Default: `4`.
- `MAX_FILE_SIZE_BYTES` — maximum `.txt` upload size. Default: `2097152` (2MB). Adjust downward to reduce costs per session.
- `MAX_PROCESSED_MESSAGES` — maximum messages fed into chunking/embedding after sampling. Default: `400`.
- `CHUNK_WINDOW_SIZE` — sliding window size for chunking. Default: `8`.
- `CHUNK_OVERLAP_SIZE` — overlap between consecutive windows. Default: `3`.
- `PINECONE_INDEX_NAME` — Pinecone index name. Default: `giftsense`. Must match the index you created in the Pinecone console.
- `PINECONE_ENVIRONMENT` — Pinecone cloud region. Set to the region of your free-tier index (e.g., `us-east-1`).
- `ALLOWED_ORIGINS` — comma-separated list of allowed frontend origins for CORS. Set to the Render static site URL: `https://giftsense.onrender.com`.
- `PORT` — set automatically by Render. The Go app reads this via `os.Getenv("PORT")`.

**Why Full Configurability via Env Vars Matters**

This design means you can tune every important parameter of the RAG pipeline — model choice, token budget, chunk size, retrieval depth — without touching or redeploying code. For a learning project, this is essential: you will want to experiment with `TOP_K=5` vs `TOP_K=3`, or compare `gpt-4o` vs `gpt-4o-mini` output quality, by simply changing an env var and restarting the service. No code change, no git commit, no rebuild required.

Render allows env var changes to trigger an automatic redeploy. The Go application's fast startup time (< 1 second for the binary) means these changes take effect in under 30 seconds.

**React Build and Static Site Configuration**

The React app is built with Vite (or Create React App, though Vite is preferred for faster builds and smaller output). The build output (`dist/` folder) is deployed as a Render Static Site.

The backend API URL is injected as an environment variable at build time: `VITE_API_URL=https://giftsense-backend.onrender.com`. The frontend code reads this from `import.meta.env.VITE_API_URL`.

**CORS Configuration**

The Go backend must allow cross-origin requests from the Render static site domain. The CORS middleware allows:
- Origin: `https://giftsense.onrender.com` (the frontend's static site URL)
- Methods: `POST, GET, OPTIONS`
- Headers: `Content-Type`

In development, CORS is configured to allow `http://localhost:3000` (local React dev server).

**What NOT to Do on Render Free Tier**

- **Do not use persistent disk.** Render free tier does not provide persistent disk storage. Any file written to disk is lost on redeploy or spin-down. All session data must be in-memory (which GiftSense already ensures).
- **Do not spawn long-running background jobs.** Render free tier has CPU throttling. A background job that constantly consumes CPU will slow request handling.
- **Do not use more than one free Web Service for a monolithic backend.** Free tier allows one Web Service — keep the Go backend as a single service.
- **Do not assume low latency from free tier.** API response times on Render free tier are higher than on paid tiers due to shared CPU. Design timeouts and user messaging accordingly.
- **Do not commit `.env` files or `render.yaml` with secrets.** Use Render's dashboard for secrets.

---

### Design Decisions
- **Frontend ping on page load** as the primary cold start mitigation — low complexity, high effectiveness.
- **Static binary with stripped debug symbols** for minimal Go binary size and fastest cold start.
- **Environment variables via Render dashboard** — never in code, never in git.

---

### Alternative Approaches
- **Fly.io free tier instead of Render:** Fly.io's free tier allows machines that sleep but wake faster than Render (~2 seconds vs. ~10-30 seconds). The cold start experience is meaningfully better. However, Fly.io is more complex to configure than Render, and Render's simplicity is more appropriate for a learning project.
- **Railway free tier:** Similar to Render, with similar constraints. Less mature tooling.
- **Vercel for frontend (instead of Render Static Site):** Vercel has excellent static site hosting and edge CDN. However, using two different deployment platforms adds complexity. Render's static site is sufficient for this use case.

---

### Trade-offs / Consequences
- **Free tier cold starts are a real UX problem.** The ping mitigation helps but does not eliminate the issue entirely if the service is cold during the ping itself. For a production commercial product, a paid Render instance (which never spins down) is worth the cost.
- **512MB RAM limits concurrency.** Each concurrent session uses ~180KB for vector data + overhead. Theoretical maximum concurrent sessions: ~200-300. In practice, CPU is the binding constraint before RAM.

---

### How This Module Connects to the Overall System
Deployment is the final wrapper around the entire system. The constraints of Render free tier have shaped architectural decisions throughout this document — from the Pinecone session-namespace design (Module 7) to the ephemeral session model (Module 11) to the frontend cold start handling (Module 13).

---

<a name="module-16"></a>
## Module 16 — Observability, Failure Modes & Production Hardening

### Objective
Design the observability and resilience layer — understanding how to monitor the system, detect failures, and harden the pipeline against the most common failure modes, all within the constraints of zero-cost tooling.

---

### Concepts to Learn
- Observability vs. monitoring: the distinction matters
- The three pillars of observability (logs, metrics, traces)
- RAG-specific failure modes
- Retry logic and circuit breakers for external API calls
- Zero-cost observability tooling
- What to measure in a RAG system

---

### Detailed Explanation

**The Three Pillars Applied to GiftSense**

*Logs:* Structured JSON logs (using Go's `slog` package, available in Go 1.21+) for every pipeline stage. Each log entry includes: session ID (a random UUID per request, not tied to user identity), stage name, duration in ms, token count (for OpenAI calls), error message (if any). Never log conversation content, chunk text, or PII.

Log levels:
- `INFO`: Normal operation events (session started, stage completed, session completed)
- `WARN`: Degraded but recoverable events (OpenAI retry needed, chunk count below minimum, budget filter removed a suggestion)
- `ERROR`: Failed requests (OpenAI API error, JSON parsing failure, validation error)

*Metrics:* On Render free tier, there is no metric collection service available without external tooling. Use simple in-process counters (Go atomic integers in a package-level variable):
- `totalSessions`: total sessions processed since startup
- `successfulSessions`: sessions that returned a result
- `openAIErrors`: OpenAI API errors encountered
- `averageLatencyMs`: running average of end-to-end request latency

Expose these at a `GET /metrics` endpoint (protected from public access via a simple token check, not exposed in the frontend). This gives a lightweight operational view without any paid tooling.

*Traces:* Distributed tracing is overkill for a single-service deployment. However, the session UUID serves as a trace ID — all log entries for a given request share the same session ID, allowing you to reconstruct the full pipeline execution for any given request from the logs.

**RAG-Specific Failure Modes**

Beyond standard API failures, RAG systems have specific failure modes to monitor:

*Retrieval failure — no relevant chunks found:* If all retrieved chunks have cosine similarity below 0.5, the retrieval has effectively failed. The model will have no grounded context and may hallucinate. Detection: log the similarity scores of retrieved chunks. Threshold alert: if max similarity < 0.5, the session should return a degraded response ("We couldn't find enough signal in this conversation to make personalized suggestions") rather than generating hallucinated output.

*JSON parsing failure from GPT-4o:* Despite JSON mode, the model occasionally returns malformed or schema-non-conformant JSON. The Go validation layer catches this and can retry the completion call with a slightly modified prompt (e.g., "Return ONLY a valid JSON object. Your last response was invalid."). Limit retries to 2.

*Budget compliance failure:* If GPT-4o returns gift suggestions outside the budget range (detected by the Go validation layer), those suggestions are filtered. If all suggestions are filtered (every suggestion violated the budget), the session falls back to retrying the completion with stronger budget emphasis in the prompt.

*Conversation too short / too uniform:* Some conversations (e.g., all messages are just emoji reactions) provide no meaningful signal. Detect this at the chunking stage (very few HasPreference: true chunks) and return a friendly message: "This conversation doesn't give us enough to work with — try uploading a longer or more varied conversation."

**Retry Logic for OpenAI API Calls**

OpenAI API calls can fail due to rate limits (HTTP 429), transient errors (HTTP 500), or timeouts. The Go HTTP client wrapper for OpenAI should implement:
- **Exponential backoff with jitter:** Wait 1s, then 2s, then 4s between retries, with ±500ms jitter to avoid thundering herd
- **Maximum 3 retries** before returning an error to the user
- **Retry only on transient errors** (429, 500, 502, 503, 504) — do not retry on 400 (bad request, not transient) or 401 (invalid API key, not retryable)

**Input Sanitization as a Security Layer**

Beyond the validation described in Module 3, the HTTP handler should:
- Limit request body size (Gin's default is 32MB — reduce this to 1MB for the upload endpoint)
- Reject requests with unexpected Content-Type headers
- Sanitize the conversation text to remove control characters that could affect parsing

**Zero-Cost Observability Tooling**

- **Render's built-in log viewer:** Render streams application logs in its dashboard. For a learning project, this is sufficient for debugging.
- **UptimeRobot (free tier):** Monitors the `/health` endpoint and sends email alerts on downtime.
- **OpenAI's usage dashboard:** Tracks API call volume and costs per day. The primary cost monitoring tool.
- **Sentry free tier:** Can capture Go errors and panics with stack traces. 5,000 errors/month free. Useful for catching unexpected failures in production.

---

### Design Decisions
- **Structured logging (JSON) from day one.** Ad-hoc `fmt.Println` debugging does not scale to production. Structured logs are filterable and searchable.
- **In-process metrics counters** rather than a metrics service. Sufficient for a free-tier deployment.
- **Similarity threshold check** as a retrieval quality gate. This prevents low-quality contexts from reaching GPT-4o and producing hallucinated responses.

---

### Alternative Approaches
- **OpenTelemetry:** The industry standard for distributed tracing and metrics. Integrates with many backends (Jaeger, Prometheus, Datadog). Excellent for production systems; overkill for this scale.
- **Prometheus + Grafana:** Full metrics stack. Beautiful dashboards. Requires running two additional services — not viable on free tier.

---

### Trade-offs / Consequences
- **In-process metrics are lost on restart.** A Render redeployment or spin-down/spin-up cycle resets all counters to zero. This is acceptable for a learning project — you are looking for trends within a session or deployment window, not long-term retention.

---

### How This Module Connects to the Overall System
Observability is a cross-cutting concern. The logging and metrics described here are woven into every pipeline stage. Without this layer, debugging production failures in the RAG pipeline is guesswork.

---

<a name="module-17"></a>
## Module 17 — Phased Build Plan

### Objective
Translate the full system design into a concrete, sequenced build plan organized into phases. Each phase ships a working, demonstrable increment of the system — not just a partial implementation that cannot be tested.

---

### Concepts to Learn
- Incremental delivery of a complex system
- Vertical slices vs. horizontal layers
- Walking skeleton architecture
- How to prioritize what to build first in a RAG system
- Technical debt decisions in a phased build

---

### Detailed Explanation

**Guiding Principle: Walking Skeleton First**

A "walking skeleton" is the thinnest possible implementation that exercises every layer of the system end-to-end. For GiftSense, this means Phase 1 should produce a system where: user uploads conversation → backend processes it → result is returned to frontend. The quality of the result in Phase 1 is deliberately minimal — the point is to have all the plumbing connected and working.

Each subsequent phase improves the quality of a specific layer without changing the end-to-end flow.

---

### Phase 1 — Walking Skeleton (MVP RAG Pipeline)

**Goal:** End-to-end system that accepts a conversation and returns gift suggestions, with Pinecone as the vector store, multi-window support via session UUIDs, and a mobile-responsive frontend.

**What is built:**
- Go backend with Gin: single `/api/analyze` endpoint
- Full config system: all parameters via env vars with defaults (`config.go`)
- Basic conversation parser (WhatsApp `.txt` format + plain text paste)
- File size enforcement at Gin middleware layer (2MB cap from `MAX_FILE_SIZE_BYTES`)
- Intelligent message sampling for large conversations (recency-biased, capped at `MAX_PROCESSED_MESSAGES`)
- Fixed-size chunking with configurable window/overlap via env vars
- OpenAI embedding calls using `EMBEDDING_MODEL` env var
- Pinecone adapter: upsert → query → delete namespace lifecycle
- Single retrieval query (multi-query deferred to Phase 3)
- GPT-4o completion call using `CHAT_MODEL` and `MAX_TOKENS` env vars
- Static shopping links (budget range embedded in hardcoded URL templates)
- React frontend: mobile-first, responsive, session UUID generated per window
- Upload component with mobile file picker + desktop drag-and-drop
- Budget tier card selector (responsive grid)
- Render deployment: both services deployed and functional

**What is deliberately skipped:**
- Anonymization (raw conversation text goes to OpenAI in Phase 1 — documented as a known limitation)
- Multi-query retrieval and metadata filtering
- Cold start handling
- Sophisticated error handling

**Learning outcomes from Phase 1:**
- Full RAG loop end-to-end, now with an external vector store
- Pinecone namespace lifecycle (create, populate, query, delete per session)
- Configuration-driven architecture: every parameter is an env var from day one
- Multi-window session isolation via client-generated UUID
- Mobile-first React layout with Tailwind breakpoints

**Estimated build time:** 3-4 weeks for a single developer (Pinecone integration and config system add ~1 week vs. in-memory version)

---

### Phase 2 — Privacy Layer (Anonymization)

**Goal:** Add the anonymization layer so that raw PII no longer reaches OpenAI.

**What is built:**
- Anonymizer module in Go: sender-field-seeded NER + regex-based proper noun detection
- Pseudonymization with stable token mapping
- Integration into the pipeline: anonymization happens between parsing and chunking
- Before/after logging (at DEBUG level only, never in production) to verify anonymization quality
- Frontend privacy notice (pre-upload banner and post-results reminder)
- Updated data handling disclosure in the UI

**What changes:**
- The Embedder now receives anonymized text only (verified by inspection of OpenAI request logs in development)
- The GPT-4o prompt now contains anonymized context only

**Learning outcomes from Phase 2:**
- Practical NER implementation in Go
- The architectural difference between policy-based and architecture-based privacy
- How anonymization affects the quality of retrieved context (some signal loss — observe and measure it)

**Estimated build time:** 1-2 weeks

---

### Phase 3 — Retrieval Quality (Multi-Query + Metadata)

**Goal:** Improve the quality of retrieved context through multi-query retrieval, metadata enrichment on chunks, and metadata-filtered retrieval.

**What is built:**
- Semantic metadata extraction during chunking (heuristic keyword-based Topics, EmotionalMarkers, HasPreference, HasWish flags)
- Multi-query retrieval (4 targeted queries instead of 1)
- Concurrent embedding of retrieval queries (using Go goroutines)
- Metadata filtering in the vector store's Query method
- Lightweight heuristic re-ranking of retrieved chunks
- Improved prompt: multi-query context assembled from diverse chunk types

**What changes:**
- Retrieval quality improves noticeably — personality insights become more specific and grounded
- Embedding stage becomes concurrent (latency improvement)

**Learning outcomes from Phase 3:**
- The impact of retrieval quality on generation quality
- Goroutine-based concurrency for I/O-bound API calls
- Metadata design in a vector store

**Estimated build time:** 1-2 weeks

---

### Phase 4 — Shopping Links & Budget Polish

**Goal:** Replace placeholder shopping links with real, budget-filtered, dynamically constructed links. Polish the budget enforcement pipeline end-to-end.

**What is built:**
- Full link generator module: Amazon India, Flipkart, Google Shopping URL construction
- Budget tier enum with min/max values and paise conversion for Amazon
- URL encoding for gift names (including non-ASCII handling)
- Backend validation of GPT-4o's budget compliance on suggestions
- Retry prompt if all suggestions fail budget validation
- Frontend gift card redesign: budget badge, three platform link buttons, disclaimer text

**What changes:**
- Gift suggestions become immediately actionable (real shopping links)
- Budget compliance is enforced at two levels: prompt and backend validation

**Learning outcomes from Phase 4:**
- URL construction as a programmatic pattern
- Defense-in-depth validation (LLM output + backend validation)
- Go's `net/url` package for URL encoding

**Estimated build time:** 1 week

---

### Phase 5 — Hardening & UX Polish

**Goal:** Production harden the system and polish the user experience for a real launch.

**What is built:**
- Cold start handling: frontend ping on page load
- User-facing cold start messaging ("server is warming up")
- Retry logic with exponential backoff in OpenAI client wrapper
- Similarity threshold quality gate in retrieval
- Graceful degraded response for low-quality conversations
- Input validation hardening (length caps, format detection, error messages)
- Structured logging (slog) throughout the pipeline
- In-process metrics endpoint
- Rate limiting middleware in Gin
- Full React UI polish: animations, loading states, error states, responsive design

**Learning outcomes from Phase 5:**
- Production hardening as a discipline distinct from feature development
- RAG-specific failure mode detection and graceful degradation
- Observability design for a constrained system

**Estimated build time:** 2 weeks

---

### Phase 6 (Optional Advanced) — Quality & Scale Enhancements

*These are not required for a functional, launchable product. They are presented as learning extensions:*

- **Screenshot OCR support:** Integration with a free-tier OCR API for image uploads
- **Hybrid retrieval (BM25 + dense):** Add sparse retrieval layer alongside embedding retrieval and merge rankings
- **HyDE retrieval:** Hypothetical document embedding for improved query-chunk alignment
- **GPT-4o-mini routing:** Complexity-based model selection for cost optimization
- **SSE progress streaming:** Real-time pipeline progress events streamed to the frontend
- **Two-call strategy:** Separate GPT-4o-mini call for personality insights, GPT-4o for gift suggestions

---

**Summary Roadmap**

| Phase | Focus | Duration | Key Learning |
|---|---|---|---|
| 1 | Walking Skeleton | 2-3 weeks | Full RAG loop, Go + OpenAI + Render |
| 2 | Privacy Layer | 1-2 weeks | Anonymization, PII handling, architecture-level privacy |
| 3 | Retrieval Quality | 1-2 weeks | Multi-query RAG, metadata, concurrency |
| 4 | Shopping Links | 1 week | URL construction, budget enforcement |
| 5 | Hardening | 2 weeks | Production resilience, observability |
| 6 | (Optional) Advanced | Ongoing | Hybrid retrieval, cost optimization |

**Total estimated build time to Phase 5 completion: 7-10 weeks for a single developer** building deliberately as a learning exercise (not sprinting).

---

---

<a name="module-18"></a>
## Module 18 — Application Configuration via Environment Variables

### Objective
Design a complete, type-safe, fail-fast configuration system that centralizes all runtime parameters — model names, token limits, file size caps, retrieval depth, vector store settings — into a single env-var-driven config layer. Understand why configurability is an architectural virtue, not just a convenience.

---

### Concepts to Learn
- The 12-Factor App methodology and its config principle
- Fail-fast startup configuration validation
- Type-safe config structs in Go
- The difference between secrets, operational parameters, and defaults
- Why configuration-driven design accelerates experimentation in RAG systems
- How env vars interact with Render's deployment model

---

### Detailed Explanation

**Why a Dedicated Config Layer Matters**

A RAG system has many tunable parameters: which model to use, how many tokens to allow, how many results to retrieve, what chunk size produces the best results. In an early-stage learning project, you will change these frequently as you experiment.

Without a dedicated config layer, you end up with hardcoded values scattered across dozens of files. Changing `TOP_K` from 3 to 5 means searching the codebase for magic numbers, changing them, rebuilding, and redeploying. Worse, you might change it in one place but miss it in another.

A dedicated config layer solves this: every tunable parameter is read from an environment variable once, at startup, into a typed `Config` struct. Every component that needs a value gets it from the config struct, not from `os.Getenv()`. The result is that changing any parameter requires exactly one action: update the env var in Render's dashboard and trigger a redeploy.

**The 12-Factor App Config Principle**

The 12-Factor App methodology (a widely adopted set of principles for building production software) states: "Store config in the environment." Specifically: anything that varies between deployments (development, staging, production) should be in environment variables, not in code.

For GiftSense, this includes:
- API keys (OpenAI, Pinecone) — vary between environments (dev uses a personal key, production uses a project key)
- Model names — vary between environments (dev uses `gpt-4o-mini` to save cost, production uses `gpt-4o`)
- File size limits — may vary between environments (dev is more permissive, production is stricter)
- Infrastructure endpoints — vary between environments (Pinecone region, index name)

**Config Categories and Their Handling**

| Category | Examples | Handling |
|---|---|---|
| **Secrets** | `OPENAI_API_KEY`, `PINECONE_API_KEY` | Required, no default, never logged, Render secret env var |
| **Model parameters** | `CHAT_MODEL`, `EMBEDDING_MODEL`, `MAX_TOKENS` | Required, sensible defaults, logged at startup (redacted for keys) |
| **Retrieval parameters** | `TOP_K`, `NUM_RETRIEVAL_QUERIES` | Optional, defaults provided, affect quality vs. cost |
| **Pipeline limits** | `MAX_FILE_SIZE_BYTES`, `MAX_PROCESSED_MESSAGES`, `CHUNK_WINDOW_SIZE` | Optional, defaults provided, affect performance and cost |
| **Infrastructure** | `PINECONE_INDEX_NAME`, `PINECONE_ENVIRONMENT`, `ALLOWED_ORIGINS`, `PORT` | Required in production, defaults for local dev |

**Fail-Fast Startup Validation**

The Go application's `config.Load()` function is called as the very first thing in `main.go`, before any adapters are constructed or the Gin router is set up. It reads all environment variables, validates them, and returns a populated `Config` struct — or returns an error that causes the application to exit immediately with a clear message.

Fail-fast startup is critical because the alternative — failing on the first API call at runtime — means a user triggers an error while actually using the application. With fail-fast validation, configuration errors are caught at deploy time, not at user-interaction time.

Validation rules:
- Secrets (`OPENAI_API_KEY`, `PINECONE_API_KEY`): must be non-empty strings. If missing, log a clear error message ("OPENAI_API_KEY environment variable is required") and exit with code 1.
- Numeric parameters (`MAX_TOKENS`, `TOP_K`, etc.): must parse as valid positive integers. If invalid, log the invalid value and the expected range, then exit.
- Model names: validated against a whitelist of known valid model strings. An unknown model name logs a warning but does not fail startup (future models may not be in the whitelist).
- `EMBEDDING_DIMENSIONS`: must match the dimensions of the chosen `EMBEDDING_MODEL`. A mismatch would cause silent failures in Pinecone (wrong-dimensional vectors). Log a warning if they don't correspond to known model/dimension pairs.

**Startup Config Log**

At startup, the application logs all configuration values at INFO level — except secrets, which are replaced with `[REDACTED]`. This log line is invaluable for debugging: when something behaves unexpectedly in production, the first thing to check is "what config was the application actually running with?" The startup log answers this immediately.

Example startup log output:
```
INFO  GiftSense starting
INFO  Config:
INFO    ChatModel=gpt-4o
INFO    EmbeddingModel=text-embedding-3-small
INFO    EmbeddingDimensions=1536
INFO    MaxTokens=1000
INFO    TopK=3
INFO    NumRetrievalQueries=4
INFO    MaxFileSizeBytes=2097152
INFO    MaxProcessedMessages=400
INFO    ChunkWindowSize=8
INFO    ChunkOverlapSize=3
INFO    PineconeIndexName=giftsense
INFO    PineconeEnvironment=us-east-1
INFO    AllowedOrigins=[https://giftsense.onrender.com]
INFO    OpenAIAPIKey=[REDACTED]
INFO    PineconeAPIKey=[REDACTED]
INFO  Server listening on :8080
```

**Local Development Configuration**

For local development, env vars are loaded from a `.env` file in the project root using a Go dotenv library (e.g., `github.com/joho/godotenv`). The `.env` file is listed in `.gitignore` and never committed. A `.env.example` file with placeholder values and comments is committed to the repository to document all required variables for new developers.

This pattern means:
- Local dev uses `.env` file (loaded by godotenv in dev mode)
- Production uses Render's environment variable dashboard
- The application code itself does not change between environments

**RAG-Specific Configuration Benefits**

The configurability directly benefits RAG experimentation:

*Changing `TOP_K`* affects how many chunks are retrieved per query. Lower K = faster, cheaper, less context. Higher K = more context, potentially better suggestions, more tokens. You can compare output quality between `TOP_K=3` and `TOP_K=5` without any code change.

*Changing `CHAT_MODEL`* between `gpt-4o` and `gpt-4o-mini` lets you benchmark output quality vs. cost on identical conversations. Set `gpt-4o-mini` in dev, `gpt-4o` in production.

*Changing `CHUNK_WINDOW_SIZE`* lets you experiment with chunking granularity. Smaller windows = more precise retrieval. Larger windows = more context per chunk. This is one of the highest-impact tuning parameters in RAG.

*Changing `MAX_TOKENS`* controls completion verbosity. Lower MAX_TOKENS = shorter, punchier insights and suggestions. Higher = more detailed output.

---

### Design Decisions
- **All configuration via env vars, zero hardcoded defaults in business logic.** Even `gpt-4o` is not hardcoded in the OpenAI adapter — it comes from the config struct.
- **`config.go` is the only file that calls `os.Getenv()`.** All other files receive values through the Config struct. This makes it trivial to trace where any value originates.
- **Fail-fast on missing secrets.** The application should never start in a broken state silently.
- **`.env` for local dev, Render dashboard for production.** No custom config file format, no YAML, no TOML — just environment variables as the 12-Factor App intends.

---

### Alternative Approaches
- **YAML or TOML config file:** More human-readable than env vars, supports nested structures. However, config files are usually committed to source control (a security risk for secrets), require file system access (unavailable on Render's ephemeral filesystem), and violate the 12-Factor App principle.
- **Hardcoded constants with build tags:** Use Go build tags to compile different constants per environment. Works but requires a separate build artifact per environment. Much more complex than env vars.
- **Viper (Go config library):** Supports env vars, config files, and remote config sources. More features than needed for GiftSense. The standard library (`os.Getenv` + `strconv`) plus a thin wrapper is sufficient.

---

### Trade-offs / Consequences
- **Env var names must be documented.** Without the `.env.example` file and the startup log, a new developer would not know which variables to set. This documentation burden is the main cost of the env-var approach.
- **Render redeploys on env var changes.** Changing `TOP_K` triggers a new deployment. This is a brief (~30 second) interruption. Acceptable for a learning/portfolio project; a production system might prefer a config service that allows hot reloads without restarts.
- **All configuration is flat (no nesting).** Environment variables cannot express hierarchical config (`openai.model.chat` is not a valid env var name). For GiftSense, the flat namespace is sufficient. For a larger system with dozens of parameters, a structured config format would be preferable.

---

### How This Module Connects to the Overall System
The `Config` struct is the first thing loaded in `main.go`. Every adapter (OpenAI embedder, OpenAI LLM client, Pinecone vector store), every use case (AnalyzeConversation), and the HTTP middleware (file size limit, CORS origins) receive their parameters from this struct. Module 18 is the wiring that makes every other module's design decisions tunable without touching code.

---

*End of GiftSense System Architecture & Learning Guide — v2.0 (Updated)*

---

## Appendix — Quick Reference

### Data Flow Summary
```
Upload (2MB max) → Session UUID (client) → Parse → Sample → Anonymize
→ Chunk (configurable window/overlap) → Embed (OpenAI, EMBEDDING_MODEL)
→ Pinecone Upsert (session namespace) → Multi-Query Retrieve (TOP_K per query)
→ Re-rank → Prompt Build → GPT-4o Complete (CHAT_MODEL, MAX_TOKENS)
→ Validate Budget → Link Generate → Pinecone Delete Namespace → Response
```

### Multi-Window Isolation
```
Browser Window 1 ─── session_id: "abc-123" ─── Pinecone namespace: "abc-123"
Browser Window 2 ─── session_id: "def-456" ─── Pinecone namespace: "def-456"
Browser Window N ─── session_id: "xyz-789" ─── Pinecone namespace: "xyz-789"
                         │
                         └── All namespaces deleted after each request completes
```

### What Each External Service Receives
| Service | What Is Sent | What Is NOT Sent |
|---|---|---|
| OpenAI Embeddings | Anonymized chunk text / query text | Real names, raw conversation |
| OpenAI Chat | Anonymized retrieved chunks + recipient metadata | Raw conversation, real names |
| Pinecone | Embedding vectors + boolean/string metadata | Any text content at all |

### Key Environment Variables
| Variable | Default | Effect |
|---|---|---|
| `CHAT_MODEL` | `gpt-4o` | Model for gift suggestions and insights |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | Model for chunk and query embedding |
| `MAX_TOKENS` | `1000` | Max output tokens per completion |
| `TOP_K` | `3` | Nearest-neighbor results per retrieval query |
| `MAX_FILE_SIZE_BYTES` | `2097152` (2MB) | Maximum uploaded conversation size |
| `CHUNK_WINDOW_SIZE` | `8` | Messages per chunk |
| `PINECONE_INDEX_NAME` | `giftsense` | Pinecone index to use |

### Per-Session Cost Estimate (Updated)
| Component | Approximate Cost |
|---|---|
| Chunk embeddings (~40 chunks) | $0.0003 |
| Query embeddings (4 queries) | < $0.0001 |
| GPT-4o completion | ~$0.036 |
| Pinecone API calls | Free tier |
| **Total** | **~$0.037** |

### Key Go Interfaces
- `Embedder`: `Embed(texts []string) ([][]float32, error)`
- `LLMClient`: `Complete(prompt string, opts Options) (string, error)`
- `VectorStore`: `Upsert(sessionID string, chunks []Chunk, vectors [][]float32) error` / `Query(sessionID string, vector []float32, topK int, filter Filter) ([]Chunk, error)` / `DeleteSession(sessionID string) error`

---
*Document version 2.0 — GiftSense Architecture Guide*
*(Changes from v1: Pinecone free tier, multi-window session support, 2MB file limit, full env-var configurability, mobile-first frontend, Module 18 added)*
