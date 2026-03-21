import React, { useState } from 'react';

interface BiometricFormProps {
    onSubmit: (data: any) => void;
}

const BiometricForm: React.FC<BiometricFormProps> = ({ onSubmit }) => {
    const [formData, setFormData] = useState({
        heart_rate: 70,
        systolic: 120,
        diastolic: 80,
        spo2: 98,
        temperature: 36.6,
        sleep_duration: 480,
        deep_sleep: 90
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData({
            ...formData,
            [e.target.name]: e.target.value
        });
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <form className="biometric-form" onSubmit={handleSubmit}>
            <h2>Ввод биометрических данных</h2>
            
            <label>Пульс (уд/мин)</label>
            <input type="number" name="heart_rate" value={formData.heart_rate} onChange={handleChange} />
            
            <label>Давление (сист/диаст)</label>
            <div className="row">
                <input type="number" name="systolic" value={formData.systolic} onChange={handleChange} />
                <span>/</span>
                <input type="number" name="diastolic" value={formData.diastolic} onChange={handleChange} />
            </div>
            
            <label>SpO2 (%)</label>
            <input type="number" name="spo2" value={formData.spo2} onChange={handleChange} />
            
            <label>Температура (°C)</label>
            <input type="number" step="0.1" name="temperature" value={formData.temperature} onChange={handleChange} />
            
            <label>Длительность сна (мин)</label>
            <input type="number" name="sleep_duration" value={formData.sleep_duration} onChange={handleChange} />
            
            <label>Глубокий сон (мин)</label>
            <input type="number" name="deep_sleep" value={formData.deep_sleep} onChange={handleChange} />
            
            <button type="submit">Отправить</button>
        </form>
    );
};

export default BiometricForm;