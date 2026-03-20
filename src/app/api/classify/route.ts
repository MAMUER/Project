import { NextRequest, NextResponse } from 'next/server';
import ZAI from 'z-ai-web-dev-sdk';
import type { ClassificationInput, UserClassification } from '@/types';

// Классификация состояния пользователя на основе 6 параметров
// Это симуляция нейросети с 6 входными нейронами
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const input: ClassificationInput = body.input;

    if (!input) {
      return NextResponse.json(
        { success: false, error: 'Classification input is required' },
        { status: 400 }
      );
    }

    // Валидация входных данных
    const {
      avgHeartRate = 75,
      restingHeartRate = 60,
      sleepQuality = 70,
      activityLevel = 30,
      stressLevel = 30,
      recoveryScore = 70
    } = input;

    // Используем LLM для классификации (в продакшене заменяется на нейросеть)
    const zai = await ZAI.create();

    const systemPrompt = `Ты — медицинский AI-ассистент для анализа физического состояния. 
Твоя задача — классифицировать состояние пользователя на основе биометрических данных.

Входные параметры (6 показателей):
1. avgHeartRate — средний пульс (норма: 60-100)
2. restingHeartRate — пульс в покое (норма: 50-70)
3. sleepQuality — качество сна (0-100, норма: >70)
4. activityLevel — активность в минутах/день (норма: >30)
5. stressLevel — уровень стресса (0-100, норма: <40)
6. recoveryScore — оценка восстановления (0-100, норма: >60)

Классы состояния:
- excellent: все показатели в норме или лучше
- good: большинство показателей в норме
- moderate: есть отклонения, но не критичные
- needs_attention: несколько показателей вне нормы
- at_risk: критические отклонения

Уровни риска (low, moderate, high, very_high):
- cardiovascularRisk: риск сердечно-сосудистых проблем
- metabolicRisk: риск метаболических нарушений
- injuryRisk: риск травм при тренировках
- overtrainingRisk: риск перетренированности

Отвечай ТОЛЬКО валидным JSON в формате:
{
  "fitnessClass": "один из классов",
  "confidence": 0.85,
  "cardiovascularRisk": "low",
  "metabolicRisk": "moderate",
  "injuryRisk": "low",
  "overtrainingRisk": "low",
  "recommendations": ["рекомендация 1", "рекомендация 2"],
  "insights": ["инсайт 1"]
}`;

    const userMessage = `Проанализируй состояние пользователя:
- Средний пульс: ${avgHeartRate} уд/мин
- Пульс в покое: ${restingHeartRate} уд/мин
- Качество сна: ${sleepQuality}%
- Активность: ${activityLevel} мин/день
- Уровень стресса: ${stressLevel}%
- Оценка восстановления: ${recoveryScore}%

Дай классификацию и рекомендации.`;

    const completion = await zai.chat.completions.create({
      messages: [
        { role: 'assistant', content: systemPrompt },
        { role: 'user', content: userMessage }
      ],
      thinking: { type: 'disabled' }
    });

    const responseText = completion.choices[0]?.message?.content || '{}';

    // Парсим ответ LLM
    let classification: UserClassification;
    try {
      // Извлекаем JSON из ответа
      const jsonMatch = responseText.match(/\{[\s\S]*\}/);
      const jsonStr = jsonMatch ? jsonMatch[0] : responseText;
      classification = JSON.parse(jsonStr);
    } catch {
      // Fallback: простая логика классификации
      classification = simpleClassification(input);
    }

    // Добавляем confidence если отсутствует
    if (!classification.confidence) {
      classification.confidence = calculateConfidence(input);
    }

    return NextResponse.json({
      success: true,
      data: classification,
      input: input
    });

  } catch (error) {
    console.error('Classification error:', error);
    return NextResponse.json(
      { success: false, error: 'Failed to classify user state' },
      { status: 500 }
    );
  }
}

// Простая алгоритмическая классификация (fallback)
function simpleClassification(input: ClassificationInput): UserClassification {
  const { avgHeartRate, restingHeartRate, sleepQuality, activityLevel, stressLevel, recoveryScore } = input;

  // Вычисляем общий балл
  const heartScore = Math.max(0, 100 - Math.abs(avgHeartRate - 70) * 2);
  const restingScore = Math.max(0, 100 - Math.abs(restingHeartRate - 60) * 2);
  const activityScore = Math.min(100, activityLevel * 2);
  const stressScore = Math.max(0, 100 - stressLevel);

  const totalScore = (heartScore + restingScore + sleepQuality + activityScore + stressScore + recoveryScore) / 6;

  // Определяем класс
  let fitnessClass: UserClassification['fitnessClass'];
  if (totalScore >= 85) fitnessClass = 'excellent';
  else if (totalScore >= 70) fitnessClass = 'good';
  else if (totalScore >= 55) fitnessClass = 'moderate';
  else if (totalScore >= 40) fitnessClass = 'needs_attention';
  else fitnessClass = 'at_risk';

  // Определяем риски
  const cardiovascularRisk = avgHeartRate > 100 || restingHeartRate > 80 ? 'high' :
    avgHeartRate > 90 || restingHeartRate > 70 ? 'moderate' : 'low';

  const metabolicRisk = activityLevel < 20 && sleepQuality < 60 ? 'high' :
    activityLevel < 30 ? 'moderate' : 'low';

  const injuryRisk = recoveryScore < 50 || stressLevel > 70 ? 'high' :
    recoveryScore < 60 ? 'moderate' : 'low';

  const overtrainingRisk = recoveryScore < 40 && activityLevel > 90 ? 'very_high' :
    recoveryScore < 50 ? 'high' : 'low';

  // Генерируем рекомендации
  const recommendations: string[] = [];
  if (sleepQuality < 70) recommendations.push('Улучшите качество сна: ложитесь раньше, исключите экраны за час до сна');
  if (activityLevel < 30) recommendations.push('Увеличьте ежедневную активность до минимум 30 минут');
  if (stressLevel > 50) recommendations.push('Практикуйте техники расслабления: медитация, дыхательные упражнения');
  if (recoveryScore < 60) recommendations.push('Добавьте дни отдыха между интенсивными тренировками');
  if (restingHeartRate > 70) recommendations.push('Добавьте кардио низкой интенсивности для улучшения работы сердца');

  return {
    fitnessClass,
    confidence: Math.min(0.95, totalScore / 100),
    cardiovascularRisk: cardiovascularRisk as RiskLevel,
    metabolicRisk: metabolicRisk as RiskLevel,
    injuryRisk: injuryRisk as RiskLevel,
    overtrainingRisk: overtrainingRisk as RiskLevel,
    recommendations,
    insights: [`Общий балл состояния: ${Math.round(totalScore)}/100`]
  };
}

function calculateConfidence(input: ClassificationInput): number {
  // Чем ближе показатели к норме, тем выше уверенность
  const { avgHeartRate, restingHeartRate, sleepQuality, stressLevel } = input;
  
  const heartDeviation = Math.abs(avgHeartRate - 75) / 50;
  const restingDeviation = Math.abs(restingHeartRate - 60) / 40;
  const sleepDeviation = Math.abs(sleepQuality - 80) / 100;
  const stressDeviation = Math.abs(stressLevel - 30) / 100;

  const avgDeviation = (heartDeviation + restingDeviation + sleepDeviation + stressDeviation) / 4;
  
  return Math.max(0.5, Math.min(0.98, 1 - avgDeviation));
}

type RiskLevel = 'low' | 'moderate' | 'high' | 'very_high';
