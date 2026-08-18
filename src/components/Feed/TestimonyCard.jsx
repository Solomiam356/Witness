import React from 'react'
export function TestimonyCard({ author, date, text }) {
    return (
        <article className="border rounded-xl p-5 bg-white mb-4 flex flex-col gap-3">

            <header className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center font-bold text-gray-600">
                        {author[0]}
                    </div>
                    <div>
                        <h3 className="font-semibold text-sm">{author}</h3>
                        <time className="text-xs text-gray-500">{date}</time>
                    </div>
                </div>
            </header>

            <p className="text-sm text-gray-800 leading-relaxed">{text}</p>

            <footer className="border-t pt-3 flex items-center justify-between text-xs text-gray-500">
                <button className="hover:text-black">To pray</button>
                <button className="hover:text-black">Share</button>
            </footer>
        </article>
    );
}