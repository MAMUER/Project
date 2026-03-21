import React, { useState } from 'react';
import { training } from '../services/api';

const TrainingPage: React.FC = () => {
    const [generating, setGenerating] = useState(false);
    const [program, setProgram] = useState<any>(null);
    const [formData, setFormData] = useState({
        training_class: 'cardio',
        goals: ['weight_loss'],
        fitness_level: 'beginner',
        age_group: 'adult',
        gender: 'female',
        has_injury: false,
        duration_weeks: 12
    });

    const handleGenerate = async () => {
        setGenerating(true);
        try {
            const response = await training.generate(formData);
            setProgram(response.data);
        } catch (err) {
            console.error(err);
        } finally {
            setGenerating(false);
        }
    };

    return (
        <div className="training-page">
            <h1>Генерация программы тренировок</h1>
            
            <div className="form-group">
                <label>Класс тренировки</label>
                <select value={formData.training_class} onChange={(e) => setFormData({...formData, training_class: e.target.value})}>
                    <option value="cardio">Кардио</option>
                    <option value="strength">Силовая</option>
                    <option value="flexibility">Гибкость</option>
                    <option value="recovery">Восстановление</option>
                    <option value="hiit">HIIT</option>
                    <option value="endurance">Выносливость</option>
                </select>
            </div>
            
            <div className="form-group">
                <label>Уровень подготовки</label>
                <select value={formData.fitness_level} onChange={(e) => setFormData({...formData, fitness_level: e.target.value})}>
                    <option value="beginner">Начинающий</option>
                    <option value="intermediate">Средний</option>
                    <option value="advanced">Продвинутый</option>
                </select>
            </div>
            
            <div className="form-group">
                <label>Длительность (недели)</label>
                <input type="number" value={formData.duration_weeks} onChange={(e) => setFormData({...formData, duration_weeks: parseInt(e.target.value)})} />
            </div>
            
            <button onClick={handleGenerate} disabled={generating}>
                {generating ? 'Генерация...' : 'Сгенерировать программу'}
            </button>
            
            {program && (
                <div className="program-result">
                    <h2>Ваша программа тренировок</h2>
                    <pre>{JSON.stringify(program, null, 2)}</pre>
                </div>
            )}
        </div>
    );
};

export default TrainingPage;