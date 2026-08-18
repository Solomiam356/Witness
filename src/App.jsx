import React from 'react'
import {MainLayout} from "./components/Layout/MainLayout"
import {AddTestimonyForm} from "./components/Feed/AddTestimonyForm"
import {TestimonyCard} from "./components/Feed/TestimonyCard"

function App() {
    return (
        <MainLayout>
            <AddTestimonyForm />
            <TestimonyCard
              author="Марія" 
              date="18 серпня 2026" 
              text="Сьогодні Бог чудовим чином допоміг мені скласти іспит..."
              />
             <TestimonyCard
              author="Інна" 
              date="15 серпня 2026" 
              text="Дякую Богу що береже мою родину."
              />
                <TestimonyCard
              author="Юлія" 
              date="12 серпня 2026" 
              text="Я вдячна Богу що маю все те що 5 років тому я навіть і не мріяла..."
              />
        </MainLayout>
    )
}
export default App;