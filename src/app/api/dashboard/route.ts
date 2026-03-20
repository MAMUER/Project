import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// Получение данных для дашборда
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const userId = searchParams.get('userId') || 'demo-user';

    // Получаем данные пользователя
    const user = await db.user.findUnique({
      where: { id: userId },
      include: {
        achievements: {
          include: { achievement: true },
          orderBy: { unlockedAt: 'desc' },
          take: 5
        },
        trainingPrograms: {
          where: { status: 'active' },
          take: 1
        }
      }
    });

    if (!user) {
      // Создаем демо-пользователя если не существует
      return NextResponse.json({
        success: true,
        data: await createDemoDashboard(userId)
      });
    }

    // Получаем биометрические данные за сегодня
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const todayBiometrics = await db.biometricData.findMany({
      where: {
        userId,
        timestamp: { gte: today }
      }
    });

    // Получаем данные за неделю для графика
    const weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);

    const weeklyBiometrics = await db.biometricData.findMany({
      where: {
        userId,
        timestamp: { gte: weekAgo }
      },
      orderBy: { timestamp: 'asc' }
    });

    // Агрегируем данные
    const todayStats = aggregateTodayStats(todayBiometrics);
    const weeklyProgress = aggregateWeeklyProgress(weeklyBiometrics);
    const healthAlerts = generateHealthAlerts(todayBiometrics, user);

    // Текущая программа
    const currentProgram = user.trainingPrograms[0] ? {
      name: user.trainingPrograms[0].name,
      progress: user.trainingPrograms[0].progress,
      daysRemaining: Math.ceil(
        (new Date(user.trainingPrograms[0].endDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24)
      )
    } : null;

    // Достижения
    const recentAchievements = user.achievements.map(ua => ({
      id: ua.achievement.id,
      name: ua.achievement.name,
      description: ua.achievement.description,
      category: ua.achievement.category,
      icon: ua.achievement.icon || '🏆',
      points: ua.achievement.points,
      level: ua.level,
      progress: ua.progress,
      unlockedAt: ua.unlockedAt
    }));

    const dashboardData = {
      user: {
        name: user.name || 'Пользователь',
        role: user.role,
        fitnessGoal: user.fitnessGoal as any
      },
      todayStats,
      weeklyProgress,
      currentProgram,
      recentAchievements,
      healthAlerts
    };

    return NextResponse.json({
      success: true,
      data: dashboardData
    });

  } catch (error) {
    console.error('Dashboard error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to fetch dashboard data' },
      { status: 500 }
    );
  }
}

function aggregateTodayStats(records: any[]) {
  if (records.length === 0) {
    return {
      steps: 0,
      calories: 0,
      activeMinutes: 0,
      heartRate: 0,
      sleepHours: 0
    };
  }

  const stepsRecords = records.filter(r => r.steps);
  const caloriesRecords = records.filter(r => r.caloriesBurned);
  const hrRecords = records.filter(r => r.heartRate);
  const sleepRecords = records.filter(r => r.sleepDuration);
  const activeRecords = records.filter(r => r.activeMinutes);

  return {
    steps: stepsRecords.length > 0 
      ? stepsRecords.reduce((sum, r) => sum + (r.steps || 0), 0) 
      : 0,
    calories: caloriesRecords.length > 0 
      ? Math.round(caloriesRecords.reduce((sum, r) => sum + (r.caloriesBurned || 0), 0)) 
      : 0,
    activeMinutes: activeRecords.length > 0 
      ? activeRecords.reduce((sum, r) => sum + (r.activeMinutes || 0), 0) 
      : 0,
    heartRate: hrRecords.length > 0 
      ? Math.round(hrRecords.reduce((sum, r) => sum + r.heartRate, 0) / hrRecords.length) 
      : 0,
    sleepHours: sleepRecords.length > 0 
      ? Math.round((sleepRecords[0].sleepDuration || 0) / 60 * 10) / 10 
      : 0
  };
}

function aggregateWeeklyProgress(records: any[]) {
  const dayMap = new Map<string, { steps: number; calories: number; workouts: number }>();

  records.forEach(record => {
    const date = record.timestamp.toISOString().split('T')[0];
    const existing = dayMap.get(date) || { steps: 0, calories: 0, workouts: 0 };
    
    existing.steps += record.steps || 0;
    existing.calories += record.caloriesBurned || 0;
    if (record.workoutType) existing.workouts += 1;
    
    dayMap.set(date, existing);
  });

  return Array.from(dayMap.entries())
    .map(([date, data]) => ({
      date,
      steps: data.steps,
      calories: data.calories,
      workouts: data.workouts
    }))
    .sort((a, b) => a.date.localeCompare(b.date));
}

function generateHealthAlerts(records: any[], user: any): string[] {
  const alerts: string[] = [];

  // Проверяем пульс
  const hrRecords = records.filter(r => r.heartRate);
  if (hrRecords.length > 0) {
    const avgHR = hrRecords.reduce((sum, r) => sum + r.heartRate, 0) / hrRecords.length;
    if (avgHR > 100) {
      alerts.push('⚠️ Повышенный средний пульс в течение дня. Рекомендуется проконсультироваться с врачом.');
    }
  }

  // Проверяем сон
  const sleepRecords = records.filter(r => r.sleepDuration);
  if (sleepRecords.length > 0) {
    const lastSleep = sleepRecords[0];
    if (lastSleep.sleepDuration < 360) { // Менее 6 часов
      alerts.push('😴 Недостаточно сна прошлой ночью. Постарайтесь лечь раньше сегодня.');
    }
    if (lastSleep.sleepQuality < 60) {
      alerts.push('💤 Низкое качество сна. Попробуйте исключить экраны за час до сна.');
    }
  }

  // Проверяем активность
  const stepsRecords = records.filter(r => r.steps);
  if (stepsRecords.length > 0) {
    const totalSteps = stepsRecords.reduce((sum, r) => sum + (r.steps || 0), 0);
    if (totalSteps < 3000) {
      alerts.push('🚶 Низкая активность сегодня. Попробуйте прогуляться 15-20 минут.');
    }
  }

  // Проверяем SpO2
  const spo2Records = records.filter(r => r.spO2);
  if (spo2Records.length > 0) {
    const lowSpO2 = spo2Records.filter(r => r.spO2 < 95);
    if (lowSpO2.length > 0) {
      alerts.push('🫁 Зафиксировано низкое насыщение кислородом. Обратите внимание.');
    }
  }

  return alerts;
}

async function createDemoDashboard(userId: string) {
  // Демо данные для нового пользователя
  const now = new Date();
  const weeklyProgress = [];
  
  for (let i = 6; i >= 0; i--) {
    const date = new Date(now);
    date.setDate(date.getDate() - i);
    weeklyProgress.push({
      date: date.toISOString().split('T')[0],
      steps: Math.round(5000 + Math.random() * 5000),
      calories: Math.round(1500 + Math.random() * 1000),
      workouts: Math.random() > 0.6 ? 1 : 0
    });
  }

  return {
    user: {
      name: 'Демо Пользователь',
      role: 'CLIENT',
      fitnessGoal: 'maintenance'
    },
    todayStats: {
      steps: 6234,
      calories: 1842,
      activeMinutes: 32,
      heartRate: 72,
      sleepHours: 7.5
    },
    weeklyProgress,
    currentProgram: {
      name: 'Программа поддержания формы',
      progress: 45,
      daysRemaining: 21
    },
    recentAchievements: [
      {
        id: '1',
        name: 'Первые шаги',
        description: 'Начните отслеживать активность',
        category: 'fitness',
        icon: '🎯',
        points: 10,
        level: 1,
        progress: 100,
        unlockedAt: new Date()
      },
      {
        id: '2',
        name: 'Неделя активности',
        description: '7 дней подряд с активностью',
        category: 'consistency',
        icon: '🔥',
        points: 50,
        level: 1,
        progress: 71,
        unlockedAt: null
      }
    ],
    healthAlerts: [
      '💡 Не забудьте выпить воды — вы пьете недостаточно жидкости'
    ]
  };
}
