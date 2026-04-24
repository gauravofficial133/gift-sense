# Greeting Card Engine — Refined Design Plan

*Created: 2026-04-22 | Last Updated: 2026-04-22*

---

## Vision

Generate Figma/Canva-level personalized greeting cards from real conversations. Four unique card previews per request — two from OpenAI, two from Claude — reflecting the emotional tone of the user's relationship. The moat is personalized content; the card engine is the packaging that makes it worth sharing.

---

## Core Principle

**Don't ask the LLM to paint. Ask it to art-direct.**

The LLM detects emotion, picks a composition recipe, selects a palette, pairs fonts, chooses illustrations, and writes the copy. Human-designed templates and professionally sourced assets handle the visual execution.

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Card tiers | Clean (simple) + Canva-level (layered) | Two quality levels in one system |
| Cards per request | 4 (2 OpenAI + 2 Claude) | Dual-model variety, user picks favorite |
| User flow | All-in-one | Single upload → insights + gifts + Spotify + 4 cards |
| Composition | Preset recipes with variation params | LLMs are spatially blind; presets guarantee quality floor |
| Color system | LLM picks from curated palettes based on detected emotion | Emotion-driven, guaranteed color harmony |
| Fonts | LLM picks from curated Google Fonts list | Creative freedom within vetted options |
| Illustrations | Hybrid — curated SVGs + AI-generated PNGs | Growing asset library over time |
| Rendering | Headless Chrome via `rod` (Go) | HTML/CSS → screenshot (preview) + PDF (export), exact match |
| Preview layout | 2x2 grid on mobile and desktop | All 4 visible, click to enlarge |
| PDF generation | On-demand for selected card only | No wasted rendering |
| Template design | Reference-driven from Canva/Pinterest examples | Professional aesthetic, replicated in HTML/CSS |
| MVP recipe count | 8-10 (3-4 clean + 5-6 layered) | Enough variety, fast to ship |
| Deployment | Local-first for MVP | Headless Chrome can't run on Vercel serverless |

---

## Architecture

### Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        CALL 1 — UNDERSTAND                      │
│                                                                 │
│  Conversation Upload                                            │
│       ↓                                                         │
│  Single LLM Analysis Call                                       │
│       ↓                                                         │
│  Extracts ALL of:                                               │
│    • Personality insights                                       │
│    • Gift suggestions                                           │
│    • Emotional tone (primary + secondary emotions)              │
│    • Spotify song recommendation                                │
│    • Card copy elements (themes, key phrases, relationship)     │
└──────────────────────────────┬──────────────────────────────────┘
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│                    CALL 2 — DESIGN (4 in parallel)              │
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ OpenAI 1 │ │ OpenAI 2 │ │ Claude 1 │ │ Claude 2 │          │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘          │
│       ↓             ↓             ↓             ↓               │
│  Each outputs:                                                  │
│    • composition_recipe (from preset list)                      │
│    • palette (from curated list, emotion-driven)                │
│    • font pairing (headline + body from curated list)           │
│    • illustration selections (from library OR generate new)     │
│    • illustration_prompt (if generating new)                    │
│    • headline, body, closing, signature text                    │
└──────────────────────────────┬──────────────────────────────────┘
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│                     RENDER (4 in parallel)                       │
│                                                                 │
│  For each card:                                                 │
│    1. If generate_illustration: true → call image model         │
│    2. Tag + save new illustration to asset library              │
│    3. Fill HTML/CSS template with recipe + palette + fonts      │
│    4. Inject illustrations (SVG or PNG)                         │
│    5. Headless Chrome → screenshot (preview PNG)                │
│                                                                 │
│  Display: 2x2 grid of 4 preview screenshots                    │
│  User picks one → Chrome renders PDF for that card only         │
└─────────────────────────────────────────────────────────────────┘
```

### LLM Art-Direction Output Schema

```json
{
  "composition_recipe": "layered_confetti_burst",
  "palette": "sunset_warmth",
  "headline_font": "Pacifico",
  "body_font": "Lato",
  "illustrations": ["balloon_cluster", "confetti_scatter"],
  "generate_illustration": true,
  "illustration_prompt": "watercolor birthday balloons floating upward, transparent background, soft edges",
  "headline": "You Make Every Year Better",
  "body": "From late-night conversations to spontaneous road trips — every moment with you is a gift I never knew I needed.",
  "closing": "With all my love",
  "signature": "Priya"
}
```

---

## Two Card Tiers

### Tier 1 — Clean & Modern

- Simple layout structures with clear visual hierarchy
- Curated SVG illustrations from open libraries
- Generous whitespace, strong typography
- Think: premium Etsy greeting card

**3-4 Clean Recipes for MVP:**

| Recipe | Structure | Best For |
|--------|-----------|----------|
| `clean_hero_top` | Illustration top (40%) + text bottom (60%) | Birthday, celebration |
| `clean_minimal` | Small accent illustration + large typography | Modern, any occasion |
| `clean_split` | Illustration left, text right | Thank you, friendship |
| `clean_framed` | Decorative border with centered content | Wedding, anniversary |

### Tier 2 — Canva-Level Layered

- Multiple overlapping elements with z-ordering
- Absolute-positioned decorative layers
- Blend modes, subtle shadows, gradient backgrounds
- AI-generated illustrations for unique visuals
- Think: Canva Pro template

**5-6 Layered Recipes for MVP:**

| Recipe | Structure | Best For |
|--------|-----------|----------|
| `layered_confetti_burst` | Scattered confetti top + hero text center + accent bottom | Birthday, celebration |
| `layered_floral_wrap` | Floral elements wrapping text from corners | Anniversary, Mother's Day |
| `layered_gradient_dream` | Gradient background + floating illustration elements + overlay text | Romantic, Valentine's |
| `layered_collage` | Multiple illustration panels with text woven between | Friendship, farewell |
| `layered_spotlight` | Dark/rich background + illuminated center text + decorative edges | Elegant, formal |
| `layered_festive` | Pattern background + decorative border + layered ornaments | Diwali, Raksha Bandhan, festivals |

### Recipe Anatomy (HTML/CSS Template)

Each recipe is a standalone HTML file with:

```
recipe-name/
├── template.html       ← HTML structure with {{slot}} placeholders
├── styles.css          ← CSS with custom properties for palette injection
├── preview.png         ← Reference screenshot (from Canva/Pinterest)
└── metadata.json       ← Occasions, moods, tier, slot definitions
```

**CSS Custom Properties (injected per card):**

```css
:root {
  --bg-primary: {{palette.background}};
  --bg-secondary: {{palette.background_secondary}};
  --text-headline: {{palette.headline}};
  --text-body: {{palette.body}};
  --accent: {{palette.accent}};
  --overlay: {{palette.overlay}};
  --font-headline: '{{headline_font}}', cursive;
  --font-body: '{{body_font}}', sans-serif;
}
```

---

## Color Palette System

### Emotion → Palette Mapping

The LLM detects emotional tone from the conversation, then picks a palette that expresses that emotion. Multiple palettes per emotion for variety.

**Curated Palette Examples:**

| Palette Name | Emotion Tags | Colors (bg, headline, body, accent, overlay) |
|-------------|-------------|----------------------------------------------|
| `sunrise_warmth` | nostalgic, warm, grateful | #FFF5E6, #D4451A, #5C3D2E, #FFB347, rgba(255,180,71,0.1) |
| `soft_rose_gold` | tender, romantic, bittersweet | #FFF0F0, #8B3A62, #5C4A4A, #E8A0BF, rgba(232,160,191,0.1) |
| `ocean_calm` | peaceful, reflective, serene | #EBF5FB, #1A5276, #2C3E50, #5DADE2, rgba(93,173,226,0.1) |
| `electric_joy` | playful, energetic, excited | #FFFDE7, #E65100, #3E2723, #FFD600, rgba(255,214,0,0.15) |
| `midnight_elegant` | sophisticated, deep, formal | #1A1A2E, #E8D5B7, #C4B59D, #B8860B, rgba(184,134,11,0.1) |
| `forest_peace` | grounded, sincere, calm | #F1F8E9, #2E7D32, #3E4A3E, #81C784, rgba(129,199,132,0.1) |
| `lavender_dream` | gentle, whimsical, caring | #F3E5F5, #6A1B9A, #4A3A5C, #CE93D8, rgba(206,147,216,0.1) |
| `golden_celebration` | joyful, proud, festive | #FFFAF0, #BF360C, #4E342E, #FFB300, rgba(255,179,0,0.15) |

**Target: 15-20 palettes covering the full emotional spectrum.**

The LLM prompt includes the full palette list with emotion tags. It picks based on the detected emotional tone — not the occasion. A funny birthday gets `electric_joy`, a sentimental birthday gets `sunrise_warmth`.

---

## Font System

### Curated Google Fonts List

The LLM picks one headline font + one body font per card.

**Headline Fonts (expressive):**

| Font | Vibe | Best For |
|------|------|----------|
| Pacifico | Playful, handwritten | Birthday, fun occasions |
| Playfair Display | Elegant serif | Wedding, anniversary |
| Dancing Script | Flowing cursive | Romantic, Valentine's |
| Lobster | Bold, friendly | Celebration, thank you |
| Cormorant Garamond | Classic, refined | Formal, elegant |
| Caveat | Casual handwritten | Friendship, farewell |
| Abril Fatface | Dramatic display | Statement cards |
| Great Vibes | Formal script | Wedding, premium |

**Body Fonts (readable):**

| Font | Vibe |
|------|------|
| Lato | Clean, neutral |
| Nunito | Soft, rounded |
| Source Serif Pro | Traditional readability |
| Quicksand | Modern, geometric |
| Merriweather | Warm serif |
| Inter | Swiss precision |

**Pairing Rules (included in LLM prompt):**
- Script/display headline + sans-serif body (most common)
- Serif headline + sans-serif body (formal)
- Never two scripts together
- Never two serifs together

---

## Illustration Asset Library

### Hybrid Approach

```
assets/
├── curated/                    ← Pre-sourced, manually vetted
│   ├── birthday/
│   │   ├── balloons.svg
│   │   ├── cake.svg
│   │   ├── confetti.svg
│   │   └── ...
│   ├── anniversary/
│   ├── valentines/
│   ├── thankyou/
│   ├── farewell/
│   ├── diwali/
│   ├── raksha-bandhan/
│   └── general/               ← Hearts, stars, flowers, borders
│       ├── decorative/
│       └── patterns/
│
├── generated/                  ← AI-generated, auto-tagged
│   ├── index.json             ← Tags, occasion, emotion, file path
│   └── images/
│       ├── gen_001_watercolor_balloons.png
│       ├── gen_002_floral_border.png
│       └── ...
│
└── fonts/                      ← Google Fonts .woff2 files (bundled for offline)
```

### Asset Library Growth Loop

1. Card generation requests an illustration
2. LLM checks `generated/index.json` for matching tags
3. Match found → reuse existing asset
4. No match → generate new via image model
5. New asset saved to `generated/images/` with tags in `index.json`
6. Library grows richer with every card generated

### Index Entry Format

```json
{
  "id": "gen_001",
  "file": "images/gen_001_watercolor_balloons.png",
  "tags": ["birthday", "balloons", "watercolor", "playful", "warm"],
  "occasion": "birthday",
  "emotions": ["joyful", "playful"],
  "style": "watercolor",
  "dimensions": "512x512",
  "created": "2026-04-22",
  "usage_count": 3
}
```

---

## Rendering Engine

### Tech Stack

- **Go** with `rod` library (headless Chrome control)
- HTML/CSS templates with slot injection
- Google Fonts loaded via `@font-face` from bundled .woff2 files
- PNG illustrations embedded as `<img>` or inline base64
- SVG illustrations embedded inline

### Render Pipeline (per card)

```
1. Load recipe HTML template
2. Inject CSS custom properties (palette colors)
3. Inject font-face declarations
4. Inject illustration assets (SVG inline or PNG as <img>)
5. Inject text content (headline, body, closing, signature)
6. Open in headless Chrome
7. Screenshot → preview.png (displayed in 2x2 grid)
8. [On user download] Print to PDF → card.pdf
```

### Card Dimensions

- **Preview:** 800x1120px (5:7 ratio, standard greeting card)
- **PDF:** A5 (148mm x 210mm) at 300 DPI
- Both rendered from the same HTML — Chrome handles the scaling

### Local-First MVP

- `rod` auto-downloads Chromium on first run
- No cloud dependency for rendering
- `go run cmd/server/main.go` starts everything
- Chrome process managed by `rod` — launches, renders, closes per batch

---

## Frontend — Results Screen Update

### 2x2 Card Grid Component

```
┌─────────────────────────────────────────┐
│          Results Screen                  │
│                                         │
│  [Personality Insights]                 │
│  [Gift Suggestions + Shopping Links]    │
│  [Spotify Recommendation]              │
│                                         │
│  ┌─── Greeting Cards ────────────────┐  │
│  │                                   │  │
│  │  ┌──────────┐  ┌──────────┐      │  │
│  │  │  Card 1  │  │  Card 2  │      │  │
│  │  │ (OpenAI) │  │ (OpenAI) │      │  │
│  │  │          │  │          │      │  │
│  │  │ [Download]│  │ [Download]│     │  │
│  │  └──────────┘  └──────────┘      │  │
│  │                                   │  │
│  │  ┌──────────┐  ┌──────────┐      │  │
│  │  │  Card 3  │  │  Card 4  │      │  │
│  │  │ (Claude) │  │ (Claude) │      │  │
│  │  │          │  │          │      │  │
│  │  │ [Download]│  │ [Download]│     │  │
│  │  └──────────┘  └──────────┘      │  │
│  │                                   │  │
│  └───────────────────────────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

### Interactions

- **Click card** → fullscreen modal with larger preview
- **Download button** → triggers PDF render on backend → downloads file
- Cards display which model generated them (subtle label)
- Loading state: skeleton cards with shimmer animation while rendering

---

## Backend — New Endpoints

### Card Generation (part of existing analyze flow)

The `/api/v1/analyze` response expands to include card previews:

```json
{
  "data": {
    "personality_insights": [...],
    "gift_suggestions": [...],
    "spotify": {...},
    "emotion": {
      "primary": "nostalgic",
      "secondary": "grateful",
      "intensity": "deep"
    },
    "cards": [
      {
        "id": "card_uuid_1",
        "preview_url": "/api/v1/cards/card_uuid_1/preview",
        "model": "openai",
        "recipe": "clean_hero_top",
        "palette": "sunrise_warmth"
      },
      {
        "id": "card_uuid_2",
        "preview_url": "/api/v1/cards/card_uuid_2/preview",
        "model": "openai",
        "recipe": "layered_confetti_burst",
        "palette": "sunrise_warmth"
      },
      {
        "id": "card_uuid_3",
        "preview_url": "/api/v1/cards/card_uuid_3/preview",
        "model": "claude",
        "recipe": "clean_minimal",
        "palette": "sunrise_warmth"
      },
      {
        "id": "card_uuid_4",
        "preview_url": "/api/v1/cards/card_uuid_4/preview",
        "model": "claude",
        "recipe": "layered_floral_wrap",
        "palette": "sunrise_warmth"
      }
    ]
  }
}
```

### New Endpoints

```
GET  /api/v1/cards/:card_id/preview    → Returns preview PNG
GET  /api/v1/cards/:card_id/download   → Triggers PDF render, returns PDF file
```

### Temporary Storage

Card HTML and previews stored in a temporary directory, cleaned up after 30 minutes or on session deletion. No persistent card storage in MVP.

---

## Implementation Phases

### Phase 1 — Foundation (Week 1)

- [ ] Set up `rod` dependency in Go backend
- [ ] Create recipe template structure (`recipes/` directory)
- [ ] Build HTML/CSS renderer that takes recipe + data → rendered HTML
- [ ] Chrome screenshot pipeline (HTML → PNG preview)
- [ ] Chrome PDF pipeline (HTML → PDF download)
- [ ] Bundle Google Fonts (.woff2 files)

### Phase 2 — Recipes & Assets (Week 1-2)

- [ ] Build 3-4 clean recipes from reference screenshots
- [ ] Build 5-6 layered recipes from reference screenshots
- [ ] Curate initial SVG asset library (5-10 per occasion category)
- [ ] Define 15-20 color palettes with emotion tags
- [ ] Test all recipes with sample data in Chrome

### Phase 3 — LLM Integration (Week 2)

- [ ] Expand analysis prompt to extract emotional tone
- [ ] Build card art-direction prompt (for both OpenAI and Claude)
- [ ] Implement 4-card parallel generation (2 OpenAI + 2 Claude)
- [ ] Parse LLM art-direction JSON output
- [ ] Wire LLM output → template renderer

### Phase 4 — Illustration Generation (Week 2-3)

- [ ] Integrate image generation API (DALL-E for OpenAI cards, Claude for Claude cards)
- [ ] Build asset library index (`generated/index.json`)
- [ ] Implement asset lookup before generation (reuse existing)
- [ ] Auto-tag and save new illustrations
- [ ] Transparent PNG handling in templates

### Phase 5 — Frontend Integration (Week 3)

- [ ] Build 2x2 card grid component
- [ ] Card preview loading with skeleton animation
- [ ] Click-to-enlarge modal
- [ ] Download button → PDF endpoint
- [ ] Model label on each card
- [ ] Mobile responsive layout for 2x2 grid

### Phase 6 — Polish & Testing (Week 3-4)

- [ ] Test all 10 recipes across all occasions
- [ ] Test emotional tone detection accuracy
- [ ] Test font rendering across recipes
- [ ] Test PDF output quality (print-ready)
- [ ] Performance optimization (parallel rendering)
- [ ] Edge cases: very long text, very short text, special characters

---

## TODO — What's Needed From You

### Before Implementation Can Start

- [ ] **Collect 8-10 reference card designs** — Browse Canva, Pinterest, Dribbble, or real greeting cards. Save screenshots of designs you love. You need:
  - 3-4 clean/minimal card designs
  - 5-6 layered/rich card designs
  - Mix of occasions (birthday, anniversary, festival, etc.)
  - Save as images in a `docs/card-references/` folder with descriptive names (e.g., `clean_minimal_birthday.png`, `layered_floral_anniversary.png`)

- [ ] **Curate initial SVG assets** (2-3 hours, one-time) — Source from free libraries:
  - [unDraw](https://undraw.co) — MIT licensed flat illustrations
  - [SVGRepo](https://svgrepo.com) — 500k+ open-licensed SVGs
  - [Humaaans](https://humaaans.com) — character illustrations
  - [Open Peeps](https://openpeeps.com) — hand-drawn characters
  - Download 5-10 SVGs per occasion category
  - Organize in `assets/curated/{occasion}/` folders

- [ ] **Provide Anthropic API key** — Needed for Claude-side card generation (2 of 4 cards use Claude)

- [ ] **Install Chrome/Chromium locally** — `rod` auto-downloads but having it pre-installed avoids first-run delays

### During Implementation

- [ ] **Review rendered recipes** — After each recipe is built from your reference, review the output and provide feedback on:
  - Does it match the reference design's feel?
  - Are the layer positions right?
  - Does the typography feel balanced?

- [ ] **Test with real conversations** — Run the pipeline with actual chat exports to verify:
  - Emotion detection accuracy
  - Copy quality (headline, body text)
  - Palette appropriateness for the detected emotion

### Ongoing (Post-MVP)

- [ ] **Expand recipe library** — Add more reference designs as needed
- [ ] **Curate more SVG assets** — Especially for Indian festivals and culturally specific occasions
- [ ] **Review AI-generated illustrations** — Periodically check `generated/` folder quality, remove bad ones
- [ ] **Collect user feedback** — Which of the 4 cards users pick most → informs future recipe priorities

---

## Success Criteria

A card is "worth sharing" when:

1. A stranger shown the card says "that looks professional" without knowing it's AI-generated
2. The recipient recognizes something personal that a generic Hallmark card wouldn't have
3. The visual quality is comparable to a mid-tier Canva Pro template
4. Preview matches PDF export exactly (same rendering engine guarantees this)
5. The emotional tone of the card matches the conversation's mood
6. At least 1 of the 4 generated cards is one the user would actually send

---

## Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| Image generation latency adds 10-15s per card | Slow UX | Generate illustrations in parallel; reuse from growing library |
| LLM picks clashing recipe + palette combinations | Ugly cards | Metadata constraints — each recipe lists compatible palettes |
| Chrome rendering is memory-heavy (4 concurrent) | Crashes on low-RAM machines | Render sequentially on <8GB RAM, parallel on ≥8GB |
| AI-generated PNGs have inconsistent style | Cards look disjointed | Include style keywords in every prompt ("flat illustration, clean edges, transparent background") |
| Font loading fails in headless Chrome | Broken typography | Bundle .woff2 locally, no CDN dependency |
| Dual-model (OpenAI + Claude) adds API complexity | More failure modes | Independent error handling per model; if one fails, return 2 cards instead of 4 |

---

## Future Enhancements (Post-MVP)

- **User editing:** Let users tweak text, swap illustrations, change palette after generation
- **Template marketplace:** Users submit and share recipe designs
- **Animation:** Animated card previews (CSS animations in the HTML template)
- **Multi-page cards:** Front + inside left + inside right + back
- **Envelope design:** Matching envelope with recipient name
- **Print integration:** Direct integration with printing services
- **Cloud rendering:** Move Chrome rendering to a dedicated service for Vercel deployment

---

*The moat is personalized content from real conversations. The card engine makes it beautiful. Both models competing for the best design makes it remarkable.*
