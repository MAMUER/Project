DROP SCHEMA IF EXISTS fitness_club_db CASCADE;

CREATE SCHEMA IF NOT EXISTS fitness_club_db;

SET search_path TO fitness_club_db;

CREATE TABLE achievements (
    id_achievement SERIAL PRIMARY KEY NOT NULL,
    achievement_description VARCHAR(45),
    achievement_title VARCHAR(45) NOT NULL,
    achievement_icon_url VARCHAR(100) NOT NULL
);

CREATE TABLE activity_type (
    id_activity SERIAL PRIMARY KEY NOT NULL,
    activity_name VARCHAR(45) NOT NULL
);

CREATE TABLE clubs (
    club_name VARCHAR(45) PRIMARY KEY NOT NULL,
    address VARCHAR(45) NOT NULL
);

CREATE TABLE news (
    id_news SERIAL PRIMARY KEY NOT NULL,
    news_title VARCHAR(45) NOT NULL,
    news_text VARCHAR(200)
);

CREATE TABLE clubs_have_news (
    club_name VARCHAR(45) NOT NULL,
    id_news INTEGER NOT NULL,
    PRIMARY KEY (club_name, id_news),
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_news) REFERENCES news (id_news) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE equipment_type (
    id_equipment_type SERIAL PRIMARY KEY NOT NULL,
    type_name VARCHAR(45) NOT NULL
);

CREATE TABLE gyms (
    id_gym SERIAL PRIMARY KEY NOT NULL,
    club_name VARCHAR(45) NOT NULL,
    gym_name VARCHAR(45) NOT NULL,
    capacity INTEGER NOT NULL,
    available_hours INTEGER NOT NULL,
    CONSTRAINT fk_gyms_club FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE equipment (
    id_equipment SERIAL PRIMARY KEY NOT NULL,
    id_equipment_type INTEGER,
    name VARCHAR(45) NOT NULL,
    quantity INTEGER,
    id_gym INTEGER,
    FOREIGN KEY (id_equipment_type) REFERENCES equipment_type (id_equipment_type) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_equipment_gym FOREIGN KEY (id_gym) REFERENCES gyms (id_gym) ON DELETE SET NULL
);

CREATE TABLE equipment_statistics (
    id_statistics SERIAL PRIMARY KEY NOT NULL,
    id_activity INTEGER NOT NULL,
    approaches INTEGER,
    kilocalories INTEGER,
    FOREIGN KEY (id_activity) REFERENCES activity_type (id_activity) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE membership_role (
    id_role SERIAL PRIMARY KEY NOT NULL,
    role_name VARCHAR(45) NOT NULL
);

CREATE TABLE members (
    id_member SERIAL PRIMARY KEY NOT NULL,
    id_role INTEGER NOT NULL,
    club_name VARCHAR(45),
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    phone_number VARCHAR(11) NOT NULL,
    email VARCHAR(45) NOT NULL,
    birth_date DATE NOT NULL,
    start_trial_date DATE NOT NULL,
    end_trial_date DATE,
    gender INTEGER NOT NULL,
    FOREIGN KEY (id_role) REFERENCES membership_role (id_role) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE users_photo (
    id_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(200) NOT NULL
);

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

CREATE TABLE feedback (
    id_feedback SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(45) NOT NULL,
    feedback_text VARCHAR(45),
    feedback_date DATE NOT NULL,
    rating SMALLINT NOT NULL,
    FOREIGN KEY (username) REFERENCES members_accounts (username) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE gyms_have_equipment (
    id_gym INTEGER NOT NULL,
    id_equipment INTEGER NOT NULL,
    PRIMARY KEY (id_gym, id_equipment),
    FOREIGN KEY (id_equipment) REFERENCES equipment (id_equipment) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_gym) REFERENCES gyms (id_gym) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE inbody_analyses (
    id_inbody_analys SERIAL PRIMARY KEY NOT NULL,
    HEIGHT FLOAT,
    weight FLOAT,
    bmi FLOAT,
    fat_percent FLOAT,
    muscle_persent FLOAT
);

CREATE TABLE members_have_achievements (
    id_member INTEGER NOT NULL,
    id_achievement INTEGER NOT NULL,
    receipt_date DATE NOT NULL,
    PRIMARY KEY (id_member, id_achievement),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_achievement) REFERENCES achievements (id_achievement) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE members_have_equipment_statistics (
    id_member INTEGER NOT NULL,
    id_statistics INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_statistics),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_statistics) REFERENCES equipment_statistics (id_statistics) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE members_have_inbody_analyses (
    id_member INTEGER NOT NULL,
    id_inbody_analys INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_inbody_analys),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_inbody_analys) REFERENCES inbody_analyses (id_inbody_analys) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE training_type (
    id_training_type SERIAL PRIMARY KEY NOT NULL,
    training_type_name VARCHAR(45) NOT NULL,
    workout_description VARCHAR(300)
);

CREATE TABLE trainers (
    id_trainer SERIAL PRIMARY KEY NOT NULL,
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    speciality VARCHAR(45),
    experience INTEGER,
    certifications INTEGER,
    phone_number VARCHAR(11) NOT NULL,
    email VARCHAR(45) NOT NULL,
    hire_date DATE NOT NULL
);

CREATE TABLE trainers_photo (
    id_trainers_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(250) NOT NULL
);

CREATE TABLE training_schedule (
    id_session SERIAL PRIMARY KEY NOT NULL,
    id_trainer INTEGER NOT NULL,
    id_training_type INTEGER,
    session_date TIMESTAMP NOT NULL,
    session_time INTEGER NOT NULL,
    FOREIGN KEY (id_trainer) REFERENCES trainers (id_trainer) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_training_type) REFERENCES training_type (id_training_type) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE members_have_training_schedule (
    id_member INTEGER NOT NULL,
    id_session INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_session),
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_session) REFERENCES training_schedule (id_session) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE visits_history (
    id_visit SERIAL PRIMARY KEY NOT NULL,
    visit_date DATE
);

CREATE TABLE members_have_visits_history (
    id_member INTEGER NOT NULL,
    id_visit INTEGER NOT NULL,
    PRIMARY KEY (id_member, id_visit),
    FOREIGN KEY (id_visit) REFERENCES visits_history (id_visit) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE nutrition_plan (
    id_plan SERIAL PRIMARY KEY NOT NULL,
    id_member INTEGER NOT NULL,
    nutrition_description VARCHAR(100),
    start_date DATE NOT NULL,
    end_date DATE,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE position(
    id_position SERIAL PRIMARY KEY NOT NULL,
    role_name VARCHAR(45) NOT NULL
);

CREATE TABLE staff (
    id_staff SERIAL PRIMARY KEY NOT NULL,
    id_position INTEGER,
    first_name VARCHAR(45) NOT NULL,
    second_name VARCHAR(45) NOT NULL,
    phone_number VARCHAR(11) NOT NULL,
    email VARCHAR(45) NOT NULL,
    hire_date DATE NOT NULL,
    staff_about VARCHAR(100),
    gender INTEGER NOT NULL,
    FOREIGN KEY (id_position) REFERENCES position(id_position) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE staff_photo (
    id_staff_photo SERIAL PRIMARY KEY NOT NULL,
    image_url VARCHAR(250) NOT NULL
);

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

CREATE TABLE staff_schedule (
    id_schedule SERIAL PRIMARY KEY NOT NULL,
    id_staff INTEGER,
    club_name VARCHAR(45) NOT NULL,
    date DATE NOT NULL,
    shift INTEGER NOT NULL,
    FOREIGN KEY (club_name) REFERENCES clubs (club_name) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (id_staff) REFERENCES staff (id_staff) ON DELETE SET NULL ON UPDATE CASCADE
);

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

-- Таблица упражнений
CREATE TABLE exercises (
    id_exercise SERIAL PRIMARY KEY,
    exercise_name VARCHAR(100) NOT NULL,
    description VARCHAR(300),
    muscle_group VARCHAR(50),
    difficulty_level INTEGER,
    equipment_required VARCHAR(100),
    estimated_calories INTEGER
);

-- Таблица программ тренировок
CREATE TABLE training_programs (
    id_program SERIAL PRIMARY KEY,
    id_member INTEGER NOT NULL,
    program_name VARCHAR(100) NOT NULL,
    goal VARCHAR(100),
    LEVEL VARCHAR(20),
    duration_weeks INTEGER,
    created_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (id_member) REFERENCES members (id_member) ON DELETE CASCADE
);

-- Таблица дней программы
CREATE TABLE program_days (
    id_day SERIAL PRIMARY KEY,
    id_program INTEGER NOT NULL,
    day_number INTEGER,
    day_name VARCHAR(20),
    muscle_groups VARCHAR(100),
    FOREIGN KEY (id_program) REFERENCES training_programs (id_program) ON DELETE CASCADE
);

-- Таблица упражнений в днях программы
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

-- Таблица связей упражнение-оборудование
CREATE TABLE exercise_equipment_requirements (
    id_requirement SERIAL PRIMARY KEY,
    id_exercise INTEGER NOT NULL,
    id_equipment_type INTEGER NOT NULL,
    quantity_required INTEGER DEFAULT 1,
    is_required BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (id_exercise) REFERENCES exercises (id_exercise) ON DELETE CASCADE,
    FOREIGN KEY (id_equipment_type) REFERENCES equipment_type (id_equipment_type) ON DELETE CASCADE
);