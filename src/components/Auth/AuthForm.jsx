import React from 'react'
export function AuthForm() {
    return (
        <div className="min-h-[80vh] flex items-center justify-center">
            <main className="w-full max-w-md border rounded-xl p-6 bg-white shadow-sm">
                <h2 className="text-2xl font-bold text-center mb-6">Log in to Witness</h2>

                <form className="flex flex-col gap-4">

                    <div className="flex flex-col gap-1">
                        <label htmlFor="email" className="text-sm font-medium">Email</label>
                        <input 
                        id="email"
                        type="email"
                        required
                        placeholder="name@example.com"
                        className="border p-2 rounded-md w-full text-sm"
                        />
                    </div>

                    <div className="flex flex-col gap-1">
                      <label htmlFor="password" className="text-sm font-medium">Password</label>
                        <input 
                        id="password"
                        type="password"
                        required
                        placeholder="********"
                        className="border p-2 rounded-md w-full text-sm"
                        />
                    </div>

                    <button 
                    type="submit"
                    className="w-full bg-black text-white py-2 rounded-md font-medium text-sm hover:bg-gray-800 transition">Sign in</button>
                </form>
            </main>
        </div>
        
    );
}