-- ==========================================
-- Invitation Codes System
-- Отдельная регистрация для врачей и админов через invite-коды
-- ==========================================

-- Таблица invite-кодов
CREATE TABLE IF NOT EXISTS invite_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL, -- 'doctor', 'admin'
    specialty VARCHAR(100),    -- только для doctors
    max_uses INT NOT NULL DEFAULT 1,
    used_count INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_codes_code ON invite_codes(code);
CREATE INDEX IF NOT EXISTS idx_invite_codes_role ON invite_codes(role);
CREATE INDEX IF NOT EXISTS idx_invite_codes_active ON invite_codes(is_active);

-- Функция для валидации и использования invite-кода
CREATE OR REPLACE FUNCTION use_invite_code(p_code VARCHAR(50))
RETURNS TABLE (
    is_valid BOOLEAN,
    role VARCHAR(50),
    specialty VARCHAR(100),
    error_message TEXT
) AS $$
DECLARE
    v_code_record RECORD;
BEGIN
    -- Ищем код
    SELECT * INTO v_code_record
    FROM invite_codes
    WHERE code = p_code AND is_active = TRUE;

    IF NOT FOUND THEN
        RETURN QUERY SELECT FALSE, NULL::VARCHAR, NULL::VARCHAR, 'Недействительный код приглашения';
        RETURN;
    END IF;

    -- Проверяем срок действия
    IF v_code_record.expires_at IS NOT NULL AND v_code_record.expires_at < NOW() THEN
        RETURN QUERY SELECT FALSE, NULL::VARCHAR, NULL::VARCHAR, 'Срок действия кода истёк';
        RETURN;
    END IF;

    -- Проверяем лимит использований
    IF v_code_record.used_count >= v_code_record.max_uses THEN
        RETURN QUERY SELECT FALSE, NULL::VARCHAR, NULL::VARCHAR, 'Код достиг лимит использований';
        RETURN;
    END IF;

    -- Увеличиваем счётчик
    UPDATE invite_codes
    SET used_count = used_count + 1
    WHERE id = v_code_record.id;

    RETURN QUERY SELECT TRUE, v_code_record.role, v_code_record.specialty, ''::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- Seed: базовые invite-коды для тестирования
-- ==========================================

-- Код для регистрации врача (можно использовать 1 раз)
INSERT INTO invite_codes (code, role, specialty, max_uses, expires_at) VALUES
    ('DOCTOR-2026-SPORT-MED-001', 'doctor', 'sports_medicine', 1, NOW() + INTERVAL '1 year')
ON CONFLICT (code) DO NOTHING;

INSERT INTO invite_codes (code, role, specialty, max_uses, expires_at) VALUES
    ('DOCTOR-2026-REHAB-002', 'doctor', 'rehabilitation', 1, NOW() + INTERVAL '1 year')
ON CONFLICT (code) DO NOTHING;

INSERT INTO invite_codes (code, role, specialty, max_uses, expires_at) VALUES
    ('DOCTOR-2026-GENERAL-003', 'doctor', 'general_practice', 5, NOW() + INTERVAL '1 year')
ON CONFLICT (code) DO NOTHING;

-- Код для регистрации админа (можно использовать только 1 раз)
INSERT INTO invite_codes (code, role, max_uses, expires_at) VALUES
    ('ADMIN-2026-SETUP-ROOT-001', 'admin', 1, NOW() + INTERVAL '6 months')
ON CONFLICT (code) DO NOTHING;

-- ==========================================
-- Функция для создания новых invite-кодов (только для админов)
-- ==========================================

CREATE OR REPLACE FUNCTION create_invite_code(
    p_role VARCHAR(50),
    p_specialty VARCHAR(100) DEFAULT NULL,
    p_max_uses INT DEFAULT 1,
    p_expires_days INT DEFAULT 30
)
RETURNS VARCHAR(50) AS $$
DECLARE
    v_code VARCHAR(50);
BEGIN
    -- Генерируем код: ROLE-YYYY-RANDOM
    v_code := UPPER(p_role) || '-' || EXTRACT(YEAR FROM NOW()) || '-' || 
              MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT);
    v_code := SUBSTRING(v_code, 1, 50);

    INSERT INTO invite_codes (code, role, specialty, max_uses, expires_at)
    VALUES (v_code, p_role, p_specialty, p_max_uses, NOW() + (p_expires_days || ' days')::INTERVAL);

    RETURN v_code;
END;
$$ LANGUAGE plpgsql;
