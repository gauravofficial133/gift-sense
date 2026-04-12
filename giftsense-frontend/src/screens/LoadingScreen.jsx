import { useState, useEffect } from 'react'

const MESSAGES = {
  text: [
    'Analysing conversation…',
    'Finding what makes them special…',
    'Picking up on the little things…',
    'Almost there…',
  ],
  audio: [
    'Listening carefully…',
    'Understanding the conversation…',
    'Finding what makes them special…',
    'Almost ready…',
  ],
  song: [
    'Feeling the music…',
    'Reading between the lyrics…',
    'Capturing every emotion…',
    'Turning feelings into gifts…',
  ],
}

const SUBTEXTS = {
  text: 'This takes about 15–20 seconds',
  audio: 'This takes about 15–20 seconds',
  song: 'This takes about 15–20 seconds',
}

export default function LoadingScreen({ inputMode = 'text' }) {
  const messages = MESSAGES[inputMode] ?? MESSAGES.text
  const [index, setIndex] = useState(0)

  useEffect(() => {
    const id = setInterval(() => {
      setIndex(i => (i + 1) % messages.length)
    }, 2500)
    return () => clearInterval(id)
  }, [messages])

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4 px-4">
      <div className="w-10 h-10 border-4 border-orange-500 border-t-transparent rounded-full animate-spin" />
      <p className="text-gray-700 text-sm font-medium text-center animate-fade-in" key={index}>
        {messages[index]}
      </p>
      <p className="text-gray-400 text-xs text-center">{SUBTEXTS[inputMode] ?? SUBTEXTS.text}</p>
    </div>
  )
}
