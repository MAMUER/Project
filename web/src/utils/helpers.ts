/**
 * Форматирование даты в локальный формат
 */
export const formatDate = (date: Date | string): string => {
    const d = new Date(date);
    return d.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
    });
};

/**
 * Форматирование времени
 */
export const formatTime = (date: Date | string): string => {
    const d = new Date(date);
    return d.toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit'
    });
};

/**
 * Форматирование даты и времени
 */
export const formatDateTime = (date: Date | string): string => {
    return `${formatDate(date)} ${formatTime(date)}`;
};

/**
 * Расчет процента выполнения
 */
export const calculateProgress = (current: number, total: number): number => {
    if (total === 0) return 0;
    return Math.round((current / total) * 100);
};

/**
 * Нормализация пульса (30-220)
 */
export const normalizeHeartRate = (value: number): number => {
    return Math.min(220, Math.max(30, value));
};

/**
 * Нормализация SpO2 (70-100)
 */
export const normalizeSpO2 = (value: number): number => {
    return Math.min(100, Math.max(70, value));
};

/**
 * Нормализация температуры (35-42)
 */
export const normalizeTemperature = (value: number): number => {
    return Math.min(42, Math.max(35, value));
};

/**
 * Получение статуса здоровья на основе показателей
 */
export const getHealthStatus = (heartRate: number, spo2: number, temperature: number): string => {
    if (heartRate < 60 || heartRate > 100) return "attention";
    if (spo2 < 95) return "warning";
    if (temperature < 36 || temperature > 37.5) return "warning";
    return "normal";
};

/**
 * Получение текстового описания статуса
 */
export const getHealthStatusText = (status: string): string => {
    const statusMap: Record<string, string> = {
        normal: "В норме",
        warning: "Требует внимания",
        attention: "Обратитесь к врачу"
    };
    return statusMap[status] || "Неизвестно";
};

/**
 * Кэширование результатов
 */
export const cacheResult = <T>(key: string, data: T, ttlMinutes: number = 5): void => {
    const item = {
        data,
        expires: Date.now() + ttlMinutes * 60 * 1000
    };
    localStorage.setItem(key, JSON.stringify(item));
};

/**
 * Получение кэшированных данных
 */
export const getCachedResult = <T>(key: string): T | null => {
    const item = localStorage.getItem(key);
    if (!item) return null;
    
    const parsed = JSON.parse(item);
    if (Date.now() > parsed.expires) {
        localStorage.removeItem(key);
        return null;
    }
    
    return parsed.data as T;
};

/**
 * Дебаунс для оптимизации ввода
 */
export const debounce = <T extends (...args: any[]) => any>(
    func: T,
    delay: number
): ((...args: Parameters<T>) => void) => {
    let timeout: NodeJS.Timeout;
    return (...args: Parameters<T>) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => func(...args), delay);
    };
};

/**
 * Форматирование длительности в минутах
 */
export const formatDuration = (minutes: number): string => {
    if (minutes < 60) return `${minutes} мин`;
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return mins > 0 ? `${hours} ч ${mins} мин` : `${hours} ч`;
};

/**
 * Получение иконки для типа тренировки
 */
export const getWorkoutIcon = (type: string): string => {
    const icons: Record<string, string> = {
        cardio: "🏃",
        strength: "💪",
        flexibility: "🧘",
        recovery: "😌",
        hiit: "⚡",
        endurance: "🚴",
        rest: "😴"
    };
    return icons[type] || "🏋️";
};

/**
 * Получение цвета для интенсивности
 */
export const getIntensityColor = (intensity: string): string => {
    const colors: Record<string, string> = {
        light: "#4caf50",
        moderate: "#ff9800",
        high: "#f44336",
        none: "#9e9e9e"
    };
    return colors[intensity] || "#757575";
};