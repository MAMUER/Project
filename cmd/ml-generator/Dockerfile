FROM python:3.11-slim

WORKDIR /app

# Установка зависимостей
COPY cmd/ml-generator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода
COPY cmd/ml-generator/main.py .
COPY cmd/ml-generator/train_gan.py .

# Создание директории для моделей
RUN mkdir -p /app/models

# Копирование моделей из корневой директории
COPY models/generator.keras /app/models/generator.keras
COPY models/generator.h5 /app/models/generator.h5

EXPOSE 8002

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8002"]