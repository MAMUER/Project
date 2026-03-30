FROM python:3.11-slim

WORKDIR /app

# Установка зависимостей
COPY cmd/ml-classifier/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода
COPY cmd/ml-classifier/main.py .
COPY cmd/ml-classifier/train.py .

# Создание директории для моделей
RUN mkdir -p /app/models

# Копирование моделей из корневой директории
COPY models/classifier.keras /app/models/classifier.keras
COPY models/scaler.pkl /app/models/scaler.pkl

EXPOSE 8001

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001"]