'use client'

import { useState } from 'react'
import { UserTab } from './user-tab'
import { ApiKeysTab } from './api-keys-tab'

export default function ProfilePage() {
    const [activeTab, setActiveTab] = useState<'user' | 'api-keys'>('user')

    return (
        <div className="container max-w-screen-lg py-10">
            <h1 className="mb-8 text-3xl font-bold">Profile</h1>

            <div className="mb-8 border-b">
                <nav className="-mb-px flex space-x-8" aria-label="Tabs">
                    <button
                        onClick={() => setActiveTab('user')}
                        className={`
              whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium transition-colors
              ${activeTab === 'user'
                                ? 'border-primary text-foreground'
                                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'}
            `}
                    >
                        User
                    </button>
                    <button
                        onClick={() => setActiveTab('api-keys')}
                        className={`
              whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium transition-colors
              ${activeTab === 'api-keys'
                                ? 'border-primary text-foreground'
                                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'}
            `}
                    >
                        User API keys
                    </button>
                </nav>
            </div>

            {activeTab === 'user' ? <UserTab /> : <ApiKeysTab />}
        </div>
    )
}
