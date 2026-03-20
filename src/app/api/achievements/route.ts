import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// Получение достижений пользователя
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const userId = searchParams.get('userId') || 'demo-user';

    // Получаем все достижения с прогрессом пользователя
    const allAchievements = await db.achievement.findMany();
    const userAchievements = await db.userAchievement.findMany({
      where: { userId },
      include: { achievement: true }
    });

    // Объединяем данные
    const achievementsWithProgress = allAchievements.map(achievement => {
      const userProgress = userAchievements.find(ua => ua.achievementId === achievement.id);
      return {
        id: achievement.id,
        name: achievement.name,
        description: achievement.description,
        category: achievement.category,
        icon: achievement.icon || '🏆',
        points: achievement.points,
        maxLevel: achievement.maxLevel,
        level: userProgress?.level || 0,
        progress: userProgress?.progress || 0,
        unlockedAt: userProgress?.unlockedAt || null
      };
    });

    // Если достижений нет в БД, возвращаем дефолтные
    if (achievementsWithProgress.length === 0) {
      return NextResponse.json({
        success: true,
        data: getDefaultAchievements()
      });
    }

    return NextResponse.json({
      success: true,
      data: achievementsWithProgress
    });

  } catch (error) {
    console.error('Achievements error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to fetch achievements' },
      { status: 500 }
    );
  }
}

// Разблокировка достижения
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { userId, achievementId } = body;

    if (!userId || !achievementId) {
      return NextResponse.json(
        { success: false, error: 'userId and achievementId are required' },
        { status: 400 }
      );
    }

    // Проверяем, не разблокировано ли уже
    const existing = await db.userAchievement.findUnique({
      where: {
        userId_achievementId: { userId, achievementId }
      }
    });

    if (existing) {
      // Увеличиваем уровень
      const achievement = await db.achievement.findUnique({
        where: { id: achievementId }
      });
      
      if (existing.level < (achievement?.maxLevel || 1)) {
        const updated = await db.userAchievement.update({
          where: { id: existing.id },
          data: {
            level: { increment: 1 },
            progress: 0
          }
        });
        return NextResponse.json({ success: true, data: updated });
      }
      
      return NextResponse.json({
        success: false,
        message: 'Achievement already at max level'
      });
    }

    // Создаем новое достижение
    const userAchievement = await db.userAchievement.create({
      data: {
        userId,
        achievementId,
        level: 1,
        progress: 100
      }
    });

    return NextResponse.json({
      success: true,
      data: userAchievement,
      message: 'Achievement unlocked!'
    });

  } catch (error) {
    console.error('Unlock achievement error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to unlock achievement' },
      { status: 500 }
    );
  }
}

function getDefaultAchievements() {
  return [
    {
      id: 'first_steps',
      name: 'Первые шаги',
      description: 'Начните отслеживать свою активность',
      category: 'fitness',
      icon: '🎯',
      points: 10,
      maxLevel: 1,
      level: 1,
      progress: 100,
      unlockedAt: new Date()
    },
    {
      id: 'week_streak',
      name: 'Неделя активности',
      description: '7 дней подряд с активностью более 5000 шагов',
      category: 'consistency',
      icon: '🔥',
      points: 50,
      maxLevel: 3,
      level: 0,
      progress: 71,
      unlockedAt: null
    },
    {
      id: 'early_bird',
      name: 'Ранняя пташка',
      description: 'Завершите 10 утренних тренировок до 8:00',
      category: 'fitness',
      icon: '🌅',
      points: 30,
      maxLevel: 2,
      level: 0,
      progress: 40,
      unlockedAt: null
    },
    {
      id: 'marathon',
      name: 'Марафонец',
      description: 'Пробегите 42 км суммарно',
      category: 'fitness',
      icon: '🏃',
      points: 100,
      maxLevel: 1,
      level: 0,
      progress: 23,
      unlockedAt: null
    },
    {
      id: 'sleep_master',
      name: 'Мастер сна',
      description: 'Спите 7+ часов 30 дней подряд',
      category: 'consistency',
      icon: '😴',
      points: 75,
      maxLevel: 2,
      level: 0,
      progress: 33,
      unlockedAt: null
    },
    {
      id: 'first_program',
      name: 'Программист',
      description: 'Создайте первую программу тренировок',
      category: 'special',
      icon: '📋',
      points: 20,
      maxLevel: 1,
      level: 1,
      progress: 100,
      unlockedAt: new Date()
    },
    {
      id: 'social_butterfly',
      name: 'Командный игрок',
      description: 'Присоединитесь к фитнес-клубу',
      category: 'social',
      icon: '🤝',
      points: 25,
      maxLevel: 1,
      level: 0,
      progress: 0,
      unlockedAt: null
    },
    {
      id: 'calories_burner',
      name: 'Сжигатель калорий',
      description: 'Сожгите 10,000 калорий суммарно',
      category: 'fitness',
      icon: '🔥',
      points: 40,
      maxLevel: 5,
      level: 1,
      progress: 0,
      unlockedAt: new Date()
    }
  ];
}
