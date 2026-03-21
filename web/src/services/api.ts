import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    token: string;
    user_id: string;
    role: string;
    expires_in: number;
}

export interface BiometricData {
    user_id: string;
    device_type: string;
    heart_rate: number;
    ecg: string;
    blood_pressure: { systolic: number; diastolic: number };
    spo2: number;
    temperature: number;
    sleep: { duration: number; deep_sleep: number };
    timestamp: string;
}

export interface GenerateProgramRequest {
    training_class: string;
    contraindications: string[];
    goals: string[];
    fitness_level: string;
    age_group: string;
    gender: string;
    has_injury: boolean;
    duration_weeks: number;
}

export const auth = {
    login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
    verify: () => api.get('/auth/verify'),
};

export const biometric = {
    submit: (data: BiometricData) => api.post('/biometric', data),
    getHistory: (userId: string, from?: string, to?: string) =>
        api.get(`/biometric/${userId}`, { params: { from, to } }),
};

export const training = {
    generate: (data: GenerateProgramRequest) => api.post('/training/generate', data),
    getPrograms: (userId: string) => api.get(`/training/programs/${userId}`),
};

export default api;