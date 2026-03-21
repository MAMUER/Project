from pydantic import BaseModel, Field
from typing import Optional, List

class ClassifyRequest(BaseModel):
    heart_rate: int = Field(..., ge=30, le=220, description="Пульс (уд/мин)")
    ecg: Optional[str] = Field(None, description="ЭКГ данные в формате CSV")
    systolic: int = Field(..., ge=70, le=200, description="Систолическое давление")
    diastolic: int = Field(..., ge=40, le=120, description="Диастолическое давление")
    spo2: int = Field(..., ge=70, le=100, description="Уровень кислорода в крови")
    temperature: float = Field(..., ge=35, le=42, description="Температура тела")
    sleep_duration: Optional[int] = Field(None, description="Длительность сна (минуты)")
    deep_sleep: Optional[int] = Field(None, description="Глубокий сон (минуты)")

class ClassifyResponse(BaseModel):
    class_name: str = Field(..., description="Рекомендуемый класс тренировки")
    confidence: float = Field(..., ge=0, le=1, description="Уверенность модели")
    analysis_data: dict = Field(default_factory=dict, description="Дополнительная аналитика")

class HealthResponse(BaseModel):
    status: str = Field(default="ok")
    model_loaded: bool = Field(default=False)