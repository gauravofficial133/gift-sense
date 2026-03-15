# GiftSense — Deployment Guide (Vercel)

> **Platform:** Vercel Hobby (free tier)
> **Vector DB:** Pinecone (serverless, existing)
> **LLM / Embeddings:** OpenAI API

---

## Architecture overview

```
GitHub repo (monorepo)
├── giftsense-backend/   → Vercel project A  (Go serverless function)
└── giftsense-frontend/  → Vercel project B  (Vite static site, CDN)
```

Vercel does not run persistent HTTP servers. The backend is adapted with a
single `api/index.go` entry point that wraps the existing Gin engine via
`ServeHTTP`. All business logic, adapters, and tests are unchanged.

---

## Codebase changes made for Vercel

| File | Change | Reason |
|------|--------|--------|
| `giftsense-backend/api/index.go` | **New** | Vercel serverless entry point — wraps Gin via `ServeHTTP`, wired with `sync.Once` so the engine is built once per cold start |
| `giftsense-backend/vercel.json` | **New** | Rewrite all inbound paths to `api/index`; sets `maxDuration: 60` |
| `giftsense-frontend/vercel.json` | **New** | SPA rewrite — all paths serve `index.html` |
| `cmd/server/main.go` | **Unchanged** | Still used for `go run ./cmd/server` local development |

**Nothing else changed.** Domain, ports, usecases, adapters, tests, config — all
identical. The Vercel entry point is additive only.

### Why wrap Gin instead of replacing it?

Vercel's Go runtime accepts any `func(http.ResponseWriter, *http.Request)`.
`gin.Engine` implements `http.Handler`, so `ginRouter.ServeHTTP(w, r)` routes
the request through the full Gin stack — middleware (CORS, size limiter,
recovery), path matching, and handlers — exactly as in local development.
No routing logic is duplicated.

### Trade-offs

| Decision | Alternative considered | Why this choice |
|----------|----------------------|-----------------|
| Single `api/index.go` catches all routes | One file per endpoint | Avoids duplicating CORS + size-limiter middleware setup; keeps Gin as the single source of routing truth |
| `sync.Once` for engine init | Init on every request | Reusing the Gin engine across warm requests avoids re-creating clients on every call; negligible overhead vs. actual OpenAI/Pinecone latency |
| `maxDuration: 60` | 300 (Fluid compute) | 60 s is safe for the Hobby plan regardless of Fluid compute status; the analyze pipeline completes in 10–30 s in practice |
| `cmd/server/main.go` kept intact | Remove it | Local dev (`go run`) and CI still work without needing Vercel CLI |

---

## Prerequisites

| Requirement | Detail |
|-------------|--------|
| Vercel account | Free Hobby plan — [vercel.com](https://vercel.com) |
| GitHub repo | Push this monorepo to a **personal** GitHub account (Hobby plan does not support org repos) |
| OpenAI API key | Access to `gpt-4o` and `text-embedding-3-small` |
| Pinecone account | Free Serverless index already created (see below) |

### Pinecone index (if not already created)

1. Log in to [app.pinecone.io](https://app.pinecone.io)
2. **Create Index** with these exact settings:

   | Setting | Value |
   |---------|-------|
   | Name | `giftsense` |
   | Dimensions | `1536` |
   | Metric | `Cosine` |
   | Type | `Serverless` |
   | Cloud / Region | `AWS us-east-1` |

3. Copy the **API key** and note the **environment** (`us-east-1`)

---

## Step-by-step deployment

### Step 1 — Push to GitHub

```bash
# from the repo root
git push origin main
```

Both `giftsense-backend/api/index.go` and the `vercel.json` files must be
committed and on the branch you will deploy from.

---

### Step 2 — Create the backend Vercel project

1. Go to [vercel.com/new](https://vercel.com/new) → **Add New Project**
2. Import the `gift-sense` GitHub repository
3. In **Configure Project**:
   - **Framework Preset:** Other
   - **Root Directory:** `giftsense-backend` ← **critical**
   - Build & Output settings: leave blank (Vercel detects Go automatically)
4. Click **Deploy** — it will fail on the first deploy because env vars are not set yet. That is expected.

---

### Step 3 — Set backend environment variables

In **Vercel Dashboard → giftsense-backend → Settings → Environment Variables**,
add the following for the **Production** environment:

| Variable | Value | Required |
|----------|-------|----------|
| `OPENAI_API_KEY` | `sk-...` | ✅ |
| `PINECONE_API_KEY` | `pcsk-...` | ✅ |
| `PINECONE_ENVIRONMENT` | `us-east-1` | ✅ |
| `ALLOWED_ORIGINS` | _(leave blank for now — fill in after Step 5)_ | ✅ |
| `PINECONE_INDEX_NAME` | `giftsense` | optional (default applied) |
| `CHAT_MODEL` | `gpt-4o` | optional |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | optional |
| `EMBEDDING_DIMENSIONS` | `1536` | optional |
| `MAX_TOKENS` | `1000` | optional |
| `TOP_K` | `3` | optional |
| `NUM_RETRIEVAL_QUERIES` | `4` | optional |
| `MAX_FILE_SIZE_BYTES` | `2097152` | optional |
| `MAX_PROCESSED_MESSAGES` | `400` | optional |
| `CHUNK_WINDOW_SIZE` | `8` | optional |
| `CHUNK_OVERLAP_SIZE` | `3` | optional |

After saving, trigger a **Redeploy** from the Deployments tab.

> **Note:** Environment variable changes only take effect on the next deployment.
> Vercel does not hot-reload running functions.

---

### Step 4 — Verify the backend

Once deployed, your backend URL will be `https://giftsense-backend-<hash>.vercel.app`
(visible in the Vercel dashboard).

```bash
# Health check
curl https://giftsense-backend-<hash>.vercel.app/health
# Expected: {"status":"ok"}
```

If you get a 500, check **Vercel Dashboard → giftsense-backend → Functions → Logs**
for the startup error (almost always a missing env var).

---

### Step 5 — Create the frontend Vercel project

1. Go to [vercel.com/new](https://vercel.com/new) → **Add New Project**
2. Import the same `gift-sense` repository
3. In **Configure Project**:
   - **Framework Preset:** Vite _(auto-detected)_
   - **Root Directory:** `giftsense-frontend` ← **critical**
   - Build Command: `vite build` _(auto-detected)_
   - Output Directory: `dist` _(auto-detected)_
4. Before deploying, set this environment variable:

   | Variable | Value |
   |----------|-------|
   | `VITE_API_URL` | `https://giftsense-backend-<hash>.vercel.app` |

5. Click **Deploy**

> `VITE_API_URL` is a **build-time** variable — Vite embeds it into the JS bundle.
> Changing it always requires a full rebuild and redeploy of the frontend.

---

### Step 6 — Wire CORS

Now that you have both URLs, go back to the **backend** project:

**Settings → Environment Variables → `ALLOWED_ORIGINS`**

Set the value to your frontend URL (no trailing slash):

```
https://giftsense-frontend-<hash>.vercel.app
```

Trigger a **Redeploy** of the backend.

Verify CORS is working:

```bash
curl -H "Origin: https://giftsense-frontend-<hash>.vercel.app" \
     -I https://giftsense-backend-<hash>.vercel.app/health
# Must include: Access-Control-Allow-Origin: https://giftsense-frontend-<hash>.vercel.app
```

---

### Step 7 — Smoke test

Open your frontend URL in a browser, upload a `.txt` WhatsApp export, fill in
the form, and submit. The first request after a cold start takes ~5–15 seconds
(Pinecone index connection + OpenAI call). Subsequent warm requests are faster.

---

## Custom domains (optional)

To use `api.giftsense.com` and `giftsense.com`:

1. **Vercel Dashboard → [project] → Settings → Domains → Add Domain**
2. Add your domain and follow the DNS verification steps (Vercel provides the
   exact records to add in your DNS provider)
3. After adding a custom domain to the backend, update `ALLOWED_ORIGINS` on the
   backend to the frontend custom domain, and `VITE_API_URL` on the frontend to
   the backend custom domain
4. Redeploy both projects

---

## Free tier limits and cold starts

| Limit | Hobby Plan |
|-------|-----------|
| Serverless function invocations | 1,000,000 / month |
| Function max duration | 60 s (conservative setting in `vercel.json`) |
| Function memory | 1,024 MB |
| Bandwidth | 100 GB / month |
| Static site (frontend) | Always-on CDN, no cold starts |
| Build minutes | 6,000 / month |

### Cold start behaviour

Vercel spins down idle function instances automatically. When the backend has not
received traffic for a period, the next request triggers a cold start:

- `sync.Once` initialises the Gin engine, OpenAI clients, and Pinecone client
  (struct-only — no network call at init time)
- Typical cold start adds **1–3 seconds** before the actual request executes
- The frontend's loading spinner remains visible throughout — no special handling
  needed in the UI

To reduce cold starts on the free tier, you can set up a cron-based health ping
(e.g., [UptimeRobot](https://uptimerobot.com) free tier pinging `/health` every
5 minutes). This keeps the function warm without any code changes.

---

## Local development (unchanged)

The Vercel entry point does not affect local development:

```bash
# Terminal 1 — backend (uses cmd/server/main.go, not api/index.go)
cd giftsense-backend
cp .env.example .env   # fill in your keys
go run ./cmd/server/

# Terminal 2 — frontend
cd giftsense-frontend
cp .env.example .env.local   # set VITE_API_URL=http://localhost:8080
npm run dev
```

Vercel CLI (`vercel dev`) can also be used to simulate the serverless environment
locally, but `go run` is simpler for day-to-day development.

---

## Troubleshooting

### Backend returns 500 on all requests
Check **Functions → Logs** in the Vercel dashboard. The startup error is logged
before the 500 is returned. Common causes:

| Log message | Fix |
|-------------|-----|
| `OPENAI_API_KEY environment variable is required` | Add the env var and redeploy |
| `PINECONE_API_KEY environment variable is required` | Add the env var and redeploy |
| `PINECONE_ENVIRONMENT environment variable is required` | Add the env var and redeploy |
| `failed to create Pinecone client` | Invalid API key format |

### CORS error in browser console
`ALLOWED_ORIGINS` must exactly match the frontend URL — no trailing slash, correct
scheme (`https://`). A mismatch causes the browser to block the response even
though the request reaches the backend.

### Frontend shows "Request failed" / network error
- Open browser DevTools → Network → check what URL the request goes to
- If it hits `localhost:8080`, the frontend was built before `VITE_API_URL` was
  set. Trigger a manual redeploy of the frontend project.
- If it hits the correct Vercel URL but returns a non-200, check the backend logs.

### "conversation has too few messages"
The `.txt` export must contain at least 5 parseable messages. Export the full chat
history from WhatsApp (not a snippet), without media filtering.

### Function timeout (60 s exceeded)
Unlikely under normal load, but if it occurs:
- Check OpenAI API status at [status.openai.com](https://status.openai.com)
- Check Pinecone status at [status.pinecone.io](https://status.pinecone.io)
- Increase `maxDuration` in `giftsense-backend/vercel.json` (up to 300 on Hobby
  with Fluid compute enabled)

---

## Automatic deploys

Every `git push` to `main` triggers a redeploy of both Vercel projects. Vercel
detects which project's Root Directory contains changed files and skips unchanged
projects automatically.

To disable auto-deploy: **Project → Settings → Git → Auto Deploy → Off**

---

## Environment variable reference

### Backend (`giftsense-backend`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | ✅ | — | OpenAI API key |
| `PINECONE_API_KEY` | ✅ | — | Pinecone API key |
| `PINECONE_ENVIRONMENT` | ✅ | — | Pinecone region (e.g. `us-east-1`) |
| `ALLOWED_ORIGINS` | ✅ | `http://localhost:5173` | Frontend URL for CORS |
| `PINECONE_INDEX_NAME` | | `giftsense` | Pinecone index name |
| `CHAT_MODEL` | | `gpt-4o` | OpenAI chat model |
| `EMBEDDING_MODEL` | | `text-embedding-3-small` | OpenAI embedding model |
| `EMBEDDING_DIMENSIONS` | | `1536` | Must match Pinecone index dimensions |
| `MAX_TOKENS` | | `1000` | Max tokens for LLM response |
| `TOP_K` | | `3` | Chunks retrieved per query |
| `NUM_RETRIEVAL_QUERIES` | | `4` | Parallel retrieval queries |
| `MAX_FILE_SIZE_BYTES` | | `2097152` | 2 MB upload limit |
| `MAX_PROCESSED_MESSAGES` | | `400` | Max messages sampled from chat |
| `CHUNK_WINDOW_SIZE` | | `8` | Sliding window size |
| `CHUNK_OVERLAP_SIZE` | | `3` | Overlap between chunks |
| `PORT` | | `8080` | Used by local `go run` only; ignored by Vercel |

### Frontend (`giftsense-frontend`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_URL` | ✅ | — | Backend Vercel URL (no trailing slash) |
