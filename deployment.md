# GiftSense — Deployment Guide (Render.com)

GiftSense deploys as two services on Render:
- **giftsense-backend** — Go/Gin web service (REST API)
- **giftsense-frontend** — React/Vite static site (CDN-hosted)

Both are defined in `render.yaml` at the project root using Render Blueprints (IaC).

---

## Prerequisites

Before deploying, make sure you have:

| Requirement | Details |
|-------------|---------|
| Render account | Free tier is fine — [render.com](https://render.com) |
| GitHub repo | Push this project to a GitHub repository |
| OpenAI API key | Needs access to `gpt-4o` and `text-embedding-3-small` |
| Pinecone account | Free tier works — [pinecone.io](https://pinecone.io) |
| Pinecone index | Created **before** first deploy (see below) |

---

## Step 1 — Create the Pinecone Index

GiftSense uses Pinecone for vector storage. Create the index manually before deploying:

1. Log in to [app.pinecone.io](https://app.pinecone.io)
2. Click **Create Index**
3. Configure with these exact settings:

   | Setting | Value |
   |---------|-------|
   | Index name | `giftsense` |
   | Dimensions | `1536` |
   | Metric | `Cosine` |
   | Type | `Serverless` |
   | Cloud | `AWS` |
   | Region | `us-east-1` |

4. Copy your **API key** from the Pinecone dashboard (you'll need it in Step 3)
5. Note your **environment** — for serverless it is `us-east-1`

---

## Step 2 — Push to GitHub

```bash
git remote add origin https://github.com/YOUR_USERNAME/gift-sense.git
git push -u origin main
```

The `render.yaml` at the project root must be committed and visible on GitHub.

---

## Step 3 — Deploy via Render Blueprint

Render Blueprints read `render.yaml` and deploy all services in one shot.

1. Log in to [dashboard.render.com](https://dashboard.render.com)
2. Click **New → Blueprint**
3. Connect your GitHub account if prompted, then select the **gift-sense** repository
4. Click **Connect**
5. Set a **Blueprint name** (e.g. `giftsense`)
6. Leave branch as `main`
7. Click **Apply** — Render will detect both services from `render.yaml`
8. You will be prompted to fill in the `sync: false` environment variables (the ones without values in `render.yaml`). Enter placeholder values for now and click **Apply** — you will update the real values in Step 4

> **Why placeholders first?** The two services need to know each other's URLs (`ALLOWED_ORIGINS` and `VITE_API_URL`), which only exist after the first deploy completes. Deploy with placeholders, get the URLs, then update both services.

Render will now build both services. The build takes 2–5 minutes:
- Backend: runs `go build -o server ./cmd/server`
- Frontend: runs `npm install && npm run build`, publishes `dist/`

---

## Step 4 — Set Secret Environment Variables

Once both services are deployed, you will have two `.onrender.com` URLs. Now set the real values.

### Backend service (`giftsense-backend`)

Go to **Dashboard → giftsense-backend → Environment**:

| Variable | Value |
|----------|-------|
| `OPENAI_API_KEY` | Your OpenAI API key (`sk-...`) |
| `PINECONE_API_KEY` | Your Pinecone API key |
| `PINECONE_ENVIRONMENT` | `us-east-1` |
| `ALLOWED_ORIGINS` | The frontend URL, e.g. `https://giftsense-frontend.onrender.com` |

Click **Save Changes** — Render will redeploy the backend automatically.

### Frontend service (`giftsense-frontend`)

Go to **Dashboard → giftsense-frontend → Environment**:

| Variable | Value |
|----------|-------|
| `VITE_API_URL` | The backend URL, e.g. `https://giftsense-backend.onrender.com` |

Click **Save Changes** — Render will rebuild and redeploy the frontend.

> **Important:** `VITE_API_URL` is a build-time variable (Vite embeds it at build time). Changing it always triggers a full rebuild of the frontend.

---

## Step 5 — Verify the Deployment

### Health check
```bash
curl https://giftsense-backend.onrender.com/health
# Expected: {"status":"ok"}
```

### CORS check
```bash
curl -H "Origin: https://giftsense-frontend.onrender.com" \
     -I https://giftsense-backend.onrender.com/health
# Expected headers include:
# Access-Control-Allow-Origin: https://giftsense-frontend.onrender.com
```

### End-to-end smoke test
Open `https://giftsense-frontend.onrender.com` in your browser, upload a `.txt` WhatsApp export, fill in the form, and submit. The first request after a cold start may take ~60 seconds (see Free Tier section below).

---

## Free Tier Limitations

| Limitation | Impact |
|------------|--------|
| **Cold starts** | Free web services spin down after **15 minutes of inactivity**. The next request takes ~60 seconds to wake the backend. The frontend (static site) is always fast — only the API has cold starts. |
| **750 instance hours/month** per workspace | At one backend instance, that's ~31 days of continuous uptime. Spun-down time doesn't count. |
| **Single instance** | No horizontal scaling on the free plan. |
| **Ephemeral filesystem** | GiftSense has no local file storage — this is not an issue since all data goes to Pinecone and OpenAI. |
| **Static site** | Deployed for free with global CDN — no cold starts. |

### Handling cold starts gracefully
The frontend's loading spinner remains visible while the backend wakes up — the `useAnalyze` hook has no timeout, so it will wait and then display results. No special handling needed.

To avoid cold starts entirely, upgrade the backend to a **Starter** plan ($7/month), which keeps the instance always-on.

---

## Automatic Deploys

After setup, every push to `main` that changes `render.yaml` triggers a Blueprint sync. Every push (regardless of `render.yaml`) triggers a redeploy of both services via Render's auto-deploy feature.

To disable auto-deploy for a service: **Dashboard → [service] → Settings → Auto-Deploy → Off**.

---

## Troubleshooting

### Backend won't start
Check **Dashboard → giftsense-backend → Logs**. Common causes:
- Missing required env var → look for `config: ... is required` in logs
- `OPENAI_API_KEY` not set → `config: OPENAI_API_KEY is required`
- `PINECONE_API_KEY` not set → `config: PINECONE_API_KEY is required`
- `PINECONE_ENVIRONMENT` not set → `config: PINECONE_ENVIRONMENT is required`

### CORS errors in browser console
`ALLOWED_ORIGINS` on the backend must match the frontend URL **exactly** (no trailing slash):
```
# Correct
ALLOWED_ORIGINS=https://giftsense-frontend.onrender.com

# Wrong (trailing slash causes mismatch)
ALLOWED_ORIGINS=https://giftsense-frontend.onrender.com/
```

### Frontend shows "Request failed" or network error
- Confirm `VITE_API_URL` is set to the backend URL (no trailing slash)
- Confirm the backend health check returns 200
- Check the browser Network tab — if the request goes to `localhost:8080`, the frontend was built before `VITE_API_URL` was set. Trigger a manual redeploy of the frontend.

### Pinecone errors
- Confirm the index name in Render matches `PINECONE_INDEX_NAME` (default: `giftsense`)
- Confirm dimensions are exactly `1536` (must match `EMBEDDING_DIMENSIONS`)
- Confirm the region matches `PINECONE_ENVIRONMENT`

### Conversation too short error
The file must contain at least 5 parseable messages. Export the full WhatsApp chat (not just a few lines) as a `.txt` file without media.

---

## Custom Domain (Optional)

To use a custom domain (e.g. `giftsense.yourdomain.com`):

1. Go to **Dashboard → giftsense-frontend → Settings → Custom Domains**
2. Add your domain and follow the DNS verification steps
3. Update `ALLOWED_ORIGINS` on the backend to your custom domain
4. Update `VITE_API_URL` on the frontend if you also set a custom domain for the backend

---

## Environment Variable Reference

### Backend (`giftsense-backend`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | ✅ | — | OpenAI API key |
| `PINECONE_API_KEY` | ✅ | — | Pinecone API key |
| `PINECONE_ENVIRONMENT` | ✅ | — | Pinecone region (e.g. `us-east-1`) |
| `ALLOWED_ORIGINS` | ✅ | — | Frontend URL for CORS (comma-separated for multiple) |
| `PINECONE_INDEX_NAME` | | `giftsense` | Must match your Pinecone index name |
| `CHAT_MODEL` | | `gpt-4o` | OpenAI chat model |
| `EMBEDDING_MODEL` | | `text-embedding-3-small` | OpenAI embedding model |
| `EMBEDDING_DIMENSIONS` | | `1536` | Must match Pinecone index dimensions |
| `MAX_TOKENS` | | `1000` | Max tokens for LLM response |
| `TOP_K` | | `3` | Chunks retrieved per query |
| `NUM_RETRIEVAL_QUERIES` | | `4` | Parallel retrieval queries |
| `MAX_FILE_SIZE_BYTES` | | `2097152` | 2 MB upload limit |
| `MAX_PROCESSED_MESSAGES` | | `400` | Max messages sampled from chat |
| `CHUNK_WINDOW_SIZE` | | `8` | Sliding window size for chunking |
| `CHUNK_OVERLAP_SIZE` | | `3` | Overlap between chunks |
| `PORT` | | `8080` | Set automatically by Render |

### Frontend (`giftsense-frontend`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_URL` | ✅ | — | Backend service URL (no trailing slash) |
