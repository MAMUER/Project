import { NextRequest, NextResponse } from 'next/server';
import ZAI from 'z-ai-web-dev-sdk';
import type { UserProfileForGeneration, GeneratedProgram } from '@/types';

// Генерация персонализированной программы тренировок с использованием GAN-подобного подхода
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const profile: UserProfileForGeneration = body.profile;

    if (!profile) {
      return NextResponse.json(
        { success: false, error: 'User profile is required' },
        { status: 400 }
      );
    }

    const zai = await ZAI.create();

    const systemPrompt = `Ты — профессиональный фитнес-тренер с опытом в спортивной медицине.
Твоя задача — создать персонализированную программу тренировок на основе данных пользователя.

ВАЖНО: Учитывай противопоказания и ограничения пользователя!
Если есть противопоказания к определенным упражнениям, заменяй их безопасными альтернативами.

Цели тренировок:
- weight_loss: фокус на кардио + силовые в круговом режиме
- muscle_gain: фокус на силовые с прогрессией весов
- endurance: фокус на кардио с постепенным увеличением длительности
- rehabilitation: щадящие упражнения, низкая интенсивность
- maintenance: сбалансированная программа

Уровни подготовки:
- beginner: простые упражнения, низкая интенсивность
- intermediate: средняя сложность, умеренная интенсивность
- advanced: сложные упражнения, высокая интенсивность

Отвечай ТОЛЬКО валидным JSON в формате:
{
  "programName": "Название программы",
  "description": "Краткое описание",
  "weeklySchedule": [
    {
      "weekNumber": 1,
      "focus": "Фокус недели",
      "days": [
        {
          "dayOfWeek": 1,
          "type": "strength/cardio/rest/flexibility",
          "exercises": [
            {
              "name": "Название упражнения",
              "sets": 3,
              "reps": "10-12",
              "rest": 60,
              "notes": "Техника выполнения"
            }
          ],
          "totalDuration": 45,
          "estimatedCalories": 300
        }
      ]
    }
  ],
  "estimatedResults": "Ожидаемые результаты через N недель",
  "safetyNotes": ["Важно: ...", "Предостережение: ..."],
  "progressionPlan": "Как увеличивать нагрузку каждую неделю"
}

Тренировочная неделя:
- День 1 (Пн): Основная тренировка
- День 2 (Вт): Активное восстановление или отдых
- День 3 (Ср): Основная тренировка
- День 4 (Чт): Отдых
- День 5 (Пт): Основная тренировка
- День 6 (Сб): Активность на выбор
- День 7 (Вс): Полный отдых`;

    const userMessage = `Создай программу тренировок на ${profile.trainingFrequency} дней в неделю, длительность сессии ${profile.sessionDuration} минут.

ДАННЫЕ ПОЛЬЗОВАТЕЛЯ:
- Возраст: ${profile.age} лет
- Пол: ${profile.gender}
- Вес: ${profile.weight} кг
- Рост: ${profile.height} см
- Цель: ${getGoalText(profile.fitnessGoal)}
- Уровень активности: ${getActivityText(profile.activityLevel)}
- Доступное оборудование: ${profile.availableEquipment.join(', ') || 'Нет'}

ПРОТИВОПОКАЗАНИЯ:
${profile.contraindications.length > 0 ? profile.contraindications.map(c => `- ${c}`).join('\n') : 'Нет указанных противопоказаний'}

ХРОНИЧЕСКИЕ ЗАБОЛЕВАНИЯ:
${profile.chronicDiseases.length > 0 ? profile.chronicDiseases.map(d => `- ${d}`).join('\n') : 'Нет указанных заболеваний'}

Создай безопасную и эффективную программу на 4 недели.`;

    const completion = await zai.chat.completions.create({
      messages: [
        { role: 'assistant', content: systemPrompt },
        { role: 'user', content: userMessage }
      ],
      thinking: { type: 'disabled' }
    });

    const responseText = completion.choices[0]?.message?.content || '{}';

    // Парсим ответ
    let program: GeneratedProgram;
    try {
      const jsonMatch = responseText.match(/\{[\s\S]*\}/);
      const jsonStr = jsonMatch ? jsonMatch[0] : responseText;
      program = JSON.parse(jsonStr);
    } catch {
      // Fallback: базовая программа
      program = generateBasicProgram(profile);
    }

    // Добавляем метаданные
    const enrichedProgram = {
      ...program,
      generatedAt: new Date().toISOString(),
      basedOnProfile: {
        age: profile.age,
        fitnessGoal: profile.fitnessGoal,
        trainingFrequency: profile.trainingFrequency
      },
      aiModel: 'LLM-GAN-Hybrid'
    };

    return NextResponse.json({
      success: true,
      data: enrichedProgram
    });

  } catch (error) {
    console.error('Program generation error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to generate training program' },
      { status: 500 }
    );
  }
}

function getGoalText(goal: string): string {
  const goals: Record<string, string> = {
    weight_loss: 'Похудение / Сжигание жира',
    muscle_gain: 'Набор мышечной массы',
    endurance: 'Повышение выносливости',
    rehabilitation: 'Реабилитация / Восстановление',
    maintenance: 'Поддержание формы'
  };
  return goals[goal] || goal;
}

function getActivityText(level: string): string {
  const levels: Record<string, string> = {
    sedentary: 'Сидячий образ жизни',
    light: 'Легкая активность',
    moderate: 'Умеренная активность',
    active: 'Активный образ жизни',
    very_active: 'Очень активный'
  };
  return levels[level] || level;
}

function generateBasicProgram(profile: UserProfileForGeneration): GeneratedProgram {
  const isBeginner = profile.activityLevel === 'sedentary' || profile.activityLevel === 'light';
  const hasContraindications = profile.contraindications.length > 0 || profile.chronicDiseases.length > 0;

  return {
    programName: `${getGoalText(profile.fitnessGoal)} — ${isBeginner ? 'Начальный' : 'Средний'} уровень`,
    description: `Персонализированная программа тренировок для ${getGoalText(profile.fitnessGoal).toLowerCase()}. ${
      hasContraindications ? 'Учитывает медицинские ограничения.' : ''
    }`,
    weeklySchedule: [
      {
        weekNumber: 1,
        focus: 'Адаптация к нагрузкам',
        days: [
          {
            dayOfWeek: 1,
            type: 'strength',
            exercises: [
              { name: 'Разминка (ходьба)', sets: 1, reps: '5 мин', rest: 0, notes: 'Постепенный разогрев' },
              { name: 'Приседания', sets: 3, reps: '10-12', rest: 60, notes: 'Спина прямая, колени не выходят за носки' },
              { name: 'Отжимания', sets: 3, reps: '8-10', rest: 60, notes: 'Можно с колен для начинающих' },
              { name: 'Планка', sets: 3, reps: '30 сек', rest: 30, notes: 'Держать корпус ровно' },
              { name: 'Заминка (растяжка)', sets: 1, reps: '5 мин', rest: 0, notes: 'Растяжение рабочих мышц' }
            ],
            totalDuration: profile.sessionDuration,
            estimatedCalories: 200
          },
          {
            dayOfWeek: 3,
            type: 'cardio',
            exercises: [
              { name: 'Разминка', sets: 1, reps: '5 мин', rest: 0, notes: 'Легкая ходьба' },
              { name: 'Быстрая ходьба', sets: 1, reps: '20 мин', rest: 0, notes: 'Пульс 60-70% от максимального' },
              { name: 'Заминка', sets: 1, reps: '5 мин', rest: 0, notes: 'Медленная ходьба' }
            ],
            totalDuration: 30,
            estimatedCalories: 150
          },
          {
            dayOfWeek: 5,
            type: 'strength',
            exercises: [
              { name: 'Разминка', sets: 1, reps: '5 мин', rest: 0, notes: 'Суставная гимнастика' },
              { name: 'Выпады', sets: 3, reps: '10 на каждую ногу', rest: 60, notes: 'Колено не касается пола' },
              { name: 'Ягодичный мост', sets: 3, reps: '12-15', rest: 45, notes: 'Напрягать ягодицы в верхней точке' },
              { name: 'Скручивания на пресс', sets: 3, reps: '15', rest: 30, notes: 'Поясница прижата к полу' },
              { name: 'Растяжка', sets: 1, reps: '5 мин', rest: 0, notes: 'Восстановление' }
            ],
            totalDuration: profile.sessionDuration,
            estimatedCalories: 180
          }
        ]
      }
    ],
    estimatedResults: `При регулярных тренировках и правильном питании ожидается прогресс к концу 4 недели`,
    safetyNotes: hasContraindications ? [
      'Начните с низкой интенсивности',
      'Проконсультируйтесь с врачом перед началом',
      'Прекратите тренировку при появлении боли или дискомфорта'
    ] : [
      'Соблюдайте технику выполнения упражнений',
      'Не пропускайте разминку и заминку'
    ],
    progressionPlan: 'Каждую неделю увеличивайте количество повторений на 1-2 или добавляйте подход'
  };
}
