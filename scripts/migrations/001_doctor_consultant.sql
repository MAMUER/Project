-- ==========================================
-- Doctor Consultant Module Migration
-- ==========================================

-- Таблица врачей
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    specialty VARCHAR(100), -- sports_medicine, rehabilitation, general_practice
    license_number VARCHAR(100) UNIQUE,
    phone VARCHAR(20),
    bio TEXT,
    rating DECIMAL(3,2) DEFAULT 0.0,
    consultation_count INT DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_doctors_email ON doctors(email);
CREATE INDEX IF NOT EXISTS idx_doctors_specialty ON doctors(specialty);

-- Таблица подписок на врача
CREATE TABLE IF NOT EXISTS doctor_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    plan_type VARCHAR(50) NOT NULL DEFAULT 'monthly', -- monthly, quarterly, yearly
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, doctor_id)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON doctor_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_doctor ON doctor_subscriptions(doctor_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON doctor_subscriptions(user_id, is_active);

-- Назначения врача (рецепты, рекомендации)
CREATE TABLE IF NOT EXISTS doctor_prescriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    prescription_type VARCHAR(50) NOT NULL, -- recommendation, diet_change, training_change, medication
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal', -- low, normal, high, urgent
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, completed, cancelled
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prescriptions_user ON doctor_prescriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_prescriptions_doctor ON doctor_prescriptions(doctor_id);
CREATE INDEX IF NOT EXISTS idx_prescriptions_status ON doctor_prescriptions(status);

-- Таблица чата между пользователем и врачом
CREATE TABLE IF NOT EXISTS consultation_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL, -- user_id или doctor_id
    sender_type VARCHAR(20) NOT NULL, -- 'user' или 'doctor'
    message TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'text', -- text, image, file, system
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_user ON consultation_messages(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_doctor ON consultation_messages(doctor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON consultation_messages(user_id, doctor_id, is_read);

-- История изменений тренировок врачом
CREATE TABLE IF NOT EXISTS doctor_training_modifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    training_plan_id UUID NOT NULL REFERENCES training_plans(id) ON DELETE CASCADE,
    modification_type VARCHAR(50) NOT NULL, -- exercise_change, intensity_change, schedule_change, diet_change
    old_value JSONB,
    new_value JSONB,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_modifications_user ON doctor_training_modifications(user_id);
CREATE INDEX IF NOT EXISTS idx_modifications_doctor ON doctor_training_modifications(doctor_id);
CREATE INDEX IF NOT EXISTS idx_modifications_plan ON doctor_training_modifications(training_plan_id);

-- Консультации (сессии)
CREATE TABLE IF NOT EXISTS consultations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled', -- scheduled, in_progress, completed, cancelled
    scheduled_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consultations_user ON consultations(user_id);
CREATE INDEX IF NOT EXISTS idx_consultations_doctor ON consultations(doctor_id);
CREATE INDEX IF NOT EXISTS idx_consultations_status ON consultations(status);

-- ==========================================
-- Seed data: базовые врачи для тестирования
-- ==========================================

INSERT INTO doctors (id, email, full_name, specialty, license_number, bio, rating) VALUES
    (gen_random_uuid(), 'dr.smirnov@fitpulse.ru', 'Смирнов Андрей Петрович', 'sports_medicine', 'LIC-001', 
     'Спортивный врач с 15-летним опытом. Специализация: реабилитация после травм, подготовка спортсменов.', 4.8),
    (gen_random_uuid(), 'dr.ivanova@fitpulse.ru', 'Иванова Мария Сергеевна', 'rehabilitation', 'LIC-002',
     'Врач-реабилитолог. Специализация: восстановление после операций, коррекция осанки.', 4.9),
    (gen_random_uuid(), 'dr.kozlov@fitpulse.ru', 'Козлов Дмитрий Иванович', 'general_practice', 'LIC-003',
     'Врач общей практики. Специализация: профилактическая медицина, фитнес-консультации.', 4.7)
ON CONFLICT DO NOTHING;

-- ==========================================
-- Функция для проверки активной подписки
-- ==========================================

CREATE OR REPLACE FUNCTION has_active_subscription(p_user_id UUID, p_doctor_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM doctor_subscriptions
        WHERE user_id = p_user_id
          AND doctor_id = p_doctor_id
          AND is_active = TRUE
          AND expires_at > NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- Триггер для обновления updated_at
-- ==========================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_doctors_updated_at
    BEFORE UPDATE ON doctors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_prescriptions_updated_at
    BEFORE UPDATE ON doctor_prescriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
