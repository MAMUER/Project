-- Создание расширений
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица профилей
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    age INT,
    gender VARCHAR(10),
    height_cm INT,
    weight_kg DECIMAL(5,2),
    fitness_level VARCHAR(50),
    goals TEXT[],
    contraindications TEXT[],
    nutrition TEXT,
    sleep_hours REAL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица биометрических данных
CREATE TABLE IF NOT EXISTS biometric_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    device_type VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы для биометрических данных
CREATE INDEX IF NOT EXISTS idx_biometric_user_metric_time ON biometric_data(user_id, metric_type, timestamp);
CREATE INDEX IF NOT EXISTS idx_biometric_timestamp ON biometric_data(timestamp);

-- Индексы для пользователей
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Таблица программ тренировок
CREATE TABLE IF NOT EXISTS training_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_data JSONB NOT NULL,
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    start_date DATE,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'active'
);

-- Таблица выполнения тренировок (исправленная)
CREATE TABLE IF NOT EXISTS workout_completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    training_plan_id UUID REFERENCES training_plans(id) ON DELETE CASCADE,
    workout_id VARCHAR(100) NOT NULL,
    scheduled_date DATE DEFAULT CURRENT_DATE,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    feedback TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица достижений
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    criteria JSONB,
    icon_url VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица пользовательских достижений
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, achievement_id)
);

-- Индексы для тренировок
CREATE INDEX IF NOT EXISTS idx_training_plans_user ON training_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_completions_user ON workout_completions(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_completions_plan ON workout_completions(training_plan_id);

-- Базовые достижения
INSERT INTO achievements (id, name, description, criteria) VALUES
    (gen_random_uuid(), 'Первый шаг', 'Первая завершенная тренировка', '{"type": "workout_count", "threshold": 1}'),
    (gen_random_uuid(), 'Десятка', '10 завершенных тренировок', '{"type": "workout_count", "threshold": 10}'),
    (gen_random_uuid(), 'Полтинник', '50 завершенных тренировок', '{"type": "workout_count", "threshold": 50}'),
    (gen_random_uuid(), 'Сто дней', '100 дней активности', '{"type": "active_days", "threshold": 100}'),
    (gen_random_uuid(), 'Мастер спорта', '1000 завершенных тренировок', '{"type": "workout_count", "threshold": 1000}')
ON CONFLICT DO NOTHING;