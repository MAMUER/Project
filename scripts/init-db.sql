-- Создание расширений
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'client', -- client, doctor, admin
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица профилей (дополнительная информация, цели, противопоказания)
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    age INT,
    gender VARCHAR(10),
    height_cm INT,
    weight_kg DECIMAL(5,2),
    fitness_level VARCHAR(50), -- beginner, intermediate, advanced
    goals TEXT[], -- массив целей (weight_loss, muscle_gain, endurance, etc)
    contraindications TEXT[], -- массив противопоказаний
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
) PARTITION BY RANGE (timestamp);

-- Создание партиций (для примера, можно автоматизировать)
CREATE TABLE IF NOT EXISTS biometric_data_2025_01 PARTITION OF biometric_data
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE IF NOT EXISTS biometric_data_2025_02 PARTITION OF biometric_data
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- и так далее, но в реальности нужно автоматическое создание

-- Индексы
CREATE INDEX IF NOT EXISTS idx_biometric_user_metric_time ON biometric_data(user_id, metric_type, timestamp);

-- Таблица программ тренировок
CREATE TABLE IF NOT EXISTS training_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_data JSONB NOT NULL, -- структура программы (недели, дни, упражнения)
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    start_date DATE,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'active', -- active, completed, archived
    metadata JSONB -- дополнительные данные (класс тренировки, confidence и т.д.)
);

-- Таблица выполнения тренировок (прогресс)
CREATE TABLE IF NOT EXISTS workout_completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    training_plan_id UUID REFERENCES training_plans(id) ON DELETE CASCADE,
    scheduled_date DATE NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    feedback JSONB, -- самооценка, ощущения
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица достижений (ачивки)
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    criteria JSONB, -- условия получения
    icon_url VARCHAR(255)
);

-- Таблица пользовательских достижений
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, achievement_id)
);