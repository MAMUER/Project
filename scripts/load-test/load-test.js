import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';

// Пользовательские метрики
const loginDuration = new Trend('login_duration');
const getProfileDuration = new Trend('get_profile_duration');
const biometricDuration = new Trend('biometric_duration');
const generatePlanDuration = new Trend('generate_plan_duration');

export const options = {
    stages: [
        { duration: '30s', target: 20 },  // разогрев до 20 пользователей
        { duration: '1m', target: 50 },   // пиковая нагрузка
        { duration: '30s', target: 0 },   // спад
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% запросов <500мс
        'login_duration': ['p(95)<300'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'https://localhost:8443';
const TEST_USER = {
    email: 'loadtest@example.com',
    password: 'LoadTest123!',
    full_name: 'Load Test User',
    role: 'client'
};

export function setup() {
    // Регистрация тестового пользователя
    const registerRes = http.post(`${BASE_URL}/api/v1/register`, JSON.stringify(TEST_USER), {
        headers: { 'Content-Type': 'application/json' },
    });
    check(registerRes, { 'register success': (r) => r.status === 200 });
    
    // Логин
    const loginRes = http.post(`${BASE_URL}/api/v1/login`, JSON.stringify({
        email: TEST_USER.email,
        password: TEST_USER.password,
    }), { headers: { 'Content-Type': 'application/json' } });
    check(loginRes, { 'login success': (r) => r.status === 200 });
    
    const token = loginRes.json('access_token');
    return { token };
}

export default function (data) {
    const headers = {
        'Authorization': `Bearer ${data.token}`,
        'Content-Type': 'application/json',
    };
    
    // 1. Получение профиля
    let start = new Date();
    let profileRes = http.get(`${BASE_URL}/api/v1/profile`, { headers });
    check(profileRes, { 'profile status 200': (r) => r.status === 200 });
    getProfileDuration.add(new Date() - start);
    sleep(0.5);
    
    // 2. Добавление биометрических данных
    start = new Date();
    const biometric = {
        metric_type: 'heart_rate',
        value: 70 + Math.random() * 30,
        timestamp: new Date().toISOString(),
        device_type: 'test_device'
    };
    let bioRes = http.post(`${BASE_URL}/api/v1/biometrics`, JSON.stringify(biometric), { headers });
    check(bioRes, { 'biometrics status 201': (r) => r.status === 201 });
    biometricDuration.add(new Date() - start);
    sleep(0.5);
    
    // 3. Генерация программы тренировок
    start = new Date();
    const plan = {
        class_name: 'endurance',
        confidence: 0.85,
        duration_weeks: 4,
        available_days: [1, 3, 5]
    };
    let planRes = http.post(`${BASE_URL}/api/v1/ml/generate-plan`, JSON.stringify(plan), { headers });
    check(planRes, { 'generate plan status 200': (r) => r.status === 200 });
    generatePlanDuration.add(new Date() - start);
    sleep(1);
}