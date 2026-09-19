import React, { useState, useEffect } from 'react'
import { MainLayout } from "./components/Layout/MainLayout"
import { AddTestimonyForm } from "./components/Feed/AddTestimonyForm"
import { TestimonyCard } from "./components/Feed/TestimonyCard"
import { apiFetch } from "./api/client"

function App() {
  const [testimonies, setTestimonies] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function loadTestimonies() {
      try {
        setLoading(true)
        const data = await apiFetch('/testimonies/feed')
        setTestimonies(Array.isArray(data) ? data : [])
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    loadTestimonies()
  }, [])

  const handleAddTestimony = async (text) => {
    try {
      const newTestimony = await apiFetch('/testimonies', {
        method: 'POST',
        body: JSON.stringify({ content: text }),
      })

      setTestimonies((prev) => [newTestimony, ...prev])
    } catch (err) {
      alert('Не вдалося зберегти: ' + err.message)
    }
  }

  return (
    <MainLayout>
      <AddTestimonyForm onAdd={handleAddTestimony} />

      {loading && <p className="text-center py-4 text-gray-500">Завантаження...</p>}
      {error && <p className="text-center py-4 text-red-500">{error}</p>}

      {!loading && !error && testimonies.map((item) => (
        <TestimonyCard
          key={item.id}
          author={item.author || "Анонім"}
          date={item.created_at ? new Date(item.created_at).toLocaleDateString("uk-UA") : "Сьогодні"}
          text={item.content}
        />
      ))}
    </MainLayout>
  )
}

export default App