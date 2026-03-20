// Типы для фитнес-платформы

export type UserRole = 'ADMIN' | 'CLUB_ADMIN' | 'TRAINER' | 'CLIENT'

export type FitnessGoal = 'weight_loss' | 'muscle_gain' | 'endurance' | 'rehabilitation' | 'maintenance'

export type ActivityLevel = 'sedentary' | 'light' | 'moderate' | 'active' | 'very_active'

export type FitnessClass = 'excellent' | 'good' | 'moderate' | 'needs_attention' | 'at_risk'

export type RiskLevel = 'low' | 'moderate' | 'high' | 'very_high'

// Биометрические данные
export interface BiometricReading {
  heartRate?: number
  heartRateVariability?: number
  restingHeartRate?: number
  bloodPressureSystolic?: number
  bloodPressureDiastolic?: number
  spO2?: number
  respiratoryRate?: number
  bodyTemperature?: number
  steps?: number
  distance?: number
  caloriesBurned?: number
  activeMinutes?: number
  sleepDuration?: number
  sleepQuality?: number
  stressLevel?: number
}

// Классификация состояния пользователя
export interface UserClassification {
  fitnessClass: FitnessClass
  confidence: number
  cardiovascularRisk: RiskLevel
  metabolicRisk: RiskLevel
  injuryRisk: RiskLevel
  overtrainingRisk: RiskLevel
  recommendations: string[]
  insights: string[]
}

// Входные данные для классификации (6 параметров для нейросети)
export interface ClassificationInput {
  avgHeartRate: number        // Средний пульс
  restingHeartRate: number    // Пульс в покое
  sleepQuality: number        // Качество сна (0-100)
  activityLevel: number       // Уровень активности (минуты/день)
  stressLevel: number         // Уровень стресса (0-100)
  recoveryScore: number       // Оценка восстановления (HRV-based)
}

// Программа тренировок
export interface TrainingProgramData {
  name: string
  description: string
  goal: FitnessGoal
  difficulty: 'beginner' | 'intermediate' | 'advanced'
  durationWeeks: number
  exercises: ProgramExerciseData[]
}

export interface ProgramExerciseData {
  name: string
  category: 'strength' | 'cardio' | 'flexibility' | 'balance'
  muscleGroups: string[]
  sets?: number
  reps?: number
  duration?: number
  restTime?: number
  weekNumber: number
  dayOfWeek: number
  order: number
}

// Данные пользователя для генерации программы
export interface UserProfileForGeneration {
  age: number
  gender: string
  weight: number
  height: number
  fitnessGoal: FitnessGoal
  activityLevel: ActivityLevel
  contraindications: string[]
  chronicDiseases: string[]
  availableEquipment: string[]
  trainingFrequency: number // дней в неделю
  sessionDuration: number   // минут за тренировку
}

// Результат генерации программы
export interface GeneratedProgram {
  programName: string
  description: string
  weeklySchedule: WeeklySchedule[]
  estimatedResults: string
  safetyNotes: string[]
  progressionPlan: string
}

export interface WeeklySchedule {
  weekNumber: number
  focus: string
  days: DailyWorkout[]
}

export interface DailyWorkout {
  dayOfWeek: number
  type: string
  exercises: ExerciseDetail[]
  totalDuration: number
  estimatedCalories: number
}

export interface ExerciseDetail {
  name: string
  sets: number
  reps: string
  rest: number
  notes: string
}

// Достижения
export interface AchievementData {
  id: string
  name: string
  description: string
  category: 'fitness' | 'consistency' | 'social' | 'special'
  icon: string
  points: number
  level: number
  progress: number
  unlockedAt?: Date
}

// Данные для дашборда
export interface DashboardData {
  user: {
    name: string
    role: UserRole
    fitnessGoal: FitnessGoal | null
  }
  todayStats: {
    steps: number
    calories: number
    activeMinutes: number
    heartRate: number
    sleepHours: number
  }
  weeklyProgress: {
    date: string
    steps: number
    calories: number
    workouts: number
  }[]
  currentProgram: {
    name: string
    progress: number
    daysRemaining: number
  } | null
  recentAchievements: AchievementData[]
  healthAlerts: string[]
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Симуляция данных с носимых устройств
export interface WearableDevice {
  type: 'apple_watch' | 'samsung_health' | 'huawei_health' | 'amazfit' | 'garmin'
  name: string
  lastSync: Date
  batteryLevel?: number
  firmwareVersion?: string
}

export interface SyncResult {
  success: boolean
  recordsSynced: number
  lastSyncTime: Date
  errors?: string[]
}
