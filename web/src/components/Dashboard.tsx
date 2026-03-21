import React from 'react';

interface DashboardProps {
    user: { id: string; name: string; role: string };
}

const Dashboard: React.FC<DashboardProps> = ({ user }) => {
    return (
        <div className="dashboard">
            <h1>Добро пожаловать, {user.name}</h1>
            <div className="stats-grid">
                <div className="stat-card">
                    <h3>Биометрические данные</h3>
                    <p>Пульс: 72 уд/мин</p>
                    <p>Давление: 120/80</p>
                    <p>SpO2: 98%</p>
                </div>
                <div className="stat-card">
                    <h3>Текущая программа</h3>
                    <p>Неделя 3 из 12</p>
                    <p>Прогресс: 65%</p>
                </div>
                <div className="stat-card">
                    <h3>Достижения</h3>
                    <p>🏆 5 тренировок подряд</p>
                    <p>⭐ 1000 калорий сожжено</p>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;