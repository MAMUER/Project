FROM python:3.11-slim

WORKDIR /app

# Установка зависимостей
COPY cmd/ml-classifier/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода
COPY cmd/ml-classifier/main.py .

# Создание директории для моделей
RUN mkdir -p /app/models

# Копирование моделей (из папки models внутри ml-classifier)
COPY cmd/ml-classifier/models/ /app/models/

EXPOSE 8001

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001"]