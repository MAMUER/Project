#!/bin/bash
BASE_URL=${1:-http://localhost:8080}
echo "Testing API at $BASE_URL"

# Health check
curl -s -o /dev/null -w "Health: %{http_code}\n" $BASE_URL/health

# Регистрация
REGISTER_RESP=$(curl -s -X POST $BASE_URL/api/v1/register \
    -H "Content-Type: application/json" \
    -d '{"email":"apitest@example.com","password":"test123","full_name":"API Test","role":"client"}')
echo "Register: $REGISTER_RESP"

# Логин
LOGIN_RESP=$(curl -s -X POST $BASE_URL/api/v1/login \
    -H "Content-Type: application/json" \
    -d '{"email":"apitest@example.com","password":"test123"}')
TOKEN=$(echo $LOGIN_RESP | jq -r '.access_token')
echo "Login: $TOKEN"

# Получение профиля
curl -s -X GET $BASE_URL/api/v1/profile \
    -H "Authorization: Bearer $TOKEN" | jq .

# Добавление биометрии
curl -s -X POST $BASE_URL/api/v1/biometrics \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"metric_type":"heart_rate","value":75,"timestamp":"2026-03-25T10:00:00Z","device_type":"test"}' | jq .

# Классификация
curl -s -X GET $BASE_URL/api/v1/ml/classify \
    -H "Authorization: Bearer $TOKEN" | jq .

# Генерация программы
curl -s -X POST $BASE_URL/api/v1/ml/generate-plan \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"class_name":"endurance","confidence":0.85,"duration_weeks":4,"available_days":[1,3,5]}' | jq .

echo "Tests completed."