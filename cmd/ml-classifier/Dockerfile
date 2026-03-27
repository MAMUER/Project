FROM python:3.11-slim

WORKDIR /app

# Установка зависимостей
COPY cmd/ml-classifier/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода и моделей
COPY cmd/ml-classifier/main.py .
COPY cmd/ml-classifier/classifier.keras /app/models/classifier.keras
COPY cmd/ml-classifier/scaler.pkl /app/models/scaler.pkl

EXPOSE 8001

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001"]