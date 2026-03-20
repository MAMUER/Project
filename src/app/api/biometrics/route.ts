import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// Получение биометрических данных пользователя
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const userId = searchParams.get('userId') || 'demo-user';
    const period = searchParams.get('period') || 'week'; // day, week, month
    const source = searchParams.get('source'); // apple_watch, samsung_health, etc.

    // Вычисляем дату начала периода
    const now = new Date();
    let startDate = new Date();
    
    switch (period) {
      case 'day':
        startDate.setHours(now.getHours() - 24);
        break;
      case 'week':
        startDate.setDate(now.getDate() - 7);
        break;
      case 'month':
        startDate.setMonth(now.getMonth() - 1);
        break;
      default:
        startDate.setDate(now.getDate() - 7);
    }

    // Получаем данные из БД
    const biometricData = await db.biometricData.findMany({
      where: {
        userId: userId,
        timestamp: {
          gte: startDate
        },
        ...(source && { source })
      },
      orderBy: {
        timestamp: 'desc'
      },
      take: 1000 // Ограничиваем количество записей
    });

    // Агрегируем данные для статистики
    const stats = aggregateStats(biometricData);

    return NextResponse.json({
      success: true,
      data: {
        records: biometricData,
        stats,
        period,
        source: source || 'all'
      }
    });

  } catch (error) {
    console.error('Error fetching biometrics:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to fetch biometric data' },
      { status: 500 }
    );
  }
}

// Добавление новых биометрических данных
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { userId, source, data } = body;

    if (!userId || !data) {
      return NextResponse.json(
        { success: false, error: 'userId and data are required' },
        { status: 400 }
      );
    }

    // Создаем запись биометрических данных
    const biometricRecord = await db.biometricData.create({
      data: {
        userId,
        source: source || 'manual',
        timestamp: data.timestamp ? new Date(data.timestamp) : new Date(),
        
        // Сердечно-сосудистая система
        heartRate: data.heartRate,
        heartRateVariability: data.heartRateVariability,
        restingHeartRate: data.restingHeartRate,
        bloodPressureSystolic: data.bloodPressureSystolic,
        bloodPressureDiastolic: data.bloodPressureDiastolic,
        
        // Дыхание
        spO2: data.spO2,
        respiratoryRate: data.respiratoryRate,
        
        // Температура
        bodyTemperature: data.bodyTemperature,
        
        // Активность
        steps: data.steps,
        distance: data.distance,
        caloriesBurned: data.caloriesBurned,
        activeMinutes: data.activeMinutes,
        floorsClimbed: data.floorsClimbed,
        
        // Сон
        sleepDuration: data.sleepDuration,
        sleepQuality: data.sleepQuality,
        deepSleepDuration: data.deepSleepDuration,
        remSleepDuration: data.remSleepDuration,
        sleepStartTime: data.sleepStartTime ? new Date(data.sleepStartTime) : null,
        sleepEndTime: data.sleepEndTime ? new Date(data.sleepEndTime) : null,
        
        // Стресс
        stressLevel: data.stressLevel,
        
        // Тренировка
        workoutType: data.workoutType,
        workoutDuration: data.workoutDuration,
        workoutIntensity: data.workoutIntensity,
        
        // Сырые данные
        rawData: data.rawData ? JSON.stringify(data.rawData) : null
      }
    });

    return NextResponse.json({
      success: true,
      data: biometricRecord
    });

  } catch (error) {
    console.error('Error saving biometrics:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to save biometric data' },
      { status: 500 }
    );
  }
}

// Агрегация статистики
function aggregateStats(records: any[]) {
  if (records.length === 0) {
    return {
      avgHeartRate: null,
      avgSteps: null,
      avgSleepDuration: null,
      avgCaloriesBurned: null,
      totalWorkouts: 0
    };
  }

  const heartRates = records.filter(r => r.heartRate).map(r => r.heartRate);
  const steps = records.filter(r => r.steps).map(r => r.steps);
  const sleepDurations = records.filter(r => r.sleepDuration).map(r => r.sleepDuration);
  const calories = records.filter(r => r.caloriesBurned).map(r => r.caloriesBurned);
  const workouts = records.filter(r => r.workoutType);

  return {
    avgHeartRate: heartRates.length > 0 
      ? Math.round(heartRates.reduce((a, b) => a + b, 0) / heartRates.length) 
      : null,
    avgSteps: steps.length > 0 
      ? Math.round(steps.reduce((a, b) => a + b, 0) / steps.length) 
      : null,
    avgSleepDuration: sleepDurations.length > 0 
      ? Math.round(sleepDurations.reduce((a, b) => a + b, 0) / sleepDurations.length) 
      : null,
    avgCaloriesBurned: calories.length > 0 
      ? Math.round(calories.reduce((a, b) => a + b, 0) / calories.length) 
      : null,
    totalWorkouts: workouts.length
  };
}
