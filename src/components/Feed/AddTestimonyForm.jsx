import React, { useState } from 'react'

export function AddTestimonyForm({ onAdd }) {
  const [text, setText] = useState('')

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!text.trim()) return
    if (onAdd) onAdd(text)
    setText('')
  }

  return (
    <div className="border rounded-xl p-4 bg-white mb-6">
      <form onSubmit={handleSubmit} className="flex flex-col gap-3">
        <textarea
          rows={3}
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Share what God has done in your life..."
          className="w-full border p-3 rounded-md text-sm resize-none focus:outline-none"
        />
        <div className="flex justify-end">
          <button 
            type="submit"
            className="bg-black text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-gray-800 transition"
          >
            Publish
          </button>
        </div>
      </form>
    </div>
  )
}