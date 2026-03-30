"""
ML Generator API Service
Generates personalized training plans using GAN
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import Optional, List, Dict
import numpy as np
import tensorflow as tf
from tensorflow import keras
import os
import json

app = FastAPI(
    title="ML Generator Service",
    description="Generates personalized training plans using GAN",
    version="1.0.0"
)

# Global variable
generator = None

TRAINING_CLASSES = {
    0: 'recovery',
    1: 'endurance_e1e2',
    2: 'threshold_e3',
    3: 'strength_hiit'
}

TRAINING_TEMPLATES = {
    'recovery': {
        'duration_range': (20, 45),
        'intensity_range': (0.3, 0.5),
        'exercises': ['walking', 'yoga', 'stretching', 'light_swimming', 'mobility'],
        'rest_ratio': 0.7,
        'name_ru': 'Восстановление'
    },
    'endurance_e1e2': {
        'duration_range': (45, 90),
        'intensity_range': (0.5, 0.7),
        'exercises': ['running', 'cycling', 'swimming', 'rowing', 'hiking'],
        'rest_ratio': 0.4,
        'name_ru': 'Базовая выносливость'
    },
    'threshold_e3': {
        'duration_range': (30, 60),
        'intensity_range': (0.7, 0.85),
        'exercises': ['tempo_run', 'threshold_intervals', 'fartlek', 'critical_power'],
        'rest_ratio': 0.3,
        'name_ru': 'Пороговая выносливость'
    },
    'strength_hiit': {
        'duration_range': (20, 45),
        'intensity_range': (0.85, 1.0),
        'exercises': ['hiit', 'strength', 'sprints', 'crossfit', 'plyometrics'],
        'rest_ratio': 0.5,
        'name_ru': 'Силовая/HIIT'
    }
}


class UserProfile(BaseModel):
    """User profile for plan generation"""
    gender: str = Field(..., description="Gender (male/female)")
    age: int = Field(..., description="Age", ge=10, le=100)
    fitness_level: str = Field(..., description="Fitness level (beginner/intermediate/advanced)")
    weight: Optional[float] = Field(None, description="Weight (kg)", ge=30, le=200)
    height: Optional[float] = Field(None, description="Height (cm)", ge=100, le=250)
    health_conditions: Optional[List[str]] = Field(None, description="Health conditions")
    goals: Optional[List[str]] = Field(None, description="Training goals")
    lifestyle: Optional[Dict] = Field(None, description="Lifestyle factors (nutrition, sleep, etc.)")


class PlanGenerationRequest(BaseModel):
    """Request for training plan generation"""
    training_class: str = Field(..., description="Training class from classifier")
    user_profile: UserProfile
    preferences: Optional[Dict] = Field(None, description="User preferences (time, equipment, etc.)")


class Exercise(BaseModel):
    """Exercise details"""
    name: str
    duration_minutes: int
    intensity: float


class TrainingPlan(BaseModel):
    """Generated training plan"""
    training_type: str
    training_type_ru: str
    duration_minutes: int
    intensity: float
    weekly_frequency: int
    primary_exercise: str
    warmup_minutes: int
    cooldown_minutes: int
    exercises: List[str]
    session_structure: List[Exercise]
    notes: List[str]
    weekly_schedule: Optional[Dict] = None


def load_generator():
    """Load trained generator model"""
    global generator
    
    model_path = '../../models/generator.h5'
    
    if os.path.exists(model_path):
        generator = keras.models.load_model(model_path)
        print(f"Generator loaded from {model_path}")
    else:
        print(f"Generator not found at {model_path}")


@app.on_event("startup")
async def startup_event():
    """Load generator on startup"""
    load_generator()


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "generator_loaded": generator is not None
    }


@app.get("/templates")
async def get_templates():
    """Get training templates"""
    return TRAINING_TEMPLATES


def encode_user_profile(profile: UserProfile) -> np.ndarray:
    """Encode user profile to model input (10 dimensions)"""
    # Normalize features
    age_norm = (profile.age - 10) / 90
    fitness_map = {'beginner': 0.3, 'intermediate': 0.6, 'advanced': 0.9}
    fitness_norm = fitness_map.get(profile.fitness_level, 0.5)
    
    weight_norm = (profile.weight or 70) / 200
    height_norm = (profile.height or 170) / 250
    
    # Goal encoding
    goal_encoded = 0.5
    if profile.goals:
        goals_lower = [g.lower() for g in profile.goals]
        if 'похудение' in goals_lower or 'weight_loss' in goals_lower:
            goal_encoded = 0.2
        elif 'набор массы' in goals_lower or 'muscle_gain' in goals_lower:
            goal_encoded = 0.8
        elif 'реабилитация' in goals_lower or 'rehabilitation' in goals_lower:
            goal_encoded = 0.1
    
    health_flag = 1.0 if profile.health_conditions else 0.0
    gender_encoded = 1.0 if profile.gender.lower() == 'male' else 0.0
    
    # Lifestyle factors
    sleep_score = 0.5
    nutrition_score = 0.5
    if profile.lifestyle:
        sleep_score = profile.lifestyle.get('sleep_hours', 7) / 10
        nutrition_score = profile.lifestyle.get('nutrition_quality', 0.5)
    
    encoded = np.array([
        age_norm,
        fitness_norm,
        weight_norm,
        height_norm,
        goal_encoded,
        health_flag,
        gender_encoded,
        sleep_score,
        nutrition_score,
        0.5  # Reserved
    ])
    
    return encoded.reshape(1, -1)


def decode_plan(plan_vector: np.ndarray, training_class: str, user_profile: UserProfile) -> dict:
    """Decode GAN output (16 dimensions) to training plan"""
    template = TRAINING_TEMPLATES.get(training_class, TRAINING_TEMPLATES['endurance_e1e2'])
    
    duration = int(plan_vector[0] * 100)
    intensity = plan_vector[1]
    rest_ratio = plan_vector[2]
    weekly_freq = int(plan_vector[3] * 7)
    
    exercise_probs = plan_vector[4:9]
    primary_exercise_idx = np.argmax(exercise_probs)
    primary_exercise = template['exercises'][primary_exercise_idx % len(template['exercises'])]
    
    warmup = int(plan_vector[9] * 100)
    cooldown = int(plan_vector[10] * 100)
    
    # Build session structure
    session_structure = [
        Exercise(name="Разминка", duration_minutes=max(5, min(20, warmup)), intensity=0.3),
        Exercise(name=primary_exercise, duration_minutes=int(duration * 0.6), intensity=intensity),
        Exercise(name="Заминка", duration_minutes=max(5, min(20, cooldown)), intensity=0.3)
    ]
    
    # Build notes
    notes = []
    if user_profile.fitness_level == 'beginner':
        notes.append("Начните с 50% от рекомендованной интенсивности")
        duration = int(duration * 0.7)
    
    if user_profile.age > 50:
        notes.append("Увеличьте время разминки и заминки")
    
    if user_profile.health_conditions:
        notes.append("Проконсультируйтесь с врачом перед началом")
    
    if user_profile.goals:
        goals_lower = [g.lower() for g in user_profile.goals]
        if 'похудение' in goals_lower:
            notes.append("Добавьте 10-15 минут кардио после основной тренировки")
        if 'набор массы' in goals_lower:
            notes.append("Сфокусируйтесь на силовых упражнениях")
        if 'реабилитация' in goals_lower:
            notes.append("Следите за техникой выполнения упражнений")
    
    # Weekly schedule
    weekly_schedule = {
        'monday': primary_exercise if weekly_freq >= 1 else 'rest',
        'wednesday': primary_exercise if weekly_freq >= 2 else 'rest',
        'friday': primary_exercise if weekly_freq >= 3 else 'rest',
        'saturday': 'active_recovery' if weekly_freq >= 4 else 'rest',
        'sunday': 'rest'
    }
    
    return {
        'training_type': training_class,
        'training_type_ru': template['name_ru'],
        'duration_minutes': max(20, min(120, duration)),
        'intensity': round(float(intensity), 2),
        'weekly_frequency': max(1, min(7, weekly_freq)),
        'primary_exercise': primary_exercise,
        'warmup_minutes': max(5, min(20, warmup)),
        'cooldown_minutes': max(5, min(20, cooldown)),
        'exercises': template['exercises'],
        'session_structure': [e.dict() for e in session_structure],
        'notes': notes,
        'weekly_schedule': weekly_schedule
    }


@app.post("/generate-plan", response_model=TrainingPlan)
async def generate_plan(request: PlanGenerationRequest):
    """Generate personalized training plan"""
    if generator is None:
        raise HTTPException(status_code=503, detail="Generator not loaded")
    
    try:
        # Get training class index
        class_idx = list(TRAINING_CLASSES.values()).index(request.training_class)
        
        # Encode user profile
        profile_encoded = encode_user_profile(request.user_profile)
        
        # Encode training class
        class_onehot = keras.utils.to_categorical([class_idx], 4)
        
        # Generate noise
        noise = np.random.normal(0, 1, (1, 32))
        
        # Generate plan
        plan_vector = generator.predict([noise, profile_encoded, class_onehot], verbose=0)[0]
        
        # Decode to human-readable format
        plan = decode_plan(plan_vector, request.training_class, request.user_profile)
        
        return TrainingPlan(**plan)
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8002)