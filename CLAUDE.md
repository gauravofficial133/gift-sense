# GiftSense (upahaar.ai)

AI gifting platform: WhatsApp/audio/Spotify → personality insights + gift suggestions + greeting cards.

## Tech Stack

**Backend:** Go 1.22, Gin, OpenAI SDK (LLM + DALL-E), Anthropic SDK, Pinecone, rod (headless Chrome for card rendering)
**Frontend:** React 19, Vite, Tailwind CSS, Lucide React
**Deployment:** Docker (backend needs Chromium), Vercel (frontend)

## Project Structure

```
giftsense-backend/
  cmd/server/main.go              ← Entry point
  config/config.go                ← All env vars, fail-fast validation
  internal/
    domain/                       ← Types, interfaces, errors (no external imports)
    port/                         ← Interface definitions (Embedder, LLMClient, VectorStore, etc.)
    usecase/                      ← Business logic (analyze, parse, chunk, card_generator, etc.)
    adapter/
      openai/                     ← Embedder + LLM adapters
      imagegen/                   ← DALL-E 3 illustration generator
      vectorstore/                ← Pinecone + in-memory store
      cardrender/                 ← Chrome rendering: chrome.go, renderer.go, template.go
      linkgen/                    ← Shopping link generator
      sarvam/                     ← Audio transcription
      spotify/                    ← Spotify integration
    delivery/
      http/                       ← Gin handlers, middleware, CORS
      dto/                        ← Request/response types
  assets/
    recipes/                      ← HTML/CSS card templates (8 premium recipes)
    webfonts/                     ← Embedded .woff2 Google Fonts (8 fonts)
    generated/                    ← AI-generated illustration cache (index.json + .b64 files)
    fonts/                        ← TTF fonts (legacy, used by old PDF renderer)

giftsense-frontend/
  src/
    components/                   ← React components (upload, form, results, shared)
    hooks/                        ← useSession, useAnalyze
    screens/                      ← InputScreen, LoadingScreen, ResultsScreen
    lib/                          ← Utilities (cardDownload.js)
    api/                          ← API client
```

## Code Rules

- Clean Architecture: domain/ and usecase/ must NOT import external packages
- TDD: test file first, then implementation
- Test naming: `Test[Function]_Should[Behavior]_When[Condition]`
- No comments in code unless explaining a complex algorithm
- Functions max 30 lines, single responsibility
- Errors wrapped with context: `fmt.Errorf("context: %w", err)`
- All env vars via config.Config — never `os.Getenv()` outside config/
- No localStorage/sessionStorage in frontend
- Mobile-first Tailwind: base → sm: → md: → lg:

## Card Rendering Pipeline (Phase 3 Complete)

```
Analyze → Detect Occasion + Emotions
  → 4 parallel LLM art-direction calls (2 OpenAI + 2 Claude, different variation styles)
  → Each LLM picks recipe + palette + fonts + writes card copy (structured JSON)
  → Optionally requests AI illustration (DALL-E 3) — cached via AssetLibrary
  → Template Engine populates HTML with palette CSS vars + fonts + content + illustration
  → Chrome renders HTML → PNG (preview at 2x, PDF at 4x / 300+ DPI)
  → Print PDF includes 3mm bleed area + trim marks
  → Multi-page PDF: front (visual) + inside (message) stitched into 2-page PDF
  → Returns Cards []*CardRender with model label
  → Frontend shows 2x2 grid, user taps to preview, edit (palette/text), download PDF
```

8 premium templates: gilded-bloom, confetti-cinema, sacred-geometry, clean-minimal, clean-split, clean-framed, layered-floral, layered-festive.
8 named palettes. 8 embedded Google Fonts. DALL-E 3 illustration generation.

## Implementation Plan

Full phased plan at: `.claude/plans/premium-card-quality-plan.md`

- **Phase 1 (DONE):** Chrome rendering pipeline, 3 premium templates, font bundling, palette expansion
- **Phase 2 (DONE):** Anthropic SDK, LLM art-direction, 4 cards/request, 5 more templates, frontend 2x2 grid, on-demand PDF endpoint
- **Phase 3 (DONE):** AI illustration generation (DALL-E 3), asset library, user editing (palette/text), print readiness (300 DPI, bleed, trim marks), multi-page cards (front + inside)

## Reference Documents

- `docs/card-design-plan.md` — Full card engine architecture vision
- `docs/premium-card-templates.md` — 3 template specs with layer-by-layer detail + 28 SVG asset list
- `market-analysis.md` — Market strategy, TAM/SAM/SOM, competitive landscape, revenue model
