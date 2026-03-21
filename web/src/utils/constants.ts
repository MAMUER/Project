/**
 * Типы тренировок
 */
export const WORKOUT_TYPES = {
    CARDIO: 'cardio',
    STRENGTH: 'strength',
    FLEXIBILITY: 'flexibility',
    RECOVERY: 'recovery',
    HIIT: 'hiit',
    ENDURANCE: 'endurance',
    REST: 'rest'
} as const;

export const WORKOUT_TYPE_LABELS: Record<string, string> = {
    [WORKOUT_TYPES.CARDIO]: 'Кардио',
    [WORKOUT_TYPES.STRENGTH]: 'Силовая',
    [WORKOUT_TYPES.FLEXIBILITY]: 'Гибкость',
    [WORKOUT_TYPES.RECOVERY]: 'Восстановление',
    [WORKOUT_TYPES.HIIT]: 'HIIT',
    [WORKOUT_TYPES.ENDURANCE]: 'Выносливость',
    [WORKOUT_TYPES.REST]: 'Отдых'
};

/**
 * Уровни интенсивности
 */
export const INTENSITY_LEVELS = {
    LIGHT: 'light',
    MODERATE: 'moderate',
    HIGH: 'high',
    NONE: 'none'
} as const;

export const INTENSITY_LABELS: Record<string, string> = {
    [INTENSITY_LEVELS.LIGHT]: 'Низкая',
    [INTENSITY_LEVELS.MODERATE]: 'Средняя',
    [INTENSITY_LEVELS.HIGH]: 'Высокая',
    [INTENSITY_LEVELS.NONE]: 'Отсутствует'
};

/**
 * Роли пользователей
 */
export const USER_ROLES = {
    USER: 'user',
    TRAINER: 'trainer',
    ADMIN: 'admin'
} as const;

export const USER_ROLE_LABELS: Record<string, string> = {
    [USER_ROLES.USER]: 'Пользователь',
    [USER_ROLES.TRAINER]: 'Тренер',
    [USER_ROLES.ADMIN]: 'Администратор'
};

/**
 * Уровни подготовки
 */
export const FITNESS_LEVELS = {
    BEGINNER: 'beginner',
    INTERMEDIATE: 'intermediate',
    ADVANCED: 'advanced'
} as const;

export const FITNESS_LEVEL_LABELS: Record<string, string> = {
    [FITNESS_LEVELS.BEGINNER]: 'Начинающий',
    [FITNESS_LEVELS.INTERMEDIATE]: 'Средний',
    [FITNESS_LEVELS.ADVANCED]: 'Продвинутый'
};

/**
 * Возрастные группы
 */
export const AGE_GROUPS = {
    YOUNG: 'young',
    ADULT: 'adult',
    SENIOR: 'senior',
    ELDERLY: 'elderly'
} as const;

export const AGE_GROUP_LABELS: Record<string, string> = {
    [AGE_GROUPS.YOUNG]: '18-35 лет',
    [AGE_GROUPS.ADULT]: '36-50 лет',
    [AGE_GROUPS.SENIOR]: '51-65 лет',
    [AGE_GROUPS.ELDERLY]: '65+ лет'
};

/**
 * Цели тренировок
 */
export const GOALS = [
    { value: 'weight_loss', label: 'Снижение веса' },
    { value: 'muscle_gain', label: 'Набор мышечной массы' },
    { value: 'endurance', label: 'Развитие выносливости' },
    { value: 'flexibility', label: 'Развитие гибкости' },
    { value: 'rehabilitation', label: 'Реабилитация' }
];

/**
 * Противопоказания
 */
export const CONTRAINDICATIONS = [
    { value: 'heart_issues', label: 'Заболевания сердца' },
    { value: 'joint_problems', label: 'Проблемы с суставами' },
    { value: 'hypertension', label: 'Гипертония' },
    { value: 'asthma', label: 'Астма' },
    { value: 'diabetes', label: 'Диабет' }
];

/**
 * API endpoints
 */
export const API_ENDPOINTS = {
    AUTH: {
        LOGIN: '/auth/login',
        REGISTER: '/auth/register',
        VERIFY: '/auth/verify',
        REFRESH: '/auth/refresh'
    },
    BIOMETRIC: {
        SUBMIT: '/biometric',
        HISTORY: (userId: string) => `/biometric/${userId}`,
        STATS: (userId: string) => `/biometric/${userId}/stats`
    },
    TRAINING: {
        GENERATE: '/training/generate',
        PROGRAMS: (userId: string) => `/training/programs/${userId}`,
        PROGRAM: (programId: string) => `/training/programs/${programId}`
    }
} as const;

/**
 * HTTP статусы
 */
export const HTTP_STATUS = {
    OK: 200,
    CREATED: 201,
    BAD_REQUEST: 400,
    UNAUTHORIZED: 401,
    FORBIDDEN: 403,
    NOT_FOUND: 404,
    INTERNAL_ERROR: 500
} as const;