'use client';

import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Activity,
  Heart,
  Moon,
  Flame,
  Footprints,
  TrendingUp,
  AlertTriangle,
  Trophy,
  Watch,
  RefreshCw,
  Dumbbell,
  Target,
  ChevronRight,
  Sparkles,
  Brain,
  Zap
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts';
import type { DashboardData, UserClassification, GeneratedProgram } from '@/types';

export default function FitnessDashboard() {
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [classification, setClassification] = useState<UserClassification | null>(null);
  const [generatedProgram, setGeneratedProgram] = useState<GeneratedProgram | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [syncing, setSyncing] = useState(false);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    setIsLoading(true);
    try {
      const response = await fetch('/api/dashboard?userId=demo-user');
      const data = await response.json();
      if (data.success) {
        setDashboardData(data.data);
      }
    } catch (error) {
      console.error('Error loading dashboard:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSyncDevice = async (deviceType: string) => {
    setSyncing(true);
    try {
      const response = await fetch('/api/sync-device', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ userId: 'demo-user', deviceType })
      });
      const data = await response.json();
      if (data.success) {
        // Обновляем дашборд после синхронизации
        await loadDashboardData();
      }
    } catch (error) {
      console.error('Sync error:', error);
    } finally {
      setSyncing(false);
    }
  };

  const handleClassify = async () => {
    try {
      const response = await fetch('/api/classify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          input: {
            avgHeartRate: dashboardData?.todayStats.heartRate || 72,
            restingHeartRate: 58,
            sleepQuality: dashboardData?.todayStats.sleepHours ? Math.min(100, dashboardData.todayStats.sleepHours * 12) : 70,
            activityLevel: dashboardData?.todayStats.activeMinutes || 30,
            stressLevel: 35,
            recoveryScore: 75
          }
        })
      });
      const data = await response.json();
      if (data.success) {
        setClassification(data.data);
      }
    } catch (error) {
      console.error('Classification error:', error);
    }
  };

  const handleGenerateProgram = async () => {
    try {
      const response = await fetch('/api/generate-program', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile: {
            age: 30,
            gender: 'male',
            weight: 75,
            height: 178,
            fitnessGoal: 'maintenance',
            activityLevel: 'moderate',
            contraindications: [],
            chronicDiseases: [],
            availableEquipment: ['dumbbells', 'pull-up bar', 'yoga mat'],
            trainingFrequency: 4,
            sessionDuration: 45
          }
        })
      });
      const data = await response.json();
      if (data.success) {
        setGeneratedProgram(data.data);
      }
    } catch (error) {
      console.error('Program generation error:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
        >
          <Activity className="w-12 h-12 text-emerald-400" />
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      {/* Header */}
      <header className="sticky top-0 z-50 backdrop-blur-xl bg-slate-900/80 border-b border-slate-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-400 to-cyan-400 flex items-center justify-center">
                <Activity className="w-6 h-6 text-slate-900" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-white">FitHealth AI</h1>
                <p className="text-xs text-slate-400">Интеллектуальная фитнес-платформа</p>
              </div>
            </div>
            
            <div className="flex items-center gap-3">
              <Badge variant="outline" className="border-emerald-500/50 text-emerald-400">
                <Sparkles className="w-3 h-3 mr-1" />
                AI-Powered
              </Badge>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleSyncDevice('apple_watch')}
                disabled={syncing}
                className="border-slate-600 text-slate-300 hover:bg-slate-700"
              >
                {syncing ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : (
                  <Watch className="w-4 h-4" />
                )}
                <span className="ml-2 hidden sm:inline">Синхронизация</span>
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 pb-20">
        {/* Welcome Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6"
        >
          <h2 className="text-2xl font-bold text-white">
            Добро пожаловать, {dashboardData?.user.name || 'Пользователь'}!
          </h2>
          <p className="text-slate-400 mt-1">
            Ваша цель: {getGoalText(dashboardData?.user.fitnessGoal || 'maintenance')}
          </p>
        </motion.div>

        {/* Health Alerts */}
        {dashboardData?.healthAlerts && dashboardData.healthAlerts.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <Card className="bg-amber-500/10 border-amber-500/30">
              <CardContent className="p-4">
                <div className="flex items-start gap-3">
                  <AlertTriangle className="w-5 h-5 text-amber-400 mt-0.5" />
                  <div>
                    <p className="text-amber-400 font-medium">Уведомления о здоровье</p>
                    <ul className="mt-2 space-y-1">
                      {dashboardData.healthAlerts.map((alert, i) => (
                        <li key={i} className="text-sm text-amber-300/80">{alert}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Today's Stats Grid */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6"
        >
          <StatsCard
            icon={<Footprints className="w-5 h-5" />}
            label="Шаги"
            value={dashboardData?.todayStats.steps || 0}
            target={10000}
            unit=""
            color="emerald"
          />
          <StatsCard
            icon={<Flame className="w-5 h-5" />}
            label="Калории"
            value={dashboardData?.todayStats.calories || 0}
            target={2500}
            unit="ккал"
            color="orange"
          />
          <StatsCard
            icon={<Heart className="w-5 h-5" />}
            label="Пульс"
            value={dashboardData?.todayStats.heartRate || 0}
            target={null}
            unit="уд/мин"
            color="rose"
          />
          <StatsCard
            icon={<Moon className="w-5 h-5" />}
            label="Сон"
            value={dashboardData?.todayStats.sleepHours || 0}
            target={8}
            unit="ч"
            color="violet"
          />
        </motion.div>

        {/* Tabs Section */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid grid-cols-2 md:grid-cols-4 gap-2 bg-slate-800/50 p-1 rounded-xl">
            <TabsTrigger value="overview" className="data-[state=active]:bg-slate-700 data-[state=active]:text-emerald-400">
              <TrendingUp className="w-4 h-4 mr-2" />
              Обзор
            </TabsTrigger>
            <TabsTrigger value="ai" className="data-[state=active]:bg-slate-700 data-[state=active]:text-emerald-400">
              <Brain className="w-4 h-4 mr-2" />
              AI Анализ
            </TabsTrigger>
            <TabsTrigger value="program" className="data-[state=active]:bg-slate-700 data-[state=active]:text-emerald-400">
              <Dumbbell className="w-4 h-4 mr-2" />
              Программа
            </TabsTrigger>
            <TabsTrigger value="achievements" className="data-[state=active]:bg-slate-700 data-[state=active]:text-emerald-400">
              <Trophy className="w-4 h-4 mr-2" />
              Достижения
            </TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-6">
            {/* Weekly Progress Chart */}
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle className="text-white">Недельная активность</CardTitle>
                <CardDescription className="text-slate-400">Ваши показатели за последние 7 дней</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={dashboardData?.weeklyProgress || []}>
                      <defs>
                        <linearGradient id="colorSteps" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                          <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                        </linearGradient>
                        <linearGradient id="colorCalories" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#f97316" stopOpacity={0.3}/>
                          <stop offset="95%" stopColor="#f97316" stopOpacity={0}/>
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis 
                        dataKey="date" 
                        stroke="#9ca3af" 
                        fontSize={12}
                        tickFormatter={(value) => new Date(value).toLocaleDateString('ru', { weekday: 'short' })}
                      />
                      <YAxis stroke="#9ca3af" fontSize={12} />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: '#1f2937',
                          border: '1px solid #374151',
                          borderRadius: '8px'
                        }}
                        labelStyle={{ color: '#fff' }}
                      />
                      <Area
                        type="monotone"
                        dataKey="steps"
                        stroke="#10b981"
                        fillOpacity={1}
                        fill="url(#colorSteps)"
                        name="Шаги"
                      />
                      <Area
                        type="monotone"
                        dataKey="calories"
                        stroke="#f97316"
                        fillOpacity={1}
                        fill="url(#colorCalories)"
                        name="Калории"
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </CardContent>
            </Card>

            {/* Current Program Progress */}
            {dashboardData?.currentProgram && (
              <Card className="bg-slate-800/50 border-slate-700">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-white">{dashboardData.currentProgram.name}</CardTitle>
                      <CardDescription className="text-slate-400">
                        Осталось {dashboardData.currentProgram.daysRemaining} дней
                      </CardDescription>
                    </div>
                    <Button variant="outline" size="sm" className="border-slate-600">
                      Детали
                      <ChevronRight className="w-4 h-4 ml-1" />
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-400">Прогресс</span>
                      <span className="text-emerald-400 font-medium">{dashboardData.currentProgram.progress}%</span>
                    </div>
                    <Progress value={dashboardData.currentProgram.progress} className="h-2" />
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* AI Analysis Tab */}
          <TabsContent value="ai" className="space-y-6">
            <div className="grid md:grid-cols-2 gap-6">
              {/* Classification Card */}
              <Card className="bg-slate-800/50 border-slate-700">
                <CardHeader>
                  <CardTitle className="text-white flex items-center gap-2">
                    <Brain className="w-5 h-5 text-emerald-400" />
                    Классификация состояния
                  </CardTitle>
                  <CardDescription className="text-slate-400">
                    AI-анализ на основе 6 биометрических параметров
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <Button
                    onClick={handleClassify}
                    className="w-full bg-gradient-to-r from-emerald-500 to-cyan-500 hover:from-emerald-600 hover:to-cyan-600"
                  >
                    <Zap className="w-4 h-4 mr-2" />
                    Запустить анализ
                  </Button>

                  {classification && (
                    <motion.div
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="space-y-4 pt-4"
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-slate-400">Класс состояния:</span>
                        <Badge className={getClassColor(classification.fitnessClass)}>
                          {getClassText(classification.fitnessClass)}
                        </Badge>
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <span className="text-slate-400">Уверенность:</span>
                        <span className="text-white">{Math.round(classification.confidence * 100)}%</span>
                      </div>

                      <div className="space-y-2 pt-2 border-t border-slate-700">
                        <p className="text-sm text-slate-400">Уровни риска:</p>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <RiskBadge label="Сердечный" level={classification.cardiovascularRisk} />
                          <RiskBadge label="Метаболизм" level={classification.metabolicRisk} />
                          <RiskBadge label="Травмы" level={classification.injuryRisk} />
                          <RiskBadge label="Перетренир." level={classification.overtrainingRisk} />
                        </div>
                      </div>

                      {classification.recommendations.length > 0 && (
                        <div className="space-y-2 pt-2 border-t border-slate-700">
                          <p className="text-sm text-slate-400">Рекомендации:</p>
                          <ul className="space-y-1">
                            {classification.recommendations.map((rec, i) => (
                              <li key={i} className="text-sm text-slate-300 flex items-start gap-2">
                                <span className="text-emerald-400 mt-1">•</span>
                                {rec}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </motion.div>
                  )}
                </CardContent>
              </Card>

              {/* Neural Network Visualization */}
              <Card className="bg-slate-800/50 border-slate-700">
                <CardHeader>
                  <CardTitle className="text-white flex items-center gap-2">
                    <Activity className="w-5 h-5 text-cyan-400" />
                    Входные параметры модели
                  </CardTitle>
                  <CardDescription className="text-slate-400">
                    6 входных нейронов для классификации
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-3">
                    <NeuralInput
                      label="Средний пульс"
                      value={dashboardData?.todayStats.heartRate || 72}
                      unit="уд/мин"
                      normal={60 <= (dashboardData?.todayStats.heartRate || 72) && (dashboardData?.todayStats.heartRate || 72) <= 100}
                    />
                    <NeuralInput
                      label="Пульс покоя"
                      value={58}
                      unit="уд/мин"
                      normal={true}
                    />
                    <NeuralInput
                      label="Качество сна"
                      value={75}
                      unit="%"
                      normal={75 >= 70}
                    />
                    <NeuralInput
                      label="Активность"
                      value={dashboardData?.todayStats.activeMinutes || 30}
                      unit="мин"
                      normal={(dashboardData?.todayStats.activeMinutes || 30) >= 30}
                    />
                    <NeuralInput
                      label="Стресс"
                      value={35}
                      unit="/ 100"
                      normal={35 < 40}
                    />
                    <NeuralInput
                      label="Восстановление"
                      value={75}
                      unit="%"
                      normal={75 >= 60}
                    />
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Program Tab */}
          <TabsContent value="program" className="space-y-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Target className="w-5 h-5 text-emerald-400" />
                  AI-генерация программы
                </CardTitle>
                <CardDescription className="text-slate-400">
                  Персонализированная программа на основе ваших данных и целей
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Button
                  onClick={handleGenerateProgram}
                  className="w-full bg-gradient-to-r from-emerald-500 to-cyan-500 hover:from-emerald-600 hover:to-cyan-600"
                >
                  <Sparkles className="w-4 h-4 mr-2" />
                  Сгенерировать программу
                </Button>

                {generatedProgram && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="space-y-4 pt-4"
                  >
                    <div className="p-4 rounded-lg bg-slate-700/50 border border-slate-600">
                      <h4 className="font-semibold text-white">{generatedProgram.programName}</h4>
                      <p className="text-sm text-slate-400 mt-1">{generatedProgram.description}</p>
                    </div>

                    {generatedProgram.weeklySchedule.map((week) => (
                      <div key={week.weekNumber} className="space-y-3">
                        <h4 className="font-medium text-white">
                          Неделя {week.weekNumber}: {week.focus}
                        </h4>
                        <div className="grid gap-2">
                          {week.days.map((day) => (
                            <div
                              key={day.dayOfWeek}
                              className="p-3 rounded-lg bg-slate-700/30 border border-slate-600/50"
                            >
                              <div className="flex items-center justify-between mb-2">
                                <span className="font-medium text-white">
                                  {getDayName(day.dayOfWeek)}
                                </span>
                                <div className="flex items-center gap-2 text-sm text-slate-400">
                                  <span>{day.totalDuration} мин</span>
                                  <span>•</span>
                                  <span>{day.estimatedCalories} ккал</span>
                                </div>
                              </div>
                              <div className="flex flex-wrap gap-1">
                                {day.exercises.slice(0, 4).map((ex, i) => (
                                  <Badge key={i} variant="outline" className="text-xs border-slate-500">
                                    {ex.name}
                                  </Badge>
                                ))}
                                {day.exercises.length > 4 && (
                                  <Badge variant="outline" className="text-xs border-slate-500">
                                    +{day.exercises.length - 4}
                                  </Badge>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}

                    {generatedProgram.safetyNotes.length > 0 && (
                      <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/30">
                        <p className="text-sm font-medium text-amber-400 mb-2">Меры предосторожности:</p>
                        <ul className="space-y-1">
                          {generatedProgram.safetyNotes.map((note, i) => (
                            <li key={i} className="text-sm text-amber-300/80">• {note}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </motion.div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* Achievements Tab */}
          <TabsContent value="achievements" className="space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {dashboardData?.recentAchievements.map((achievement) => (
                <AchievementCard key={achievement.id} achievement={achievement} />
              ))}
            </div>
          </TabsContent>
        </Tabs>
      </main>

      {/* Footer */}
      <footer className="fixed bottom-0 left-0 right-0 bg-slate-900/95 backdrop-blur-xl border-t border-slate-700/50 py-3 px-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">
              Синхронизация с:
            </span>
            <div className="flex gap-2">
              {['apple_watch', 'samsung_health', 'huawei_health'].map((device) => (
                <Button
                  key={device}
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSyncDevice(device)}
                  className="text-slate-400 hover:text-white"
                >
                  <Watch className="w-4 h-4" />
                </Button>
              ))}
            </div>
          </div>
          <div className="text-xs text-slate-500">
            Powered by AI • FitHealth Platform
          </div>
        </div>
      </footer>
    </div>
  );
}

// Component: Stats Card
function StatsCard({ 
  icon, 
  label, 
  value, 
  target, 
  unit, 
  color 
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  target: number | null;
  unit: string;
  color: string;
}) {
  const colorClasses: Record<string, string> = {
    emerald: 'from-emerald-500/20 to-emerald-500/5 border-emerald-500/30',
    orange: 'from-orange-500/20 to-orange-500/5 border-orange-500/30',
    rose: 'from-rose-500/20 to-rose-500/5 border-rose-500/30',
    violet: 'from-violet-500/20 to-violet-500/5 border-violet-500/30'
  };

  const iconColors: Record<string, string> = {
    emerald: 'text-emerald-400',
    orange: 'text-orange-400',
    rose: 'text-rose-400',
    violet: 'text-violet-400'
  };

  const progress = target ? Math.min(100, (value / target) * 100) : null;

  return (
    <Card className={`bg-gradient-to-br ${colorClasses[color]} border`}>
      <CardContent className="p-4">
        <div className="flex items-center gap-2 mb-2">
          <div className={iconColors[color]}>{icon}</div>
          <span className="text-sm text-slate-400">{label}</span>
        </div>
        <div className="flex items-baseline gap-1">
          <span className="text-2xl font-bold text-white">{value.toLocaleString()}</span>
          {unit && <span className="text-sm text-slate-400">{unit}</span>}
        </div>
        {progress !== null && (
          <Progress value={progress} className="h-1 mt-2" />
        )}
      </CardContent>
    </Card>
  );
}

// Component: Neural Input
function NeuralInput({ 
  label, 
  value, 
  unit, 
  normal 
}: {
  label: string;
  value: number;
  unit: string;
  normal: boolean;
}) {
  return (
    <div className={`p-3 rounded-lg border ${normal ? 'bg-emerald-500/10 border-emerald-500/30' : 'bg-amber-500/10 border-amber-500/30'}`}>
      <p className="text-xs text-slate-400 mb-1">{label}</p>
      <div className="flex items-baseline gap-1">
        <span className="text-lg font-semibold text-white">{value}</span>
        <span className="text-xs text-slate-400">{unit}</span>
      </div>
      <div className={`mt-1 text-xs ${normal ? 'text-emerald-400' : 'text-amber-400'}`}>
        {normal ? '✓ В норме' : '⚠ Внимание'}
      </div>
    </div>
  );
}

// Component: Risk Badge
function RiskBadge({ label, level }: { label: string; level: string }) {
  const colors: Record<string, string> = {
    low: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    moderate: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    high: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
    very_high: 'bg-red-500/20 text-red-400 border-red-500/30'
  };

  const labels: Record<string, string> = {
    low: 'Низкий',
    moderate: 'Средний',
    high: 'Высокий',
    very_high: 'Очень высокий'
  };

  return (
    <div className={`p-2 rounded border ${colors[level]}`}>
      <p className="text-xs opacity-70">{label}</p>
      <p className="font-medium text-sm">{labels[level]}</p>
    </div>
  );
}

// Component: Achievement Card
function AchievementCard({ achievement }: { achievement: any }) {
  const isUnlocked = achievement.unlockedAt;
  
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className={`relative p-4 rounded-xl border ${
        isUnlocked 
          ? 'bg-gradient-to-br from-emerald-500/20 to-cyan-500/10 border-emerald-500/30' 
          : 'bg-slate-800/50 border-slate-700 opacity-60'
      }`}
    >
      {!isUnlocked && (
        <div className="absolute inset-0 bg-slate-900/50 rounded-xl flex items-center justify-center">
          <span className="text-2xl">🔒</span>
        </div>
      )}
      
      <div className="text-3xl mb-2">{achievement.icon}</div>
      <h4 className="font-semibold text-white text-sm">{achievement.name}</h4>
      <p className="text-xs text-slate-400 mt-1">{achievement.description}</p>
      
      {achievement.maxLevel > 1 && (
        <div className="mt-2">
          <div className="flex justify-between text-xs mb-1">
            <span className="text-slate-400">Уровень</span>
            <span className="text-white">{achievement.level}/{achievement.maxLevel}</span>
          </div>
          <Progress value={(achievement.level / achievement.maxLevel) * 100} className="h-1" />
        </div>
      )}
      
      <div className="flex items-center justify-between mt-3 pt-2 border-t border-slate-700/50">
        <span className="text-xs text-slate-400">{achievement.category}</span>
        <span className="text-xs font-medium text-emerald-400">+{achievement.points} pts</span>
      </div>
    </motion.div>
  );
}

// Helpers
function getGoalText(goal: string): string {
  const goals: Record<string, string> = {
    weight_loss: 'Похудение',
    muscle_gain: 'Набор массы',
    endurance: 'Выносливость',
    rehabilitation: 'Реабилитация',
    maintenance: 'Поддержание формы'
  };
  return goals[goal] || goal;
}

function getClassText(cls: string): string {
  const classes: Record<string, string> = {
    excellent: 'Отличное',
    good: 'Хорошее',
    moderate: 'Умеренное',
    needs_attention: 'Требует внимания',
    at_risk: 'В зоне риска'
  };
  return classes[cls] || cls;
}

function getClassColor(cls: string): string {
  const colors: Record<string, string> = {
    excellent: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    good: 'bg-cyan-500/20 text-cyan-400 border-cyan-500/30',
    moderate: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    needs_attention: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
    at_risk: 'bg-red-500/20 text-red-400 border-red-500/30'
  };
  return colors[cls] || '';
}

function getDayName(day: number): string {
  const days = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
  return days[day - 1] || `День ${day}`;
}
