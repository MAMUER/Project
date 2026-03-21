import React from 'react';
import Dashboard from '../components/Dashboard';
import { useAuth } from '../hooks/useAuth';

const DashboardPage: React.FC = () => {
    const { user } = useAuth();

    if (!user) {
        return <div>Загрузка...</div>;
    }

    return (
        <div className="dashboard-page">
            <header>
                <h1>HealthFit Dashboard</h1>
                <div className="user-info">{user.email}</div>
            </header>
            <main>
                <Dashboard user={{ id: user.id, name: user.email, role: user.role }} />
            </main>
        </div>
    );
};

export default DashboardPage;