import React, { useState, useEffect } from 'react'
import { MainLayout } from "./components/Layout/MainLayout"
import { AddTestimonyForm } from "./components/Feed/AddTestimonyForm"
import { TestimonyCard } from "./components/Feed/TestimonyCard"
import { AuthForm } from "./components/Auth/AuthForm"
import { apiFetch } from "./api/client"

function App() {
  const [testimonies, setTestimonies] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const [token, setToken] = useState(localStorage.getItem('token'))
  const [activeTab, setActiveTab] = useState('feed')

  useEffect(() => {
    if (activeTab !== 'feed') return

    async function loadTestimonies() {
      try {
        setLoading(true)
        setError(null)

        const res = await apiFetch('/testimonies/feed')
        setTestimonies(res.data || [])
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    loadTestimonies()
  }, [activeTab])

  const handleAuthSuccess = (newToken) => {
    localStorage.setItem('token', newToken)
    setToken(newToken)
    setActiveTab('feed')
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setActiveTab('auth')
  }

  const handleAddTestimony = async ({title, content}) => {
    if (!token) {
      alert("Для публікації свідчення потрібно увійти в акаунт!")
      setActiveTab('auth')
      return
    }

    try {
      const newTestimony = await apiFetch('/testimonies', {
        method: 'POST',
        body: JSON.stringify({ title, content }),
      })

      alert("Свідчення відправлено! Воно з'явиться у стрічці після автоматичної модерації.")

      if (newTestimony && newTestimony.id) {
        setTestimonies((prev) => [newTestimony, ...prev])
      }
    } catch (err) {
      alert('Не вдалося зберегти: ' + err.message)
    }
  }

  return (
    <MainLayout
      activeTab={activeTab}
      onTabChange={setActiveTab}
      isAuthenticated={!!token}
      onLogout={handleLogout}
      >
      {activeTab === 'feed' && ( <>
      <AddTestimonyForm onAdd={handleAddTestimony} />

      {loading && <p className="text-center py-4 text-gray-500">Завантаження свідчень...</p>}
      {error && <p className="text-center py-4 text-red-500">{error}</p>}

    {!loading && !error && testimonies.length === 0 && (
        <p className="text-center py-4 text-gray-400">Поки що немає свідчень. Станьте першим, хто поділиться своєю історією!</p>)}

      {!loading && !error && testimonies.map((item) => (
        <TestimonyCard
          key={item.id}
          title={item.title}
          author={item.display_name || "Анонім"}
          date={item.created_at ? new Date(item.created_at).toLocaleDateString("uk-UA") : "Сьогодні"}
          text={item.content}
        />
      ))}
      </>
  )}

      {activeTab === 'prayer-room' && (
        <div className="text-center py-10">
          <h2 className="text-2xl font-bold mb-4">Prayer Room</h2>
          <p className="text-gray-600">Анонімна кімната молитов (WebSockets скоро...)</p>
        </div>
      )}

      {activeTab === 'profile' && (
        <div className="text-center py-10">
          <h2 className="text-2xl font-bold mb-2">Profile</h2>
          {token ? (
            <button
              onClick={handleLogout}
              className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
            >
              Вийти з акаунту
            </button>
          ) : (
            <p className="text-gray-600">Увійдіть, щоб переглянути свій профіль.</p>
          )}
         </div>
      )}
         {activeTab === 'auth' && (
          <AuthForm onSuccess={handleAuthSuccess} />
        )}
        </MainLayout>
  )
 }

 export default App