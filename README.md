# HealthFit Platform

Интеллектуальная веб-платформа для персонифицированного управления физической активностью и предиктивной оценки рисков здоровью на основе анализа биометрических данных.

## Функциональность

- Сбор биометрических данных с носимых устройств
- Анализ функционального состояния с помощью нейросетей
- Генерация персонализированных программ тренировок (GAN)
- Ролевая модель: пользователь, тренер, администратор
- Соревновательные элементы (достижения, статистика)

## Технологический стек

| Компонент | Технология |
|-----------|------------|
| Backend | Go 1.25, gRPC, Gorilla Mux |
| ML | Python 3.11, TensorFlow, FastAPI |
| База данных | PostgreSQL, Redis |
| Очереди | RabbitMQ |
| Frontend | React 18, TypeScript, Redux Toolkit |
| Контейнеризация | Docker, Kubernetes |
| CI/CD | GitHub Actions |

## Быстрый старт

### Предварительные требования

- Go 1.25+
- Python 3.11+
- Docker и Docker Compose
- Node.js 20+ (для фронтенда)

### Запуск

```bash
# Клонирование репозитория
git clone https://github.com/mamuer/healthfit-platform.git
cd healthfit-platform

# Установка зависимостей
make deps

# Запуск инфраструктуры
make docker-up

# Запуск всех сервисов
make run

Доступ к сервисам
Сервис	URL
API Gateway	http://localhost:8080
Frontend	http://localhost:3000
RabbitMQ UI	http://localhost:15672 (guest/guest)
PostgreSQL	localhost:5432 (healthfit/healthfit123)
ML Classification	http://localhost:8001
ML Generation	http://localhost:8002
Структура проекта
'''
healthfit-platform/
├── cmd/                    # Точки входа Go сервисов
├── internal/               # Внутренняя логика
├── pkg/                    # Публичные библиотеки
├── proto/                  # Protobuf контракты
├── ml/                     # Python ML сервисы
├── web/                    # React фронтенд
├── deploy/                 # Docker и Kubernetes
└── docs/                   # Документация
'''
API Endpoints
Аутентификация
POST /api/v1/auth/login - вход

GET /api/v1/auth/verify - проверка токена

Биометрия
POST /api/v1/biometric - отправка данных

GET /api/v1/biometric/{user_id} - история

Тренировки
POST /api/v1/training/generate - генерация программы

GET /api/v1/training/programs/{user_id} - программы пользователя

Лицензия
MIT