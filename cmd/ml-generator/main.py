import os
import json
import random
import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional
import uvicorn
import logging
from datetime import datetime, timedelta

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("ml-generator")

app = FastAPI(title="ML Generator Service (GAN)", version="1.0.0")

# Модели данных
class GenerateRequest(BaseModel):
    class_name: str = Field(..., description="Класс тренировки: endurance, strength, recovery, interval")
    confidence: float = Field(default=0.8, ge=0, le=1)
    duration_weeks: int = Field(default=4, ge=1, le=52)
    available_days: List[int] = Field(default=[1, 3, 5])
    user_goals: Optional[List[str]] = None
    fitness_level: Optional[str] = "intermediate"
    contraindications: Optional[List[str]] = None

class GenerateResponse(BaseModel):
    plan_id: str
    plan_data: dict

# База упражнений
EXERCISES = {
    "endurance": {
        "cardio": ["Бег", "Велосипед", "Плавание", "Гребля", "Скакалка", "Эллипс"],
        "strength": ["Выпады", "Приседания", "Отжимания", "Подтягивания", "Планка"],
        "stretching": ["Растяжка ног", "Растяжка спины", "Йога-позы"]
    },
    "strength": {
        "cardio": ["Бурпи", "Спринт", "Бег", "Велосипед"],
        "strength": ["Жим лежа", "Приседания со штангой", "Становая тяга", "Подтягивания", "Отжимания на брусьях", "Тяга штанги"],
        "stretching": ["Растяжка грудных", "Растяжка спины", "Динамическая растяжка"]
    },
    "recovery": {
        "cardio": ["Ходьба", "Легкий бег", "Плавание", "Йога"],
        "strength": ["Растяжка", "Пилатес", "Мобильность"],
        "stretching": ["Йога", "Растяжка всех групп мышц", "Дыхательные практики"]
    },
    "interval": {
        "cardio": ["Интервальный бег", "Табата", "Бурпи-интервалы", "Велоспринты"],
        "strength": ["Круговая тренировка", "AMRAP", "EMOM"],
        "stretching": ["Динамическая растяжка", "Заминка"]
    }
}

INTENSITY = {
    "low": {"duration": 20, "sets": 2, "reps": 10},
    "medium": {"duration": 35, "sets": 3, "reps": 12},
    "high": {"duration": 50, "sets": 4, "reps": 15}
}

def generate_workout(day: int, class_name: str, intensity: str, fitness_level: str) -> dict:
    """Генерация одной тренировки"""
    exercises = EXERCISES.get(class_name, EXERCISES["endurance"])
    intensity_cfg = INTENSITY.get(intensity, INTENSITY["medium"])
    
    # Выбор упражнений
    cardio = random.sample(exercises["cardio"], min(2, len(exercises["cardio"])))
    strength = random.sample(exercises["strength"], min(3, len(exercises["strength"])))
    stretching = random.sample(exercises["stretching"], min(2, len(exercises["stretching"])))
    
    workout = {
        "day": day,
        "type": class_name,
        "intensity": intensity,
        "duration_minutes": intensity_cfg["duration"],
        "warmup": ["Разминка суставов", "Легкий бег 5 минут"],
        "main_part": [
            {"name": ex, "sets": intensity_cfg["sets"], "reps": intensity_cfg["reps"]}
            for ex in strength
        ],
        "cardio": [{"name": ex, "minutes": intensity_cfg["duration"] // 3} for ex in cardio],
        "cooldown": stretching,
        "notes": [
            "Следите за пульсом",
            "Пейте воду во время тренировки",
            "При болях прекратите упражнение"
        ]
    }
    
    # Корректировка для начинающих
    if fitness_level == "beginner":
        workout["duration_minutes"] = max(15, workout["duration_minutes"] - 10)
        for item in workout["main_part"]:
            item["sets"] = max(2, item["sets"] - 1)
            item["reps"] = max(8, item["reps"] - 2)
    
    return workout

def generate_plan(request: GenerateRequest) -> dict:
    """Генерация полной программы тренировок"""
    # Определяем интенсивность на основе класса и confidence
    intensity_map = {
        "endurance": "medium",
        "strength": "high",
        "recovery": "low",
        "interval": "high"
    }
    intensity = intensity_map.get(request.class_name, "medium")
    
    # Корректировка по confidence
    if request.confidence < 0.6:
        intensity = "low"
    elif request.confidence > 0.9:
        intensity = "high"
    
    # Генерация недель
    weeks = []
    for week in range(1, request.duration_weeks + 1):
        week_data = {
            "week": week,
            "focus": f"Неделя {week}: {request.class_name}",
            "workouts": []
        }
        
        # Генерация тренировок на указанные дни
        for day in request.available_days:
            workout = generate_workout(day, request.class_name, intensity, request.fitness_level)
            week_data["workouts"].append(workout)
        
        weeks.append(week_data)
    
    # Общая информация
    plan = {
        "name": f"Программа тренировок: {request.class_name}",
        "class": request.class_name,
        "confidence": request.confidence,
        "duration_weeks": request.duration_weeks,
        "intensity": intensity,
        "available_days": request.available_days,
        "goals": request.user_goals or ["Общее укрепление здоровья"],
        "weeks": weeks,
        "recommendations": [
            "Занимайтесь в комфортном темпе",
            "Следите за самочувствием",
            "При появлении боли обратитесь к врачу",
            "Пейте воду до, во время и после тренировки",
            "Делайте разминку и заминку"
        ]
    }
    
    # Добавляем рекомендации по противопоказаниям
    if request.contraindications:
        plan["contraindications_warning"] = [
            f"⚠️ Учитывая противопоказание: {c}. Проконсультируйтесь с врачом перед началом."
            for c in request.contraindications
        ]
    
    return plan

@app.get("/health")
async def health():
    return {"status": "ok", "service": "ml-generator", "timestamp": datetime.now().isoformat()}

@app.post("/generate", response_model=GenerateResponse)
async def generate(request: GenerateRequest):
    logger.info(f"Generating plan for class: {request.class_name}, weeks: {request.duration_weeks}")
    
    try:
        plan_data = generate_plan(request)
        
        # Генерируем уникальный ID
        import uuid
        plan_id = str(uuid.uuid4())
        
        return GenerateResponse(
            plan_id=plan_id,
            plan_data=plan_data
        )
        
    except Exception as e:
        logger.error(f"Generation error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def root():
    return {
        "service": "ML Generator (GAN)",
        "version": "1.0.0",
        "endpoints": {
            "generate": "POST /generate - Генерация программы тренировок",
            "health": "GET /health - Проверка здоровья"
        }
    }

if __name__ == "__main__":
    port = int(os.getenv("ML_GENERATOR_PORT", 8002))
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="info")