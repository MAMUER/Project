FROM python:3.11-slim

WORKDIR /app

COPY cmd/ml-generator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY cmd/ml-generator/main.py .

RUN mkdir -p /app/models
COPY models/ /app/models/

EXPOSE 8002

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8002"]