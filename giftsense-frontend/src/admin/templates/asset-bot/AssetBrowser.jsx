import { useState, useEffect } from 'react'
import { Search } from 'lucide-react'
import { assetApi } from '../../../api/adminApi'

export default function AssetBrowser() {
  const [assets, setAssets] = useState([])
  const [style, setStyle] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    assetApi.list([], style).then(d => setAssets(d.assets || [])).finally(() => setLoading(false))
  }, [style])

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <h2 className="text-base font-bold text-gray-900">Asset Library</h2>
        <select value={style} onChange={e => setStyle(e.target.value)} className="rounded-lg border border-gray-200 px-2 py-1 text-xs">
          <option value="">All styles</option>
          {['watercolor', 'flat', '3d', 'hand-drawn', 'geometric', 'botanical'].map(s => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      {loading ? (
        <p className="text-sm text-gray-400">Loading...</p>
      ) : assets.length === 0 ? (
        <p className="text-sm text-gray-400">No assets found. Generate some using the Asset Bot.</p>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
          {assets.map(asset => (
            <div key={asset.id} className="bg-white rounded-lg border border-gray-200 p-2 hover:shadow-md transition-shadow">
              <div className="aspect-square bg-gray-50 rounded flex items-center justify-center mb-2 overflow-hidden">
                {asset.thumbnail_b64 ? (
                  <img src={`data:image/png;base64,${asset.thumbnail_b64}`} alt={asset.id} className="w-full h-full object-cover" />
                ) : (
                  <span className="text-xs text-gray-300">No preview</span>
                )}
              </div>
              <p className="text-[10px] text-gray-600 truncate font-medium">{asset.id}</p>
              <p className="text-[10px] text-gray-400">{asset.style || 'unknown'}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
