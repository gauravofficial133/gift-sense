import { useState } from 'react'
import AssetBrowser from './templates/asset-bot/AssetBrowser'
import AssetBotPanel from './templates/asset-bot/AssetBotPanel'

export default function AssetPage() {
  const [tab, setTab] = useState('browse')

  return (
    <div className="p-8">
      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setTab('browse')}
          className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
            tab === 'browse' ? 'bg-gray-800 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          Browse
        </button>
        <button
          onClick={() => setTab('generate')}
          className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
            tab === 'generate' ? 'bg-purple-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          Generate
        </button>
      </div>

      {tab === 'browse' ? <AssetBrowser /> : <AssetBotPanel />}
    </div>
  )
}
