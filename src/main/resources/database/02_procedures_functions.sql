-- Create necessary tables first
CREATE TABLE IF NOT EXISTS equipment_supplies (
    id_supply INT,
    date_supply TIMESTAMP
);

CREATE TABLE IF NOT EXISTS news_audit (
    id INT,
    date_changed TIMESTAMP
);

CREATE TABLE IF NOT EXISTS training_schedule_audit (
    id_session INT,
    changed_at DATE
);

-- Trigger functions
CREATE OR REPLACE FUNCTION achievements_after_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO members_have_achievements (receipt_date)
    VALUES (NOW());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION equipment_after_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO equipment_supplies (id_supply, date_supply)
    VALUES (NEW.id_equipment, NOW());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION members_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.end_trial_date < NEW.start_trial_date THEN 
        RAISE EXCEPTION 'Cannot insert a member with an end trial date earlier start date.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION members_accounts_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM members_accounts WHERE username = NEW.username) THEN 
        RAISE EXCEPTION 'Cannot insert a new member with an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION members_accounts_before_update()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM members_accounts WHERE username = NEW.username AND username != OLD.username) THEN 
        RAISE EXCEPTION 'Cannot change to an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION news_after_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO news_audit (id, date_changed)
    VALUES (NEW.id_news, NOW());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION nutrition_plan_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.end_date < NEW.start_date THEN 
        RAISE EXCEPTION 'Cannot insert a plan with a date in the past.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staff_accounts_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM staff_accounts WHERE username = NEW.username) THEN 
        RAISE EXCEPTION 'Cannot make a new account with an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staff_accounts_before_update()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM staff_accounts WHERE username = NEW.username AND username != OLD.username) THEN 
        RAISE EXCEPTION 'Cannot change to an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trainers_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM training_schedule WHERE id_trainer = OLD.id_trainer;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trainers_accounts_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM trainers_accounts WHERE username = NEW.username) THEN 
        RAISE EXCEPTION 'Cannot make a new account with an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trainers_accounts_before_update()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM trainers_accounts WHERE username = NEW.username AND username != OLD.username) THEN 
        RAISE EXCEPTION 'Cannot change to an existing username.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION training_schedule_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.session_date < CURRENT_DATE THEN 
        RAISE EXCEPTION 'Cannot insert a session with a date in the past.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION training_schedule_after_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO training_schedule_audit (id_session, changed_at)
    VALUES (NEW.id_session, CURRENT_DATE);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION visits_history_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.visit_date < CURRENT_DATE THEN 
        RAISE EXCEPTION 'Cannot insert a visit with a date in the past.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Delete cascade triggers
CREATE OR REPLACE FUNCTION training_schedule_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM training_schedule 
    WHERE training_schedule.id_training_type = 5 
    AND training_schedule.id_session = OLD.id_session;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION visits_history_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM visits_history WHERE visits_history.id_visit = OLD.id_visit;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION achievements_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM achievements WHERE achievements.id_achievement = OLD.id_achievement;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION inbody_analyses_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM inbody_analyis WHERE inbody_analyses.id_inbody_analys = OLD.id_inbody_analys;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION equipment_statistics_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM equipment_statistics WHERE equipment_statistics.id_statistics = OLD.id_statistics;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION users_photo_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM users_photo WHERE users_photo.id_photo = OLD.id_photo;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trainers_photo_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM trainers_photo WHERE trainers_photo.id_trainers_photo = OLD.id_trainers_photo;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staff_photo_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM staff_photo WHERE staff_photo.id_staff_photo = OLD.id_staff_photo;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
CREATE OR REPLACE TRIGGER achievements_after_update
    AFTER UPDATE ON achievements
    FOR EACH ROW EXECUTE FUNCTION achievements_after_update();

CREATE OR REPLACE TRIGGER equipment_after_update
    AFTER UPDATE ON equipment
    FOR EACH ROW EXECUTE FUNCTION equipment_after_update();

CREATE OR REPLACE TRIGGER members_before_insert
    BEFORE INSERT ON members
    FOR EACH ROW EXECUTE FUNCTION members_before_insert();

CREATE OR REPLACE TRIGGER members_accounts_before_insert
    BEFORE INSERT ON members_accounts
    FOR EACH ROW EXECUTE FUNCTION members_accounts_before_insert();

CREATE OR REPLACE TRIGGER members_accounts_before_update
    BEFORE UPDATE ON members_accounts
    FOR EACH ROW EXECUTE FUNCTION members_accounts_before_update();

CREATE OR REPLACE TRIGGER news_after_update
    AFTER UPDATE ON news
    FOR EACH ROW EXECUTE FUNCTION news_after_update();

CREATE OR REPLACE TRIGGER nutrition_plan_before_insert
    BEFORE INSERT ON nutrition_plan
    FOR EACH ROW EXECUTE FUNCTION nutrition_plan_before_insert();

CREATE OR REPLACE TRIGGER staff_accounts_before_insert
    BEFORE INSERT ON staff_accounts
    FOR EACH ROW EXECUTE FUNCTION staff_accounts_before_insert();

CREATE OR REPLACE TRIGGER staff_accounts_before_update
    BEFORE UPDATE ON staff_accounts
    FOR EACH ROW EXECUTE FUNCTION staff_accounts_before_update();

CREATE OR REPLACE TRIGGER trainers_after_delete
    AFTER DELETE ON trainers
    FOR EACH ROW EXECUTE FUNCTION trainers_after_delete();

CREATE OR REPLACE TRIGGER trainers_accounts_before_insert
    BEFORE INSERT ON trainers_accounts
    FOR EACH ROW EXECUTE FUNCTION trainers_accounts_before_insert();

CREATE OR REPLACE TRIGGER trainers_accounts_before_update
    BEFORE UPDATE ON trainers_accounts
    FOR EACH ROW EXECUTE FUNCTION trainers_accounts_before_update();

CREATE OR REPLACE TRIGGER training_schedule_before_insert
    BEFORE INSERT ON training_schedule
    FOR EACH ROW EXECUTE FUNCTION training_schedule_before_insert();

CREATE OR REPLACE TRIGGER training_schedule_after_update
    AFTER UPDATE ON training_schedule
    FOR EACH ROW EXECUTE FUNCTION training_schedule_after_update();

CREATE OR REPLACE TRIGGER visits_history_before_insert
    BEFORE INSERT ON visits_history
    FOR EACH ROW EXECUTE FUNCTION visits_history_before_insert();

CREATE OR REPLACE TRIGGER training_schedule_after_delete
    AFTER DELETE ON members_have_training_schedule
    FOR EACH ROW EXECUTE FUNCTION training_schedule_after_delete();

CREATE OR REPLACE TRIGGER visits_history_after_delete
    AFTER DELETE ON members_have_visits_history
    FOR EACH ROW EXECUTE FUNCTION visits_history_after_delete();

CREATE OR REPLACE TRIGGER achievements_after_delete
    AFTER DELETE ON members_have_achievements
    FOR EACH ROW EXECUTE FUNCTION achievements_after_delete();

CREATE OR REPLACE TRIGGER users_photo_after_delete
    AFTER DELETE ON members_accounts
    FOR EACH ROW EXECUTE FUNCTION users_photo_after_delete();

CREATE OR REPLACE TRIGGER trainers_photo_after_delete
    AFTER DELETE ON trainers_accounts
    FOR EACH ROW EXECUTE FUNCTION trainers_photo_after_delete();

CREATE OR REPLACE TRIGGER staff_photo_after_delete
    AFTER DELETE ON staff_accounts
    FOR EACH ROW EXECUTE FUNCTION staff_photo_after_delete();

CREATE OR REPLACE PROCEDURE members_have_inbody_analysis_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM members_have_inbody_analysis WHERE id_inbody_analys = 1;
END;
$$;

CREATE OR REPLACE TRIGGER inbody_analyses_after_delete
    AFTER DELETE ON members_have_inbody_analysis
    FOR EACH ROW EXECUTE FUNCTION inbody_analyses_after_delete();

CREATE OR REPLACE PROCEDURE members_have_achievements_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM members_have_achievements WHERE id_achievement = 1;
END;
$$;

CREATE OR REPLACE PROCEDURE members_have_visits_history_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM members_have_visits_history WHERE id_visit = 1;
END;
$$;

CREATE OR REPLACE PROCEDURE members_have_training_schedule_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM members_have_training_schedule WHERE id_session = 1;
END;
$$;

CREATE OR REPLACE PROCEDURE visits_history_add()
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO visits_history(visit_date) VALUES ('2024-09-11');
END;
$$;

CREATE OR REPLACE PROCEDURE staff_add()
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO staff (id_position, first_name, second_name, hire_date, staff_about) 
    VALUES (3, 'Сергей', 'Михайлов', '2020-10-2', 'Какой-то мужик', 1);
END;
$$;

CREATE OR REPLACE PROCEDURE training_type_add()
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO training_type(training_type_name, workout_description) 
    VALUES ('Прыжки', 'Групповые прыжки вверх с двух ног. Отличный способ бесполезно провести время. Развлекайтесь!');
END;
$$;

CREATE OR REPLACE PROCEDURE visits_history_6_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM visits_history WHERE id_visit = 6;
END;
$$;

CREATE OR REPLACE PROCEDURE staff_5_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM staff WHERE id_staff = 5;
END;
$$;

CREATE OR REPLACE PROCEDURE training_type_delete()
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM training_type WHERE id_training_type = 5;
END;
$$;

-- Reporting procedures
CREATE OR REPLACE PROCEDURE members_feedback()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT first_name, second_name, username, feedback_text, rating, feedback_date
    FROM members
    JOIN members_accounts USING (id_member)
    JOIN feedback USING (username)
    ORDER BY feedback_date ASC;
END;
$$;

CREATE OR REPLACE PROCEDURE members_inbody_analyses()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT first_name, second_name, height, weight, bmi, fat_percent, muscle_persent
    FROM members
    JOIN members_have_inbody_analysis USING (id_member)
    JOIN inbody_analysis USING (id_inbody_analys);
END;
$$;

CREATE OR REPLACE PROCEDURE members_nutrition_plan()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT m.first_name, m.second_name, np.nutrition_description, np.created_date, np.updated_date
    FROM members m
    LEFT JOIN nutrition_plan np ON m.id_member = np.id_member
    ORDER BY m.first_name, m.second_name;
END;
$$;

CREATE OR REPLACE PROCEDURE members_trainings()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT username, first_name, second_name, session_date, session_time
    FROM members_accounts 
    JOIN members USING (id_member)
    JOIN members_have_training_schedule USING (id_member)
    JOIN training_schedule USING (id_session)
    JOIN training_type USING (id_training_type)
    ORDER BY session_date ASC;
END;
$$;

CREATE OR REPLACE PROCEDURE members_visits()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT first_name, second_name, visit_date
    FROM members
    JOIN members_have_visits_history USING(id_member)
    JOIN visits_history USING(id_visit)
    ORDER BY visit_date ASC;
END;
$$;

CREATE OR REPLACE PROCEDURE staff_info()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT first_name, second_name, username, role_name
    FROM staff_accounts
    JOIN staff USING(id_staff)
    JOIN position USING(id_position);
END;
$$;

CREATE OR REPLACE PROCEDURE staff_schedule()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT first_name, second_name, staff_schedule.club_name, shift, weekday
    FROM staff, staff_schedule, clubs
    WHERE staff.id_staff = staff_schedule.id_staff
    AND staff_schedule.club_name = clubs.club_name;
END;
$$;

CREATE OR REPLACE PROCEDURE trainers_trainings()
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT username, first_name, second_name, session_date, session_time
    FROM trainers_accounts 
    JOIN trainers USING (id_trainer)
    JOIN training_schedule USING (id_trainer)
    JOIN training_type USING (id_training_type)
    ORDER BY session_date ASC;
END;
$$;

-- Functions
CREATE OR REPLACE FUNCTION TotalMembersWithNutritionPlan()
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM nutrition_plan;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalTrainersWithTrainingType(id_training_type INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(DISTINCT id_trainer) INTO total FROM training_schedule WHERE id_training_type = id_training_type;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalStaffInPosition(id_position INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM staff WHERE id_position = id_position;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalNewsForClub(club_name VARCHAR)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM clubs_have_news WHERE club_name = club_name;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION AverageRatingFeedbackForMember(username VARCHAR)
RETURNS FLOAT
LANGUAGE plpgsql
AS $$
DECLARE
    avg_rating FLOAT;
BEGIN
    SELECT AVG(rating) INTO avg_rating FROM feedback WHERE username = username;
    RETURN avg_rating;
END;
$$;

CREATE OR REPLACE FUNCTION members_amount_1()
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE 
    a INT;
BEGIN
    SELECT COUNT(members.first_name) INTO a 
    FROM members, members_have_training_schedule, training_schedule
    WHERE training_schedule.id_session = 1
    AND members_have_training_schedule.id_member = members.id_member
    AND training_schedule.id_session = members_have_training_schedule.id_session;
    RETURN a;
END;
$$;

CREATE OR REPLACE FUNCTION TotalMemberOnTrainingDate(dateon DATE)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM training_schedule WHERE session_date = dateon;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalVisitsOnDate(dateon DATE)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM visits_history WHERE visit_date = dateon;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalEquipmentOfType(id INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM equipment WHERE id_equipment_type = id;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalSessionsForTrainer(id INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM training_schedule WHERE id_trainer = id;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalGymsInClub(nameon VARCHAR)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM gyms WHERE club_name = nameon;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalAchievementsForMember(id INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM members_have_achievements WHERE id_member = id;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalMembersInClub(nameon VARCHAR)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM members WHERE club_name = nameon;
    RETURN total;
END;
$$;

CREATE OR REPLACE FUNCTION TotalMembersWithAchievement(id INT)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    total INT;
BEGIN
    SELECT COUNT(*) INTO total FROM members_have_achievements WHERE id_achievement = id;
    RETURN total;
END;
$$;

-- Функция для вставки или обновления плана питания
CREATE OR REPLACE FUNCTION upsert_nutrition_plan(
    p_member_id INTEGER,
    p_nutrition_description VARCHAR(200)
)
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    existing_plan_id INTEGER;
BEGIN
    -- Проверяем, существует ли уже план для этого пользователя
    SELECT id_plan INTO existing_plan_id 
    FROM nutrition_plan 
    WHERE id_member = p_member_id;
    
    IF existing_plan_id IS NOT NULL THEN
        -- Обновляем существующий план
        UPDATE nutrition_plan 
        SET nutrition_description = p_nutrition_description,
            updated_date = CURRENT_DATE
        WHERE id_member = p_member_id;
        RETURN existing_plan_id;
    ELSE
        -- Вставляем новый план
        INSERT INTO nutrition_plan (id_member, nutrition_description)
        VALUES (p_member_id, p_nutrition_description)
        RETURNING id_plan INTO existing_plan_id;
        RETURN existing_plan_id;
    END IF;
END;
$$;

-- Новый триггер для проверки перед вставкой (если не используем UPSERT функцию)
CREATE OR REPLACE FUNCTION nutrition_plan_before_insert()
RETURNS TRIGGER AS $$
BEGIN
    -- Проверяем, нет ли уже плана у этого пользователя
    IF EXISTS (SELECT 1 FROM nutrition_plan WHERE id_member = NEW.id_member) THEN
        RAISE EXCEPTION 'Member already has a nutrition plan. Use UPDATE instead.';
    END IF;
    
    -- Устанавливаем даты
    NEW.created_date = COALESCE(NEW.created_date, CURRENT_DATE);
    NEW.updated_date = CURRENT_DATE;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER nutrition_plan_before_insert
    BEFORE INSERT ON nutrition_plan
    FOR EACH ROW EXECUTE FUNCTION nutrition_plan_before_insert();

-- Триггер для обновления updated_date
CREATE OR REPLACE FUNCTION nutrition_plan_before_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_date = CURRENT_DATE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER nutrition_plan_before_update
    BEFORE UPDATE ON nutrition_plan
    FOR EACH ROW EXECUTE FUNCTION nutrition_plan_before_update();