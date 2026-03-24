import asyncio
import numpy as np
import tensorflow as tf
from fastapi import FastAPI, HTTPException
import uvicorn
import logging
from pydantic import BaseModel
from typing import Optional

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="HealthFit Classification Service", version="1.0.0")

# Модель нейросети (6 входов -> класс тренировки)
class TrainingClassifier:
    def __init__(self):
        # Сначала определяем классы, потом строим модель
        self.classes = ['cardio', 'strength', 'flexibility', 'recovery', 'hiit', 'endurance']
        self.model = self._build_model()
        
    def _build_model(self):
        model = tf.keras.Sequential([
            tf.keras.layers.Dense(32, activation='relu', input_shape=(6,)),
            tf.keras.layers.Dropout(0.2),
            tf.keras.layers.Dense(16, activation='relu'),
            tf.keras.layers.Dropout(0.2),
            tf.keras.layers.Dense(len(self.classes), activation='softmax')
        ])
        model.compile(optimizer='adam', loss='categorical_crossentropy', metrics=['accuracy'])
        return model
    
    def predict(self, features):
        input_data = np.array([[
            features['heart_rate'],
            features['ecg_mean'],
            features['systolic'],
            features['diastolic'],
            features['spo2'],
            features['temperature']
        ]])
        
        pred = self.model.predict(input_data, verbose=0)
        class_idx = np.argmax(pred[0])
        confidence = float(pred[0][class_idx])
        
        return self.classes[class_idx], confidence

classifier = TrainingClassifier()

# Pydantic модели
class ClassifyRequest(BaseModel):
    heart_rate: int
    ecg: str
    systolic: int
    diastolic: int
    spo2: int
    temperature: float
    sleep_duration: int
    deep_sleep: int

class ClassifyResponse(BaseModel):
    class_name: str
    confidence: float
    analysis_data: dict

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/classify", response_model=ClassifyResponse)
async def classify(request: ClassifyRequest):
    try:
        ecg_values = [int(x) for x in request.ecg.split(',')] if request.ecg else [0]
        ecg_mean = sum(ecg_values) / len(ecg_values) if ecg_values else 0
        
        features = {
            'heart_rate': request.heart_rate,
            'ecg_mean': ecg_mean,
            'systolic': request.systolic,
            'diastolic': request.diastolic,
            'spo2': request.spo2,
            'temperature': request.temperature
        }
        
        class_name, confidence = classifier.predict(features)
        
        return ClassifyResponse(
            class_name=class_name,
            confidence=confidence,
            analysis_data={
                "heart_rate_status": "normal" if 60 <= request.heart_rate <= 100 else "abnormal",
                "spo2_status": "normal" if request.spo2 >= 95 else "low",
                "recommended_intensity": "moderate" if class_name in ['cardio', 'endurance'] else "light"
            }
        )
    except Exception as e:
        logger.error(f"Classification error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)