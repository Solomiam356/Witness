const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api';

export async function apiFetch(endpoint, options = {}) {
    const res = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
        credentials: 'include',
    });

    if (!res.ok) {
        const error = await res.json().catch(() => ({message: 'Помилка мережі'}));
        throw new Error(error.message || 'Щось пішло не так');
    }

    return res.json();
}