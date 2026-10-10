import React, { useState } from 'react'

export function AddTestimonyForm({ onAdd }) {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()

    if (!title.trim() || !content.trim()) return

    try {
      setIsSubmitting(true)

      if (onAdd) {
        await onAdd({
          title: title.trim(),
          content: content.trim()
        })
      }

      setTitle('')
      setContent('')
    } catch (err) {
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="border border-gray-200 rounded-xl p-4 bg-white mb-6 shadow-sm">
      <form onSubmit={handleSubmit} className="flex flex-col gap-3">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Заголовок свідчення (наприклад, Божий захист у дорозі)..."
          className="w-full border p-3 rounded-md text-sm font-medium focus:outline-none focus:border-black transition"
        />
        
        <textarea
          rows={3}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Поділіться детальніше, що саме Бог зробив у вашому житті..."
          className="w-full border p-3 rounded-md text-sm resize-none focus:outline-none focus:border-black transition"
        />

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting || !title.trim() || !content.trim()}
            className="bg-black text-white px-5 py-2 rounded-md text-sm font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            {isSubmitting ? 'Публікація...' : 'Опублікувати'}
          </button>
        </div>
      </form>
    </div>
  )
}