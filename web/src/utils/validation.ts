/**
 * Валидация email
 */
export const isValidEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@([^\s@]+\.)+[^\s@]+$/;
    return emailRegex.test(email);
};

/**
 * Валидация пароля (минимум 8 символов, буквы и цифры)
 */
export const isValidPassword = (password: string): boolean => {
    return password.length >= 8 && /[A-Za-z]/.test(password) && /[0-9]/.test(password);
};

/**
 * Валидация пульса
 */
export const isValidHeartRate = (value: number): boolean => {
    return value >= 30 && value <= 220;
};

/**
 * Валидация давления
 */
export const isValidBloodPressure = (systolic: number, diastolic: number): boolean => {
    return systolic >= 70 && systolic <= 200 && diastolic >= 40 && diastolic <= 120;
};

/**
 * Валидация SpO2
 */
export const isValidSpO2 = (value: number): boolean => {
    return value >= 70 && value <= 100;
};

/**
 * Валидация температуры
 */
export const isValidTemperature = (value: number): boolean => {
    return value >= 35 && value <= 42;
};

/**
 * Валидация длительности сна (минуты)
 */
export const isValidSleepDuration = (minutes: number): boolean => {
    return minutes >= 0 && minutes <= 1440; // до 24 часов
};

/**
 * Валидация формы биометрических данных
 */
export interface ValidationResult {
    valid: boolean;
    errors: Record<string, string>;
}

export const validateBiometricData = (data: {
    heart_rate: number;
    systolic: number;
    diastolic: number;
    spo2: number;
    temperature: number;
}): ValidationResult => {
    const errors: Record<string, string> = {};
    
    if (!isValidHeartRate(data.heart_rate)) {
        errors.heart_rate = "Пульс должен быть в диапазоне 30-220 уд/мин";
    }
    
    if (!isValidBloodPressure(data.systolic, data.diastolic)) {
        errors.blood_pressure = "Давление должно быть: сист. 70-200, диаст. 40-120";
    }
    
    if (!isValidSpO2(data.spo2)) {
        errors.spo2 = "SpO2 должен быть в диапазоне 70-100%";
    }
    
    if (!isValidTemperature(data.temperature)) {
        errors.temperature = "Температура должна быть в диапазоне 35-42°C";
    }
    
    return {
        valid: Object.keys(errors).length === 0,
        errors
    };
};

/**
 * Валидация формы регистрации
 */
export const validateRegistration = (data: {
    email: string;
    password: string;
    confirmPassword: string;
}): ValidationResult => {
    const errors: Record<string, string> = {};
    
    if (!data.email) {
        errors.email = "Email обязателен";
    } else if (!isValidEmail(data.email)) {
        errors.email = "Введите корректный email";
    }
    
    if (!data.password) {
        errors.password = "Пароль обязателен";
    } else if (!isValidPassword(data.password)) {
        errors.password = "Пароль должен содержать минимум 8 символов, буквы и цифры";
    }
    
    if (data.password !== data.confirmPassword) {
        errors.confirmPassword = "Пароли не совпадают";
    }
    
    return {
        valid: Object.keys(errors).length === 0,
        errors
    };
};