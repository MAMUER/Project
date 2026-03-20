import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// Симуляция синхронизации данных с носимых устройств
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { userId, deviceType } = body;

    if (!userId || !deviceType) {
      return NextResponse.json(
        { success: false, error: 'userId and deviceType are required' },
        { status: 400 }
      );
    }

    // Симулируем синхронизацию данных
    const simulatedData = generateSimulatedData(deviceType);
    
    // Сохраняем данные в БД
    const savedRecords = [];
    for (const dataPoint of simulatedData) {
      const record = await db.biometricData.create({
        data: {
          userId,
          source: deviceType,
          timestamp: dataPoint.timestamp,
          ...dataPoint.metrics
        }
      });
      savedRecords.push(record);
    }

    return NextResponse.json({
      success: true,
      data: {
        deviceType,
        recordsSynced: savedRecords.length,
        lastSyncTime: new Date().toISOString(),
        sampleData: savedRecords.slice(0, 3) // Первые 3 записи для preview
      }
    });

  } catch (error) {
    console.error('Sync error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to sync device data' },
      { status: 500 }
    );
  }
}

// Получение списка подключенных устройств
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const userId = searchParams.get('userId') || 'demo-user';

    // Получаем последние синхронизации по каждому источнику
    const lastSyncs = await db.biometricData.groupBy({
      by: ['source'],
      where: { userId },
      _max: {
        timestamp: true
      },
      _count: {
        id: true
      }
    });

    const devices = lastSyncs.map(sync => ({
      type: sync.source,
      name: getDeviceName(sync.source),
      lastSync: sync._max.timestamp,
      recordsCount: sync._count.id
    }));

    return NextResponse.json({
      success: true,
      data: devices
    });

  } catch (error) {
    console.error('Error fetching devices:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to fetch connected devices' },
      { status: 500 }
    );
  }
}

function getDeviceName(source: string): string {
  const names: Record<string, string> = {
    apple_watch: 'Apple Watch',
    samsung_health: 'Samsung Health',
    huawei_health: 'Huawei Health',
    amazfit: 'Amazfit T-Rex 3',
    garmin: 'Garmin Connect',
    manual: 'Ручной ввод'
  };
  return names[source] || source;
}

function generateSimulatedData(deviceType: string) {
  const now = new Date();
  const data = [];

  // Генерируем данные за последние 24 часа с интервалом в 1 час
  for (let i = 23; i >= 0; i--) {
    const timestamp = new Date(now.getTime() - i * 60 * 60 * 1000);
    const isNight = timestamp.getHours() >= 22 || timestamp.getHours() <= 6;
    const isMorning = timestamp.getHours() >= 7 && timestamp.getHours() <= 9;
    const isAfternoon = timestamp.getHours() >= 12 && timestamp.getHours() <= 14;
    const isEvening = timestamp.getHours() >= 18 && timestamp.getHours() <= 20;

    // Базовый пульс с вариациями
    let baseHeartRate = 72;
    if (isNight) baseHeartRate = 58 + Math.random() * 5;
    else if (isMorning) baseHeartRate = 75 + Math.random() * 10;
    else if (isAfternoon) baseHeartRate = 70 + Math.random() * 8;
    else if (isEvening) baseHeartRate = 78 + Math.random() * 15; // Возможная тренировка
    else baseHeartRate = 72 + Math.random() * 10;

    // SpO2 обычно стабилен
    const spO2 = 96 + Math.random() * 3;

    // Шаги накапливаются в течение дня
    let steps = 0;
    if (!isNight) {
      const dayProgress = (timestamp.getHours() - 7) / 14; // Прогресс дня
      steps = Math.round((dayProgress * 8000 + Math.random() * 500) / 24);
    }

    // Калории
    const caloriesBurned = Math.round(80 + Math.random() * 40);

    data.push({
      timestamp,
      metrics: {
        heartRate: Math.round(baseHeartRate),
        heartRateVariability: Math.round(35 + Math.random() * 20),
        restingHeartRate: isMorning ? Math.round(58 + Math.random() * 5) : null,
        spO2: Math.round(spO2),
        steps,
        caloriesBurned,
        activeMinutes: isEvening && Math.random() > 0.5 ? 45 : null,
        stressLevel: Math.round(20 + Math.random() * 30),
        bodyTemperature: 36.4 + Math.random() * 0.4
      }
    });
  }

  // Добавляем данные о сне (последняя ночь)
  const sleepStart = new Date(now);
  sleepStart.setHours(23, 0, 0, 0);
  sleepStart.setDate(sleepStart.getDate() - 1);

  const sleepEnd = new Date(sleepStart);
  sleepEnd.setHours(7, 30, 0, 0);

  data.push({
    timestamp: sleepStart,
    metrics: {
      sleepDuration: 510, // 8.5 часов
      sleepQuality: Math.round(70 + Math.random() * 20),
      deepSleepDuration: Math.round(80 + Math.random() * 40),
      remSleepDuration: Math.round(90 + Math.random() * 30),
      sleepStartTime: sleepStart,
      sleepEndTime: sleepEnd
    }
  });

  // Добавляем данные тренировки если это устройство поддерживает
  if (['apple_watch', 'garmin', 'amazfit'].includes(deviceType)) {
    const workoutTime = new Date(now);
    workoutTime.setHours(19, 0, 0, 0);
    
    if (workoutTime < now) {
      data.push({
        timestamp: workoutTime,
        metrics: {
          workoutType: 'running',
          workoutDuration: 45,
          workoutIntensity: 'moderate',
          heartRate: Math.round(140 + Math.random() * 20),
          caloriesBurned: Math.round(350 + Math.random() * 100),
          distance: 5.2 + Math.random() * 2
        }
      });
    }
  }

  return data;
}
