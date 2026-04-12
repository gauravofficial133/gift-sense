import { Search, X, Music } from 'lucide-react'
import { useSpotifySearch } from '../../hooks/useSpotifySearch'

export default function SpotifySongPicker({ selectedTrack, onSelect, onClear }) {
  const { query, results, isSearching, error, search, clear } = useSpotifySearch()

  if (selectedTrack) {
    return (
      <div className="flex flex-col gap-3">
        {/* Selected track card */}
        <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-gray-50 p-3">
          {selectedTrack.albumArt ? (
            <img
              src={selectedTrack.albumArt}
              alt=""
              className="w-12 h-12 rounded object-cover"
            />
          ) : (
            <div className="w-12 h-12 rounded bg-gray-200 flex items-center justify-center">
              <Music className="w-5 h-5 text-gray-400" />
            </div>
          )}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">{selectedTrack.trackName}</p>
            <p className="text-xs text-gray-500 truncate">{selectedTrack.artist}</p>
          </div>
          <button
            type="button"
            onClick={() => { onClear(); clear() }}
            className="shrink-0 p-1.5 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Spotify embed player */}
        <div className="rounded-lg overflow-hidden">
          <iframe
            src={`https://open.spotify.com/embed/track/${selectedTrack.trackId}?utm_source=generator&theme=0`}
            width="100%"
            height="152"
            frameBorder="0"
            allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
            loading="lazy"
            title={`${selectedTrack.trackName} by ${selectedTrack.artist}`}
          />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {/* Search input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          type="text"
          value={query}
          onChange={e => search(e.target.value)}
          placeholder="Search for a song..."
          className="w-full rounded-lg border border-gray-200 bg-white pl-9 pr-3 py-2.5 text-sm text-gray-900
            placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
        />
        {isSearching && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            <div className="w-4 h-4 border-2 border-gray-300 border-t-orange-500 rounded-full animate-spin" />
          </div>
        )}
      </div>

      {/* Error */}
      {error && <p className="text-xs text-red-500">{error}</p>}

      {/* Results list */}
      {results.length > 0 && (
        <div className="flex flex-col rounded-lg border border-gray-200 divide-y divide-gray-100 overflow-hidden">
          {results.map(track => (
            <button
              key={track.id}
              type="button"
              onClick={() => onSelect({
                trackId: track.id,
                trackName: track.name,
                artist: track.artist,
                albumArt: track.album_art,
              })}
              className="flex items-center gap-3 p-3 text-left hover:bg-gray-50 transition-colors"
            >
              {track.album_art ? (
                <img src={track.album_art} alt="" className="w-10 h-10 rounded object-cover shrink-0" />
              ) : (
                <div className="w-10 h-10 rounded bg-gray-100 flex items-center justify-center shrink-0">
                  <Music className="w-4 h-4 text-gray-400" />
                </div>
              )}
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-gray-900 truncate">{track.name}</p>
                <p className="text-xs text-gray-500 truncate">{track.artist}</p>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* Empty state */}
      {query && !isSearching && results.length === 0 && !error && (
        <p className="text-xs text-gray-400 text-center py-4">No tracks found. Try a different search.</p>
      )}
    </div>
  )
}
