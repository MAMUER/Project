import os
import json
import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional
import uvicorn
import logging
from datetime import datetime

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("ml-classifier")

app = FastAPI(title="ML Classifier Service", version="1.0.0")

# Модель данных для запроса
class BiometricFeatures(BaseModel):
    heart_rate: float = Field(..., ge=30, le=250, description="Пульс (уд/мин)")
    ecg: float = Field(..., ge=0, le=2, description="ЭКГ индекс (нормализованный)")
    blood_pressure_systolic: float = Field(..., ge=70, le=250, description="Давление систолическое")
    blood_pressure_diastolic: float = Field(..., ge=40, le=150, description="Давление диастолическое")
    spo2: float = Field(..., ge=70, le=100, description="Сатурация кислорода (%)")
    temperature: float = Field(..., ge=35, le=42, description="Температура тела (°C)")
    sleep_hours: float = Field(..., ge=0, le=24, description="Длительность сна (часы)")

class UserContext(BaseModel):
    age: Optional[int] = None
    gender: Optional[str] = None
    fitness_level: Optional[str] = None
    goals: Optional[List[str]] = None
    contraindications: Optional[List[str]] = None

class ClassifyRequest(BaseModel):
    features: BiometricFeatures
    user_context: Optional[UserContext] = None

class ClassifyResponse(BaseModel):
    class_name: str
    confidence: float
    intensity: str
    recommendations: List[str]

# Карта классов тренировок
CLASSES = {
    "endurance": {
        "name": "Выносливость",
        "intensity": "medium",
        "recommendations": [
            "Увеличьте продолжительность кардио-тренировок",
            "Добавьте интервальные нагрузки",
            "Следите за пульсом в зоне 60-70% от максимума"
        ]
    },
    "strength": {
        "name": "Силовая",
        "intensity": "high",
        "recommendations": [
            "Увеличьте рабочий вес на 5-10%",
            "Сократите отдых между подходами",
            "Добавьте базовые упражнения"
        ]
    },
    "recovery": {
        "name": "Восстановление",
        "intensity": "low",
        "recommendations": [
            "Снизьте интенсивность тренировок",
            "Увеличьте время сна",
            "Добавьте растяжку и массаж"
        ]
    },
    "interval": {
        "name": "Интервальная",
        "intensity": "high",
        "recommendations": [
            "Чередуйте высокую и низкую интенсивность",
            "Следите за восстановлением между интервалами",
            "Используйте пульсометр"
        ]
    }
}

def predict_class(features: BiometricFeatures) -> dict:
    """
    Простая логика классификации на основе правил.
    В реальном проекте здесь будет нейросеть (TensorFlow/PyTorch)
    """
    score = {
        "endurance": 0,
        "strength": 0,
        "recovery": 0,
        "interval": 0
    }
    
    # Правила для endurance (выносливость)
    if 120 <= features.heart_rate <= 160 and features.sleep_hours >= 7:
        score["endurance"] += 30
    if features.spo2 >= 96:
        score["endurance"] += 20
    
    # Правила для strength (силовая)
    if features.heart_rate <= 140 and features.blood_pressure_systolic >= 110:
        score["strength"] += 30
    if features.sleep_hours >= 8:
        score["strength"] += 20
    
    # Правила для recovery (восстановление)
    if features.sleep_hours < 6 or features.heart_rate > 80:
        score["recovery"] += 40
    if features.temperature > 37:
        score["recovery"] += 30
    
    # Правила для interval (интервальная)
    if features.heart_rate > 150 and features.spo2 >= 95:
        score["interval"] += 40
    if features.sleep_hours >= 7:
        score["interval"] += 20
    
    # Выбираем класс с максимальным score
    best_class = max(score, key=score.get)
    confidence = score[best_class] / 100 if score[best_class] > 0 else 0.5
    
    return {
        "class_name": best_class,
        "confidence": confidence,
        "intensity": CLASSES[best_class]["intensity"],
        "recommendations": CLASSES[best_class]["recommendations"]
    }

@app.get("/health")
async def health():
    return {"status": "ok", "service": "ml-classifier", "timestamp": datetime.now().isoformat()}

@app.post("/classify", response_model=ClassifyResponse)
async def classify(request: ClassifyRequest):
    logger.info(f"Classifying features: heart_rate={request.features.heart_rate}")
    
    try:
        result = predict_class(request.features)
        
        # Учитываем контекст пользователя
        if request.user_context:
            if request.user_context.contraindications:
                if "heart" in str(request.user_context.contraindications).lower():
                    result["intensity"] = "low"
                    result["recommendations"].append("Учитывая сердечные ограничения, снизьте нагрузку")
            
            if request.user_context.fitness_level == "beginner":
                result["intensity"] = "low"
            elif request.user_context.fitness_level == "advanced":
                result["intensity"] = "high"
        
        return ClassifyResponse(**result)
        
    except Exception as e:
        logger.error(f"Classification error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def root():
    return {
        "service": "ML Classifier",
        "version": "1.0.0",
        "endpoints": {
            "classify": "POST /classify - Классифицировать состояние пользователя",
            "health": "GET /health - Проверка здоровья"
        }
    }

if __name__ == "__main__":
    port = int(os.getenv("ML_CLASSIFIER_PORT", 8001))
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="info")