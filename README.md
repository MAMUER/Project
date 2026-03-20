# Архитектура системы FitHealth AI Platform

## 1. Обзор системы

FitHealth AI Platform — интеллектуальная веб-платформа для персонифицированного управления физической активностью и предиктивной оценки рисков здоровью на основе анализа биометрических данных.

### Целевая аудитория
- **Фитнес-центры** — управление клиентами, программами тренировок, оборудованием
- **Реабилитационные центры** — мониторинг восстановления пациентов
- **Индивидуальные пользователи** — персональный фитнес-ассистент

---

## 2. Диаграмма компонентов (Target Architecture)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Web App   │  │  iOS App    │  │ Android App │  │  Admin Panel│        │
│  │  (Next.js)  │  │  (Swift)    │  │  (Kotlin)   │  │  (React)    │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
└─────────┼────────────────┼────────────────┼────────────────┼────────────────┘
          │                │                │                │
          └────────────────┴────────────────┴────────────────┘
                                    │
                              API Gateway (Kong/Nginx)
                                    │
┌───────────────────────────────────┼─────────────────────────────────────────┐
│                          MICROSERVICES LAYER (Go)                           │
│                                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   Auth       │  │   User       │  │  Biometric   │  │  Training    │   │
│  │   Service    │  │   Service    │  │   Service    │  │   Service    │   │
│  │   :8001      │  │   :8002      │  │   :8003      │  │   :8004      │   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │
│         │                 │                 │                 │            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   Device     │  │ Notification │  │  Achievement │  │   Report     │   │
│  │   Gateway    │  │   Service    │  │   Service    │  │   Service    │   │
│  │   :8005      │  │   :8006      │  │   :8007      │  │   :8008      │   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │
│         │                 │                 │                 │            │
└─────────┼─────────────────┼─────────────────┼─────────────────┼────────────┘
          │                 │                 │                 │
          └─────────────────┴─────────────────┴─────────────────┘
                                    │
                            gRPC / REST API
                                    │
┌───────────────────────────────────┼─────────────────────────────────────────┐
│                           ML SERVICES LAYER (Python)                        │
│                                                                             │
│  ┌──────────────────────────┐  ┌──────────────────────────┐               │
│  │   Classification API     │  │   Program Generator API  │               │
│  │   (FastAPI + TensorFlow) │  │   (FastAPI + GAN)        │               │
│  │   :9001                  │  │   :9002                  │               │
│  │                          │  │                          │               │
│  │   ┌──────────────────┐   │  │   ┌──────────────────┐   │               │
│  │   │  Neural Network  │   │  │   │   Generator G    │   │               │
│  │   │  (6 inputs)      │   │  │   │   Discriminator D│   │               │
│  │   └──────────────────┘   │  │   └──────────────────┘   │               │
│  └──────────────────────────┘  └──────────────────────────┘               │
│                                                                             │
│  ┌──────────────────────────┐  ┌──────────────────────────┐               │
│  │   Risk Assessment API    │  │   Recommendation Engine  │               │
│  │   (FastAPI + PyTorch)    │  │   (FastAPI + scikit)     │               │
│  │   :9003                  │  │   :9004                  │               │
│  └──────────────────────────┘  └──────────────────────────┘               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
┌───────────────────────────────────┼─────────────────────────────────────────┐
│                          DATA LAYER                                         │
│                                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  PostgreSQL  │  │    Redis     │  │   RabbitMQ   │  │    MinIO     │   │
│  │  (Primary)   │  │   (Cache)    │  │   (Queue)    │  │  (Storage)   │   │
│  │   :5432      │  │   :6379      │  │   :5672      │  │   :9000      │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Структура репозитория

```
fithealth-platform/
├── cmd/                          # Точки входа сервисов
│   ├── api-gateway/              # API Gateway
│   ├── auth-service/             # Сервис аутентификации
│   ├── user-service/             # Сервис пользователей
│   ├── biometric-service/        # Сервис биометрии
│   ├── training-service/         # Сервис тренировок
│   ├── device-gateway/           # Шлюз устройств
│   └── notification-service/     # Сервис уведомлений
│
├── internal/                     # Внутренний код сервисов
│   ├── auth/
│   │   ├── handler/              # HTTP/gRPC handlers
│   │   ├── service/              # Бизнес-логика
│   │   ├── repository/           # Доступ к данным
│   │   └── model/                # Модели данных
│   ├── user/
│   ├── biometric/
│   ├── training/
│   └── ...
│
├── pkg/                          # Общие пакеты
│   ├── database/                 # Подключение к БД
│   ├── cache/                    # Redis клиент
│   ├── queue/                    # RabbitMQ клиент
│   ├── logger/                   # Логирование
│   ├── middleware/               # Общие middleware
│   └── utils/                    # Утилиты
│
├── proto/                        # Protobuf определения
│   ├── auth.proto
│   ├── user.proto
│   ├── biometric.proto
│   └── training.proto
│
├── ml-services/                  # ML микросервисы (Python)
│   ├── classification/
│   │   ├── app/
│   │   │   ├── main.py           # FastAPI приложение
│   │   │   ├── models/           # TensorFlow модели
│   │   │   ├── routers/          # API роутеры
│   │   │   └── services/         # Бизнес-логика
│   │   ├── Dockerfile
│   │   └── requirements.txt
│   ├── program-generator/
│   │   ├── app/
│   │   │   ├── main.py
│   │   │   ├── gan/              # GAN архитектура
│   │   │   ├── routers/
│   │   │   └── services/
│   │   └── ...
│   └── risk-assessment/
│
├── web/                          # Фронтенд (Next.js)
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   └── types/
│   └── ...
│
├── mobile/                       # Мобильные приложения
│   ├── ios/                      # Swift (HealthKit)
│   └── android/                  # Kotlin (Health Connect)
│
├── deployments/                  # Kubernetes манифесты
│   ├── base/
│   │   ├── deployments/
│   │   ├── services/
│   │   ├── configmaps/
│   │   └── secrets/
│   └── overlays/
│       ├── development/
│       └── production/
│
├── docs/                         # Документация
│   ├── api/                      # API документация
│   ├── architecture/             # Диаграммы
│   └── ml/                       # ML документация
│
├── scripts/                      # Скрипты
│   ├── migrate.sh
│   ├── seed-data.sh
│   └── deploy.sh
│
├── docker-compose.yml            # Локальная разработка
├── Makefile                      # Команды сборки
└── README.md
```

---

## 4. Пошаговый план разработки

### Этап 1: MVP (Месяц 1-2)

**Цель**: Базовая функциональность для индивидуальных пользователей

#### Неделя 1-2: Инфраструктура
- [ ] Настройка Docker Compose для локальной разработки
- [ ] Базовая структура Go микросервисов
- [ ] PostgreSQL схема и миграции
- [ ] Redis для кэширования

#### Неделя 3-4: Auth + User Service
- [ ] JWT аутентификация
- [ ] Регистрация/авторизация
- [ ] Профиль пользователя
- [ ] Ролевая модель (User, Trainer, Admin)

#### Неделя 5-6: Biometric Service
- [ ] API для приема биометрических данных
- [ ] Хранение с шифрованием (AES-256)
- [ ] Агрегация статистики
- [ ] Симуляция данных устройств (для разработки)

#### Неделя 7-8: Web Frontend
- [ ] Next.js дашборд
- [ ] Визуализация биометрии (Recharts)
- [ ] Профиль пользователя
- [ ] Настройки

### Этап 2: ML Integration (Месяц 3-4)

#### Неделя 9-10: Classification Service
- [ ] Python FastAPI сервис
- [ ] Нейросеть (6 входов, классификация 5 классов)
- [ ] gRPC интеграция с Go сервисами
- [ ] API endpoint `/api/classify`

#### Неделя 11-12: Program Generator
- [ ] GAN архитектура для генерации программ
- [ ] База упражнений
- [ ] Учет противопоказаний
- [ ] API endpoint `/api/generate-program`

#### Неделя 13-14: Risk Assessment
- [ ] Модель оценки рисков
- [ ] Интеграция с классификацией
- [ ] Алерты и уведомления

#### Неделя 15-16: Frontend AI Features
- [ ] UI для классификации
- [ ] Просмотр сгенерированных программ
- [ ] Визуализация рисков

### Этап 3: Device Integration (Месяц 5-6)

#### Неделя 17-18: Device Gateway
- [ ] HealthKit интеграция (iOS)
- [ ] Health Connect интеграция (Android)
- [ ] Samsung Health API
- [ ] Huawei Health API

#### Неделя 19-20: Mobile Apps
- [ ] iOS приложение (SwiftUI)
- [ ] Android приложение (Jetpack Compose)
- [ ] Синхронизация с носимыми устройствами

#### Неделя 21-22: Real-time Processing
- [ ] WebSocket для real-time данных
- [ ] Обработка пиковых нагрузок
- [ ] RabbitMQ для асинхронной обработки

#### Неделя 23-24: Testing & Optimization
- [ ] Unit тесты (Go, Python)
- [ ] Integration тесты
- [ ] Нагрузочное тестирование (k6)
- [ ] Оптимизация производительности

### Этап 4: Enterprise Features (Месяц 7-8)

#### Неделя 25-26: Club Management
- [ ] Администрирование фитнес-клубов
- [ ] Управление оборудованием
- [ ] Расписание и запись

#### Неделя 27-28: Trainer Features
- [ ] Управление клиентами тренера
- [ ] Создание программ для клиентов
- [ ] Мониторинг прогресса

#### Неделя 29-30: Gamification
- [ ] Система достижений
- [ ] Рейтинги и соревнования
- [ ] Бейджи и награды

#### Неделя 31-32: Java Integration
- [ ] REST API для интеграции с Java Spring Boot
- [ ] Синхронизация данных
- [ ] Общая аутентификация

### Этап 5: Production (Месяц 9-10)

#### Неделя 33-34: Kubernetes Deployment
- [ ] Helm чарты
- [ ] Автоскейлинг
- [ ] Ingress конфигурация

#### Неделя 35-36: Monitoring & Logging
- [ ] Prometheus метрики
- [ ] Grafana дашборды
- [ ] ELK/Loki логирование
- [ ] Алертинг

#### Неделя 37-38: Security Hardening
- [ ] Security audit
- [ ] Penetration testing
- [ ] GDPR/152-ФЗ compliance
- [ ] Data encryption at rest

#### Неделя 39-40: Documentation & Training
- [ ] API документация (OpenAPI)
- [ ] Пользовательская документация
- [ ] Административная документация

---

## 5. Детали реализации ключевых модулей

### 5.1 Модуль сбора данных с носимых устройств

#### Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Apple Watch│     │Samsung Watch│     │ Huawei Watch│
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  HealthKit  │     │Health Connect│    │Huawei Health│
│   (iOS)     │     │  (Android)   │    │    API      │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    Mobile App (Companion)
                           │
                           ▼
                  ┌─────────────────┐
                  │  Device Gateway │
                  │  (Go Service)   │
                  └────────┬────────┘
                           │
                    gRPC / REST
                           │
                           ▼
                  ┌─────────────────┐
                  │ Biometric Service│
                  └─────────────────┘
```

#### Протоколы и периодичность

| Тип данных | Периодичность | Протокол | Формат |
|------------|---------------|----------|--------|
| Пульс | Каждые 5 мин | REST/gRPC | JSON |
| Шаги | Каждые 15 мин | REST/gRPC | JSON |
| Сон | После пробуждения | REST/gRPC | JSON |
| ЭКГ | По запросу | REST/gRPC | Binary + JSON metadata |
| Давление | Каждые 30 мин | REST/gRPC | JSON |

#### Пример кода Go сервиса

```go
// internal/biometric/handler/grpc.go
package handler

import (
    "context"
    pb "github.com/fithealth/proto/biometric"
)

type BiometricHandler struct {
    service *service.BiometricService
    pb.UnimplementedBiometricServiceServer
}

func (h *BiometricHandler) StreamBiometricData(stream pb.BiometricService_StreamBiometricDataServer) error {
    for {
        data, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.StreamResponse{
                Success: true,
                Message: "Data received",
            })
        }
        if err != nil {
            return status.Errorf(codes.Internal, "stream error: %v", err)
        }

        // Валидация и обработка данных
        if err := h.service.ProcessBiometricData(stream.Context(), data); err != nil {
            log.Printf("Error processing data: %v", err)
        }
    }
}
```

### 5.2 Модуль обработки и хранения биометрических данных

#### Структура таблиц PostgreSQL

```sql
-- Биометрические данные (партиционирование по времени)
CREATE TABLE biometric_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    source VARCHAR(50) NOT NULL, -- apple_watch, samsung_health, etc.
    timestamp TIMESTAMPTZ NOT NULL,
    
    -- Сердечно-сосудистая система
    heart_rate INTEGER CHECK (heart_rate BETWEEN 30 AND 220),
    heart_rate_variability INTEGER,
    resting_heart_rate INTEGER,
    blood_pressure_systolic INTEGER,
    blood_pressure_diastolic INTEGER,
    
    -- Дыхание
    spo2 INTEGER CHECK (spo2 BETWEEN 70 AND 100),
    respiratory_rate INTEGER,
    
    -- Активность
    steps INTEGER,
    distance_km DECIMAL(10, 3),
    calories_burned INTEGER,
    active_minutes INTEGER,
    
    -- Сон
    sleep_duration_minutes INTEGER,
    sleep_quality INTEGER CHECK (sleep_quality BETWEEN 1 AND 100),
    
    -- Зашифрованные сырые данные
    raw_data_encrypted BYTEA,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Создание партиций на месяц
CREATE TABLE biometric_data_2024_01 PARTITION OF biometric_data
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Индексы
CREATE INDEX idx_biometric_user_timestamp ON biometric_data(user_id, timestamp DESC);
CREATE INDEX idx_biometric_source ON biometric_data(user_id, source);
```

### 5.3 Модуль ML-классификации (Python, FastAPI)

#### Архитектура нейросети

```python
# ml-services/classification/app/models/classifier.py
import tensorflow as tf
from tensorflow.keras import layers, models

class FitnessClassifier:
    """
    Нейросеть с 6 входами для классификации состояния пользователя.
    
    Входные параметры:
    1. avg_heart_rate - средний пульс
    2. resting_heart_rate - пульс в покое
    3. sleep_quality - качество сна (0-100)
    4. activity_level - уровень активности (минуты/день)
    5. stress_level - уровень стресса (0-100)
    6. recovery_score - оценка восстановления (0-100)
    
    Выход: 5 классов состояния
    """
    
    def __init__(self):
        self.model = self._build_model()
        
    def _build_model(self):
        model = models.Sequential([
            layers.Input(shape=(6,)),
            layers.BatchNormalization(),
            
            layers.Dense(64, activation='relu'),
            layers.Dropout(0.3),
            
            layers.Dense(32, activation='relu'),
            layers.Dropout(0.2),
            
            layers.Dense(16, activation='relu'),
            
            layers.Dense(5, activation='softmax')  # 5 классов
        ])
        
        model.compile(
            optimizer='adam',
            loss='sparse_categorical_crossentropy',
            metrics=['accuracy']
        )
        
        return model
    
    def predict(self, inputs: np.ndarray) -> dict:
        """
        inputs: массив [avg_hr, resting_hr, sleep, activity, stress, recovery]
        """
        # Нормализация
        normalized = self._normalize(inputs)
        
        # Предсказание
        probabilities = self.model.predict(normalized.reshape(1, -1))
        
        classes = ['excellent', 'good', 'moderate', 'needs_attention', 'at_risk']
        predicted_class = classes[np.argmax(probabilities)]
        confidence = float(np.max(probabilities))
        
        return {
            'fitness_class': predicted_class,
            'confidence': confidence,
            'probabilities': dict(zip(classes, probabilities[0]))
        }
```

#### FastAPI endpoint

```python
# ml-services/classification/app/routers/classify.py
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional

router = APIRouter()

class ClassificationInput(BaseModel):
    avg_heart_rate: int
    resting_heart_rate: int
    sleep_quality: int
    activity_level: int
    stress_level: int
    recovery_score: int

class ClassificationOutput(BaseModel):
    fitness_class: str
    confidence: float
    cardiovascular_risk: str
    metabolic_risk: str
    injury_risk: str
    overtraining_risk: str
    recommendations: list[str]

@router.post("/classify", response_model=ClassificationOutput)
async def classify_user_state(input_data: ClassificationInput):
    try:
        result = await classifier.predict(input_data)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

### 5.4 Модуль генерации программ (Python, GAN)

#### GAN архитектура

```python
# ml-services/program-generator/app/gan/generator.py
import torch
import torch.nn as nn

class ProgramGenerator(nn.Module):
    """
    Generator сеть для создания персонализированных программ тренировок.
    
    Вход: шум + условия (возраст, пол, цель, противопоказания)
    Выход: структура программы тренировок
    """
    
    def __init__(self, latent_dim=100, condition_dim=50):
        super().__init__()
        
        self.latent_dim = latent_dim
        self.condition_dim = condition_dim
        
        # Embedding для категориальных признаков
        self.goal_embedding = nn.Embedding(5, 16)  # 5 целей
        self.level_embedding = nn.Embedding(3, 8)  # 3 уровня
        
        # Generator network
        self.model = nn.Sequential(
            nn.Linear(latent_dim + condition_dim, 256),
            nn.LeakyReLU(0.2),
            nn.BatchNorm1d(256),
            
            nn.Linear(256, 512),
            nn.LeakyReLU(0.2),
            nn.BatchNorm1d(512),
            
            nn.Linear(512, 1024),
            nn.LeakyReLU(0.2),
            nn.BatchNorm1d(1024),
            
            nn.Linear(1024, 2048),  # Выход: закодированная программа
            nn.Tanh()
        )
    
    def forward(self, noise, conditions):
        # Объединяем шум и условия
        x = torch.cat([noise, conditions], dim=1)
        return self.model(x)

class ProgramDiscriminator(nn.Module):
    """Discriminator для оценки качества программы"""
    
    def __init__(self, program_dim=2048, condition_dim=50):
        super().__init__()
        
        self.model = nn.Sequential(
            nn.Linear(program_dim + condition_dim, 1024),
            nn.LeakyReLU(0.2),
            nn.Dropout(0.3),
            
            nn.Linear(1024, 512),
            nn.LeakyReLU(0.2),
            nn.Dropout(0.3),
            
            nn.Linear(512, 256),
            nn.LeakyReLU(0.2),
            
            nn.Linear(256, 1),
            nn.Sigmoid()
        )
    
    def forward(self, program, conditions):
        x = torch.cat([program, conditions], dim=1)
        return self.model(x)
```

### 5.5 Интеграция с Java Spring Boot

```go
// pkg/java_integration/client.go
package java_integration

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type JavaClient struct {
    baseURL    string
    httpClient *http.Client
    authToken  string
}

func NewJavaClient(baseURL string) *JavaClient {
    return &JavaClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// Получение данных о оборудовании клуба
func (c *JavaClient) GetClubEquipment(ctx context.Context, clubID string) ([]Equipment, error) {
    url := fmt.Sprintf("%s/api/clubs/%s/equipment", c.baseURL, clubID)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.authToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var equipment []Equipment
    if err := json.NewDecoder(resp.Body).Decode(&equipment); err != nil {
        return nil, err
    }
    
    return equipment, nil
}

// Отправка данных о тренировке
func (c *JavaClient) SyncTrainingSession(ctx context.Context, session *TrainingSession) error {
    url := fmt.Sprintf("%s/api/training-sessions", c.baseURL)
    
    body, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.authToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to sync session: status %d", resp.StatusCode)
    }
    
    return nil
}
```

---

## 6. CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        
      - name: Python lint
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install ruff && ruff check ml-services/

  test:
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Go tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Run Python tests
        run: |
          pip install -r ml-services/classification/requirements.txt
          pytest ml-services/classification/tests/

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push services
        run: |
          for service in auth user biometric training; do
            docker build -t ghcr.io/fithealth/${service}:${{ github.sha }} -f cmd/${service}-service/Dockerfile .
            docker push ghcr.io/fithealth/${service}:${{ github.sha }}
          done

  deploy:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/auth-service auth=ghcr.io/fithealth/auth:${{ github.sha }}
          kubectl set image deployment/user-service user=ghcr.io/fithealth/user:${{ github.sha }}
          kubectl set image deployment/biometric-service biometric=ghcr.io/fithealth/biometric:${{ github.sha }}
```

---

## 7. Kubernetes Deployment

```yaml
# deployments/base/deployments/biometric-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: biometric-service
  labels:
    app: biometric-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: biometric-service
  template:
    metadata:
      labels:
        app: biometric-service
    spec:
      containers:
        - name: biometric
          image: ghcr.io/fithealth/biometric:latest
          ports:
            - containerPort: 8003
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: fithealth-secrets
                  key: database-url
            - name: REDIS_URL
              valueFrom:
                configMapKeyRef:
                  name: fithealth-config
                  key: redis-url
          livenessProbe:
            httpGet:
              path: /health
              port: 8003
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8003
            initialDelaySeconds: 5
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: biometric-service
spec:
  selector:
    app: biometric-service
  ports:
    - port: 8003
      targetPort: 8003
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: biometric-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: biometric-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

---

## 8. Безопасность и защита данных

### Шифрование биометрических данных

```go
// pkg/encryption/aes.go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

type BiometricEncryptor struct {
    key []byte // 32 байта для AES-256
}

func NewBiometricEncryptor(key string) *BiometricEncryptor {
    return &BiometricEncryptor{
        key: []byte(key),
    }
}

func (e *BiometricEncryptor) Encrypt(plaintext []byte) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *BiometricEncryptor) Decrypt(ciphertext string) ([]byte, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return nil, err
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

### Соответствие 152-ФЗ

1. **Согласие на обработку** — обязательное согласие при регистрации
2. **Минимизация данных** — сбор только необходимых данных
3. **Право на удаление** — возможность полного удаления данных
4. **Локализация** — данные хранятся на серверах в РФ
5. **Аудит доступа** — логирование всех операций с персональными данными

---

## 9. Тестирование

### Unit тесты (Go)

```go
// internal/biometric/service/biometric_test.go
package service_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestProcessBiometricData(t *testing.T) {
    tests := []struct {
        name    string
        input   *pb.BiometricData
        wantErr bool
    }{
        {
            name: "valid heart rate data",
            input: &pb.BiometricData{
                UserId:    "user-123",
                Source:    "apple_watch",
                Timestamp: time.Now().Unix(),
                HeartRate: 72,
            },
            wantErr: false,
        },
        {
            name: "invalid heart rate",
            input: &pb.BiometricData{
                UserId:    "user-123",
                HeartRate: 300, // Невозможное значение
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewBiometricService(mockRepo)
            err := service.ProcessBiometricData(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Нагрузочное тестирование (k6)

```javascript
// tests/load/biometric.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 100 },   // Ramp up
        { duration: '5m', target: 100 },   // Steady state
        { duration: '2m', target: 200 },   // Spike
        { duration: '2m', target: 0 },     // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% запросов < 500ms
        http_req_failed: ['rate<0.01'],   // <1% ошибок
    },
};

export default function() {
    const payload = JSON.stringify({
        user_id: 'user-123',
        source: 'apple_watch',
        heart_rate: Math.floor(Math.random() * 40) + 60,
        timestamp: Date.now(),
    });
    
    const res = http.post('http://api.fithealth.local/api/biometrics', payload, {
        headers: { 'Content-Type': 'application/json' },
    });
    
    check(res, {
        'status is 201': (r) => r.status === 201,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    sleep(1);
}
```

---

## 10. Мониторинг

### Prometheus метрики

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    BiometricDataReceived = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "biometric_data_received_total",
            Help: "Total number of biometric data points received",
        },
        []string{"source", "user_id"},
    )
    
    ClassificationLatency = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "classification_latency_seconds",
            Help:    "Latency of classification requests",
            Buckets: prometheus.DefBuckets,
        },
    )
    
    ActiveUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users_total",
            Help: "Number of currently active users",
        },
    )
)
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "FitHealth Platform",
    "panels": [
      {
        "title": "Biometric Data Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(biometric_data_received_total[5m])",
            "legendFormat": "{{source}}"
          }
        ]
      },
      {
        "title": "Classification Latency",
        "type": "heatmap",
        "targets": [
          {
            "expr": "classification_latency_seconds"
          }
        ]
      }
    ]
  }
}
```

---

## Заключение

Данная архитектура обеспечивает:

1. **Масштабируемость** — микросервисы независимо масштабируются
2. **Надежность** — резервирование критических компонентов
3. **Безопасность** — шифрование данных, соответствие законодательству
4. **Производительность** — кэширование, асинхронная обработка
5. **Гибкость** — возможность интеграции с внешними системами

Прототип на Next.js демонстрирует ключевые концепции и может служить основой для развития полноценной платформы.
