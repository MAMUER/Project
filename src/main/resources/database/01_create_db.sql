DROP SCHEMA IF EXISTS fitness_club_db CASCADE;

CREATE SCHEMA IF NOT EXISTS fitness_club_db;

SET search_path TO fitness_club_db;

-- Создание пользователя
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'springroot') THEN
        CREATE USER springroot WITH PASSWORD 'spring';
    END IF;
END
$$;

GRANT ALL PRIVILEGES ON DATABASE fitness_club_db TO springroot;

-- 1. ТАБЛИЦА ТИПОВ ЗАЛОВ (НОВАЯ)
CREATE TABLE gym_types (
    gym_type_id SERIAL PRIMARY KEY NOT NULL,
    type_name VARCHAR(100) NOT NULL,
    description VARCHAR(200)
);

-- 2. ТАБЛИЦА КЛУБОВ (ОБНОВЛЕНА)
CREATE TABLE clubs (
    club_name VARCHAR(45) PRIMARY KEY NOT NULL,
    address VARCHAR(45) NOT NULL,
    schedule JSONB
);

-- 3. ТАБЛИЦА ДОЛЖНОСТЕЙ
CREATE TABLE position(
    id_position SERIAL PRIMARY KEY NOT NULL,
    role_name VARCHAR(45) NOT NULL
);

-- 5. ТАБЛИЦА ЧЛЕНОВ КЛУБА (обновленная)
CREATE TABLE members (
    id_member SERIAL PRIMARY KEY NOT NULL,
    club_name VARCHAR(45),
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    birth_date DATE NOT NULL,
    gender INTEGER NOT NULL CHECK (gender IN (0, 1, 2)),
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE SET NULL ON UPDATE CASCADE
);

-- 6. ТАБЛИЦА ДОСТИЖЕНИЙ
CREATE TABLE achievements (
    id_achievement SERIAL PRIMARY KEY NOT NULL,
    achievement_description VARCHAR(45),
    achievement_title VARCHAR(45) NOT NULL,
    achievement_icon_url VARCHAR(100) NOT NULL
);

-- 7. ТАБЛИЦА СВЯЗИ ЧЛЕНОВ И ДОСТИЖЕНИЙ
CREATE TABLE members_have_achievements (
    id_member INTEGER NOT NULL,
    id_achievement INTEGER NOT NULL,
    receipt_date DATE NOT NULL,
    PRIMARY KEY (id_member, id_achievement),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_achievement) REFERENCES achievements (id_achievement) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 8. ТАБЛИЦА INBODY АНАЛИЗОВ
CREATE TABLE inbody_analysis (
    id_inbody_analysis SERIAL PRIMARY KEY NOT NULL,
    HEIGHT FLOAT,
    weight FLOAT,
    bmi FLOAT,
    fat_percent FLOAT,
    muscle_percent FLOAT
);

-- 9. ТАБЛИЦА СВЯЗИ ЧЛЕНОВ И INBODY АНАЛИЗОВ
CREATE TABLE members_have_inbody_analysis (
    id_member INTEGER NOT NULL,
    id_inbody_analysis INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_inbody_analysis),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_inbody_analysis) REFERENCES inbody_analysis (id_inbody_analysis) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 10. ТАБЛИЦА ИСТОРИИ ПОСЕЩЕНИЙ
CREATE TABLE visits_history (
    id_visit SERIAL PRIMARY KEY NOT NULL,
    visit_date DATE
);

-- 11. ТАБЛИЦА СВЯЗИ ЧЛЕНОВ И ИСТОРИИ ПОСЕЩЕНИЙ
CREATE TABLE members_have_visits_history (
    id_member INTEGER NOT NULL,
    id_visit INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_visit),
    FOREIGN KEY (id_visit) REFERENCES visits_history (id_visit) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 12. ТАБЛИЦА ПЛАНОВ ПИТАНИЯ (ОБНОВЛЕНА)
CREATE TABLE nutrition_plan (
    id_plan SERIAL PRIMARY KEY NOT NULL,
    plan_name VARCHAR(100) NOT NULL,
    nutrition_description VARCHAR(200),
    goal_type VARCHAR(50), -- цель
    difficulty_level VARCHAR(20), -- легкий, средний, сложный
    calories_per_day INTEGER,
    protein_percent INTEGER,
    carbs_percent INTEGER,
    fat_percent INTEGER
);

-- 13. ТАБЛИЦА ТИПОВ ОБОРУДОВАНИЯ (ОБНОВЛЕНА)
CREATE TABLE equipment_type (
    id_equipment_type SERIAL PRIMARY KEY NOT NULL,
    type_name VARCHAR(45) NOT NULL
);

-- 14. ТАБЛИЦА ЗАЛОВ (ОБНОВЛЕНА)
CREATE TABLE gyms (
    id_gym SERIAL PRIMARY KEY NOT NULL,
    club_name VARCHAR(45) NOT NULL,
    gym_type_id INTEGER NOT NULL,
    capacity INTEGER NOT NULL,
    available_hours INTEGER NOT NULL,
    CONSTRAINT fk_gyms_club FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_gyms_gym_type FOREIGN KEY (gym_type_id) REFERENCES gym_types (gym_type_id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 15. ТАБЛИЦА ОБОРУДОВАНИЯ (ОБНОВЛЕНА)
CREATE TABLE equipment (
    id_equipment SERIAL PRIMARY KEY NOT NULL,
    id_equipment_type INTEGER,
    quantity INTEGER,
    club_name VARCHAR(45),
    FOREIGN KEY (id_equipment_type) REFERENCES equipment_type (id_equipment_type) ON DELETE SET NULL ON UPDATE CASCADE,
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE SET NULL ON UPDATE CASCADE
);

-- 16. ТАБЛИЦА СВЯЗИ ОБОРУДОВАНИЯ И КЛУБОВ (ОБНОВЛЕНА)
CREATE TABLE clubs_have_equipment (
    club_name VARCHAR(45) NOT NULL,
    id_equipment INTEGER NOT NULL,
    PRIMARY KEY (club_name, id_equipment),
    FOREIGN KEY (id_equipment) REFERENCES equipment (id_equipment) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 17. ТАБЛИЦА ТИПОВ ТРЕНИРОВОК
CREATE TABLE training_type (
    id_training_type SERIAL PRIMARY KEY NOT NULL,
    training_type_name VARCHAR(45) NOT NULL,
    workout_description VARCHAR(300)
);

-- 18. ТАБЛИЦА ТРЕНЕРОВ (ОБНОВЛЕНА)
CREATE TABLE trainers (
    id_trainer SERIAL PRIMARY KEY NOT NULL,
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    speciality VARCHAR(45),
    experience INTEGER,
    certifications INTEGER,
    hire_date DATE NOT NULL
);

-- 19. ТАБЛИЦА РАСПИСАНИЯ ТРЕНИРОВОК
CREATE TABLE training_schedule (
    id_session SERIAL PRIMARY KEY NOT NULL,
    id_trainer INTEGER NOT NULL,
    id_training_type INTEGER,
    session_date TIMESTAMP NOT NULL,
    session_time INTEGER NOT NULL,
    FOREIGN KEY (id_trainer) REFERENCES trainers (id_trainer) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_training_type) REFERENCES training_type (id_training_type) ON DELETE SET NULL ON UPDATE CASCADE
);

-- 20. ТАБЛИЦА СВЯЗИ ЧЛЕНОВ И РАСПИСАНИЯ
CREATE TABLE members_have_training_schedule (
    id_member INTEGER NOT NULL,
    id_session INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_session),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_session) REFERENCES training_schedule (id_session) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 24. ТАБЛИЦА НОВОСТЕЙ
CREATE TABLE news (
    id_news SERIAL PRIMARY KEY NOT NULL,
    news_title VARCHAR(45) NOT NULL,
    news_text VARCHAR(200)
);

-- 25. ТАБЛИЦА СВЯЗИ КЛУБОВ И НОВОСТЕЙ
CREATE TABLE clubs_have_news (
    club_name VARCHAR(45) NOT NULL,
    id_news INTEGER NOT NULL,
    PRIMARY KEY (club_name, id_news),
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_news) REFERENCES news (id_news) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 26. ТАБЛИЦА ПЕРСОНАЛА (ОБНОВЛЕНА)
CREATE TABLE staff (
    id_staff SERIAL PRIMARY KEY NOT NULL,
    id_position INTEGER,
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    hire_date DATE NOT NULL,
    staff_about VARCHAR(100),
    FOREIGN KEY (id_position) REFERENCES position(id_position) ON DELETE SET NULL ON UPDATE CASCADE
);

-- 27. ТАБЛИЦА РАСПИСАНИЯ ПЕРСОНАЛА
CREATE TABLE staff_schedule (
    id_schedule SERIAL PRIMARY KEY NOT NULL,
    id_staff INTEGER,
    club_name VARCHAR(45) NOT NULL,
    date DATE NOT NULL,
    shift INTEGER NOT NULL,
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_staff) REFERENCES staff (id_staff) ON DELETE SET NULL ON UPDATE CASCADE
);

-- 28. ТАБЛИЦА ФОТО ПОЛЬЗОВАТЕЛЕЙ
CREATE TABLE users_photo (
    id_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(200) NOT NULL
);

-- 29. ТАБЛИЦА АККАУНТОВ ЧЛЕНОВ
CREATE TABLE members_accounts (
    username VARCHAR(45) PRIMARY KEY NOT NULL,
    id_member INTEGER NOT NULL,
    id_photo INTEGER DEFAULT 1,
    PASSWORD VARCHAR(100) NOT NULL,
    account_creation_date DATE NOT NULL,
    last_login DATE,
    user_role VARCHAR(45) NOT NULL,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_photo) REFERENCES users_photo (id_photo) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

-- 30. ТАБЛИЦА ОТЗЫВОВ
CREATE TABLE feedback (
    id_feedback SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(45) NOT NULL,
    feedback_text VARCHAR(45),
    feedback_date DATE NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    FOREIGN KEY (username) REFERENCES members_accounts (username) ON DELETE CASCADE ON UPDATE CASCADE
);

-- 31. ТАБЛИЦА ФОТО ТРЕНЕРОВ
CREATE TABLE trainers_photo (
    id_trainers_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(250) NOT NULL
);

-- 32. ТАБЛИЦА АККАУНТОВ ТРЕНЕРОВ
CREATE TABLE trainers_accounts (
    username VARCHAR(45) PRIMARY KEY NOT NULL,
    id_trainer INTEGER NOT NULL,
    id_trainers_photo INTEGER DEFAULT 1,
    PASSWORD VARCHAR(100) NOT NULL,
    last_login DATE,
    account_creation_date DATE NOT NULL,
    user_role VARCHAR(45) NOT NULL,
    FOREIGN KEY (id_trainer) REFERENCES trainers (id_trainer) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_trainers_photo) REFERENCES trainers_photo (id_trainers_photo) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

-- 33. ТАБЛИЦА ФОТО ПЕРСОНАЛА
CREATE TABLE staff_photo (
    id_staff_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(250) NOT NULL
);

-- 34. ТАБЛИЦА АККАУНТОВ ПЕРСОНАЛА
CREATE TABLE staff_accounts (
    id_staff INTEGER NOT NULL,
    username VARCHAR(45) PRIMARY KEY NOT NULL,
    id_staff_photo INTEGER DEFAULT 1,
    PASSWORD VARCHAR(100) NOT NULL,
    last_login DATE,
    account_creation_date DATE NOT NULL,
    user_role VARCHAR(45) NOT NULL,
    FOREIGN KEY (id_staff) REFERENCES staff (id_staff) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_staff_photo) REFERENCES staff_photo (id_staff_photo) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

-- 35. ТАБЛИЦА УПРАЖНЕНИЙ
CREATE TABLE exercises (
    id_exercise SERIAL PRIMARY KEY,
    exercise_name VARCHAR(100) NOT NULL,
    description VARCHAR(300),
    muscle_group VARCHAR(100),
    difficulty_level INTEGER,
    equipment_required VARCHAR(100),
    estimated_calories INTEGER
);

-- 36. ТАБЛИЦА ПРОГРАММ ТРЕНИРОВОК
CREATE TABLE training_programs (
    id_program SERIAL PRIMARY KEY,
    id_member INTEGER NOT NULL,
    id_nutrition_plan INTEGER,
    program_name VARCHAR(100) NOT NULL,
    goal VARCHAR(100),
    LEVEL VARCHAR(20),
    duration_weeks INTEGER,
    created_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE,
    FOREIGN KEY (id_nutrition_plan) REFERENCES nutrition_plan (id_plan) ON DELETE SET NULL
);

-- 37. ТАБЛИЦА ДНЕЙ ПРОГРАММЫ
CREATE TABLE program_days (
    id_day SERIAL PRIMARY KEY,
    id_program INTEGER NOT NULL,
    day_number INTEGER,
    day_name VARCHAR(20),
    muscle_groups VARCHAR(100),
    FOREIGN KEY (id_program) REFERENCES training_programs (id_program) ON DELETE CASCADE
);

-- 38. ТАБЛИЦА УПРАЖНЕНИЙ В ДНЯХ ПРОГРАММЫ
CREATE TABLE program_exercises (
    id_program_exercise SERIAL PRIMARY KEY,
    id_day INTEGER NOT NULL,
    id_exercise INTEGER NOT NULL,
    SETS INTEGER,
    reps INTEGER,
    weight DOUBLE PRECISION,
    rest_seconds INTEGER,
    order_index INTEGER,
    FOREIGN KEY (id_day) REFERENCES program_days (id_day) ON DELETE CASCADE,
    FOREIGN KEY (id_exercise) REFERENCES exercises (id_exercise) ON DELETE CASCADE
);

-- 39. ТАБЛИЦА ТРЕБОВАНИЙ К ОБОРУДОВАНИЮ ДЛЯ УПРАЖНЕНИЙ
CREATE TABLE exercise_equipment_requirements (
    id_requirement SERIAL PRIMARY KEY,
    id_exercise INTEGER NOT NULL,
    id_equipment_type INTEGER NOT NULL,
    quantity_required INTEGER DEFAULT 1,
    is_required BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (id_exercise) REFERENCES exercises (id_exercise) ON DELETE CASCADE,
    FOREIGN KEY (id_equipment_type) REFERENCES equipment_type (id_equipment_type) ON DELETE CASCADE
);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA fitness_club_db TO springroot;

ALTER DEFAULT PRIVILEGES IN SCHEMA fitness_club_db
GRANT ALL PRIVILEGES ON TABLES TO springroot;

CREATE INDEX idx_members_club_name ON members (club_name);

CREATE INDEX idx_training_schedule_date ON training_schedule (session_date);

CREATE INDEX idx_visits_history_date ON visits_history (visit_date);

CREATE INDEX idx_training_schedule_trainer ON training_schedule (id_trainer);

CREATE INDEX idx_feedback_rating ON feedback (rating);

COMMENT ON TABLE members IS 'Таблица членов фитнес-клуба';

COMMENT ON COLUMN members.gender IS '0-женский, 1-мужской, 2-другой';