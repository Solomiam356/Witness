import React from 'react'

export function MainLayout({ children }) {
    return (
        <div className="min-h-screen bg-gray-50 grid grid-cols-1 md:grid-cols-[200px_1fr_280px] gap-6 p-4 max-w-7xl mx-auto"> 

            <aside className="hidden md:block border-r pr-4"> 
                <h1 className="text-xl font-bold mb-6">Witness</h1>
                <nav className="flex flex-col gap-3">
                    <a href="/feed" className="font-medium">Feed</a>
                    <a href="/prayer-room" className="font-medium">Prayer Room</a>
                    <a href="/profile" className="font-medium">Profile</a>
                </nav>
            </aside>

            <main className="w-full max-w-2xl mx-auto">
                {children}
            </main>

            <aside className="hidden md:block border-l pr-4"> 
                <div className="p-4 bg-white border rounded-lg">
                    <h2 className="font-semibold text-sm mb-2">Daily verse</h2>
                    <p className="text-sm text-gray-600">"I'm the light of the world..."</p>
                </div>
            </aside>

        </div>
    );
}