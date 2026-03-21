import asyncio
import numpy as np
import tensorflow as tf
from fastapi import FastAPI, HTTPException
import uvicorn
import logging
from pydantic import BaseModel
from typing import List, Optional
import json

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="HealthFit Generation Service (GAN)", version="1.0.0")

class TrainingProgramGenerator:
    def __init__(self):
        self.generator = self._build_generator()
        self.discriminator = self._build_discriminator()
        self.gan = self._build_gan()
        self.load_weights()
    
    def _build_generator(self):
        # Генератор: вход - латентный вектор (100) + условия (противопоказания, цели)
        model = tf.keras.Sequential([
            tf.keras.layers.Dense(128, activation='relu', input_shape=(100 + 20,)),  # 20 параметров условий
            tf.keras.layers.BatchNormalization(),
            tf.keras.layers.Dense(256, activation='relu'),
            tf.keras.layers.BatchNormalization(),
            tf.keras.layers.Dense(512, activation='relu'),
            tf.keras.layers.BatchNormalization(),
            tf.keras.layers.Dense(7 * 30, activation='tanh')  # 7 дней * 30 недель = 210 параметров программы
        ])
        return model
    
    def _build_discriminator(self):
        model = tf.keras.Sequential([
            tf.keras.layers.Dense(512, activation='relu', input_shape=(210,)),
            tf.keras.layers.Dropout(0.3),
            tf.keras.layers.Dense(256, activation='relu'),
            tf.keras.layers.Dropout(0.3),
            tf.keras.layers.Dense(1, activation='sigmoid')
        ])
        return model
    
    def _build_gan(self):
        self.discriminator.trainable = False
        model = tf.keras.Sequential([self.generator, self.discriminator])
        model.compile(optimizer='adam', loss='binary_crossentropy')
        return model
    
    def load_weights(self):
        try:
            self.generator.load_weights('models/generator_weights.h5')
            logger.info("Generator weights loaded")
        except:
            logger.warning("No pre-trained weights found, using untrained model")
    
    def generate_program(self, training_class, user_conditions, duration_weeks=52):
        # Создаем латентный вектор
        latent_dim = 100
        noise = np.random.normal(0, 1, (1, latent_dim))
        
        # Кодируем условия
        conditions_vector = self._encode_conditions(training_class, user_conditions)
        input_vector = np.concatenate([noise, conditions_vector], axis=1)
        
        # Генерируем программу
        generated = self.generator.predict(input_vector, verbose=0)
        
        # Декодируем в читаемый формат
        program = self._decode_program(generated[0], duration_weeks)
        
        return program
    
    def _encode_conditions(self, training_class, user_conditions):
        # 20-мерный вектор условий
        vec = np.zeros(20)
        
        # Кодируем класс тренировки (one-hot)
        classes = ['cardio', 'strength', 'flexibility', 'recovery', 'hiit', 'endurance']
        if training_class in classes:
            vec[classes.index(training_class)] = 1
        
        # Кодируем противопоказания
        contraindications = ['heart_issues', 'joint_problems', 'hypertension', 'asthma', 'diabetes']
        for i, cond in enumerate(contraindications):
            if cond in user_conditions.get('contraindications', []):
                vec[6 + i] = 1
        
        # Кодируем цели
        goals = ['weight_loss', 'muscle_gain', 'endurance', 'flexibility', 'rehabilitation']
        for i, goal in enumerate(goals):
            if goal in user_conditions.get('goals', []):
                vec[11 + i] = 1
        
        # Кодируем уровень подготовки
        fitness_level = user_conditions.get('fitness_level', 'beginner')
        levels = {'beginner': 0, 'intermediate': 0.5, 'advanced': 1}
        vec[16] = levels.get(fitness_level, 0)
        
        # Кодируем возрастную группу
        age_group = user_conditions.get('age_group', 'adult')
        groups = {'young': 0, 'adult': 0.33, 'senior': 0.66, 'elderly': 1}
        vec[17] = groups.get(age_group, 0.33)
        
        # Кодируем пол (0=жен, 1=муж)
        vec[18] = 1 if user_conditions.get('gender') == 'male' else 0
        
        # Кодируем наличие травм
        vec[19] = 1 if user_conditions.get('has_injury', False) else 0
        
        return vec.reshape(1, -1)
    
    def _decode_program(self, raw_program, duration_weeks):
        # raw_program: 210 параметров (7 дней * 30 недель)
        # Декодируем в структурированную программу
        program = []
        weeks = min(duration_weeks, 30)
        
        for week in range(weeks):
            week_data = []
            for day in range(7):
                idx = week * 7 + day
                value = float(raw_program[idx])
                # Преобразуем в тип тренировки и интенсивность
                if value < -0.7:
                    workout_type = "rest"
                    intensity = "none"
                elif value < -0.3:
                    workout_type = "recovery"
                    intensity = "light"
                elif value < 0.1:
                    workout_type = "cardio"
                    intensity = "moderate"
                elif value < 0.5:
                    workout_type = "strength"
                    intensity = "moderate"
                elif value < 0.9:
                    workout_type = "hiit"
                    intensity = "high"
                else:
                    workout_type = "endurance"
                    intensity = "high"
                
                week_data.append({
                    "day": day + 1,
                    "workout_type": workout_type,
                    "intensity": intensity,
                    "duration_minutes": int(30 + abs(value) * 60)
                })
            
            program.append({
                "week": week + 1,
                "schedule": week_data,
                "notes": self._generate_week_notes(week_data)
            })
        
        return program
    
    def _generate_week_notes(self, week_data):
        # Генерация рекомендаций на неделю
        has_hiit = any(d['workout_type'] == 'hiit' for d in week_data)
        has_rest = any(d['workout_type'] == 'rest' for d in week_data)
        
        notes = []
        if has_hiit:
            notes.append("Включите в HIIT-тренировки интервалы 30/90 сек")
        if has_rest:
            notes.append("Дни отдыха важны для восстановления")
        notes.append("Пейте воду до, во время и после тренировок")
        
        return notes

generator = TrainingProgramGenerator()

# Pydantic модели
class UserConditions(BaseModel):
    contraindications: List[str] = []
    goals: List[str] = []
    fitness_level: str = "beginner"
    age_group: str = "adult"
    gender: str = "female"
    has_injury: bool = False

class GenerateRequest(BaseModel):
    training_class: str
    user_conditions: UserConditions
    duration_weeks: int = 52

class WorkoutDay(BaseModel):
    day: int
    workout_type: str
    intensity: str
    duration_minutes: int

class TrainingWeek(BaseModel):
    week: int
    schedule: List[WorkoutDay]
    notes: List[str]

class GenerateResponse(BaseModel):
    program: List[TrainingWeek]
    summary: dict

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/generate", response_model=GenerateResponse)
async def generate(request: GenerateRequest):
    try:
        program = generator.generate_program(
            training_class=request.training_class,
            user_conditions=request.user_conditions.dict(),
            duration_weeks=request.duration_weeks
        )
        
        # Статистика программы
        total_sessions = sum(1 for week in program for day in week['schedule'] if day['workout_type'] != 'rest')
        
        return GenerateResponse(
            program=program,
            summary={
                "total_weeks": len(program),
                "total_training_sessions": total_sessions,
                "workout_types_distribution": {
                    "cardio": sum(1 for w in program for d in w['schedule'] if d['workout_type'] == 'cardio'),
                    "strength": sum(1 for w in program for d in w['schedule'] if d['workout_type'] == 'strength'),
                    "hiit": sum(1 for w in program for d in w['schedule'] if d['workout_type'] == 'hiit'),
                    "recovery": sum(1 for w in program for d in w['schedule'] if d['workout_type'] == 'recovery'),
                    "rest": sum(1 for w in program for d in w['schedule'] if d['workout_type'] == 'rest')
                }
            }
        )
    except Exception as e:
        logger.error(f"Generation error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8002)