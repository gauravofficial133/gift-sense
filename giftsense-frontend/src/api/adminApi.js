const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

async function request(path, options = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

export const templateApi = {
  list: () => request('/api/v1/admin/templates'),
  get: (id) => request(`/api/v1/admin/templates/${id}`),
  create: (tpl) => request('/api/v1/admin/templates', { method: 'POST', body: JSON.stringify(tpl) }),
  update: (id, tpl) => request(`/api/v1/admin/templates/${id}`, { method: 'PUT', body: JSON.stringify(tpl) }),
  delete: (id) => request(`/api/v1/admin/templates/${id}`, { method: 'DELETE' }),
  preview: (id) => request(`/api/v1/admin/templates/${id}/preview`, { method: 'POST' }),
  duplicate: (id) => request(`/api/v1/admin/templates/${id}/duplicate`, { method: 'POST' }),
}

export const assetApi = {
  list: (tags, style) => {
    const params = new URLSearchParams()
    if (tags?.length) tags.forEach(t => params.append('tags', t))
    if (style) params.set('style', style)
    return request(`/api/v1/admin/assets?${params}`)
  },
  planPrompt: (req) => request('/api/v1/admin/assets/plan-prompt', { method: 'POST', body: JSON.stringify(req) }),
  generate: (req) => request('/api/v1/admin/assets/generate', { method: 'POST', body: JSON.stringify(req) }),
  upload: (req) => request('/api/v1/admin/assets/upload', { method: 'POST', body: JSON.stringify(req) }),
}
