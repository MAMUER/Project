// ML Classification and Plan Generation functions

// ===================== ML Classification =====================

async function loadBiometricParams() {
    try {
        const records = await getBiometricRecords('', null, null, 50);

        let hr = 0, hrCount = 0;
        let hrv = 50;
        let spo2 = 0, spo2Count = 0;
        let temp = 0, tempCount = 0;
        let bp = 0, bpCount = 0;

        records.records?.forEach(rec => {
            switch (rec.metric_type) {
                case 'heart_rate':
                    hr += rec.value;
                    hrCount++;
                    break;
                case 'hrv':
                    hrv = rec.value;
                    break;
                case 'spo2':
                    spo2 += rec.value;
                    spo2Count++;
                    break;
                case 'temperature':
                    temp += rec.value;
                    tempCount++;
                    break;
                case 'blood_pressure':
                    bp += rec.value;
                    bpCount++;
                    break;
            }
        });

        const hrEl = document.getElementById('hrValue');
        const hrvEl = document.getElementById('hrvValue');
        const spo2El = document.getElementById('spo2Value');
        const tempEl = document.getElementById('tempValue');
        const bpEl = document.getElementById('bpValue');

        if (hrEl) hrEl.textContent = hrCount > 0 ? Math.round(hr / hrCount) : '--';
        if (hrvEl) hrvEl.textContent = Math.round(hrv);
        if (spo2El) spo2El.textContent = spo2Count > 0 ? Math.round(spo2 / spo2Count) : '--';
        if (tempEl) tempEl.textContent = tempCount > 0 ? (temp / tempCount).toFixed(1) : '--';
        if (bpEl) bpEl.textContent = bpCount > 0 ? Math.round(bp / bpCount) + '/80' : '--';

    } catch (error) {
        console.error('Failed to load biometric params:', error);
    }
}

async function classifyTraining() {
    const btn = document.getElementById('classifyBtn');
    if (!btn) return;

    btn.disabled = true;
    btn.textContent = '⏳ Анализ...';

    try {
        const response = await fetch('/api/v1/ml/classify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            }
        });

        const result = await response.json();
        displayResult(result);

    } catch (error) {
        alert('Ошибка классификации: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = '🔍 Классифицировать Тип Тренировки';
    }
}

function displayResult(result) {
    const resultCard = document.getElementById('resultCard');
    const resultContent = document.getElementById('resultContent');
    if (!resultCard || !resultContent) return;

    const classColors = {
        'recovery': 'type-recovery',
        'endurance_e1e2': 'type-endurance',
        'threshold_e3': 'type-threshold',
        'strength_hiit': 'type-hiit'
    };

    const classIcons = {
        'recovery': '🔵',
        'endurance_e1e2': '🟢',
        'threshold_e3': '🟡',
        'strength_hiit': '🔴'
    };

    resultCard.style.display = 'block';
    resultContent.innerHTML = `
        <div class="classification-result ${classColors[result.predicted_class] || ''}">
            <h3>${classIcons[result.predicted_class] || '📊'} ${result.predicted_class_ru}</h3>
            <p class="confidence">Уверенность: ${(result.confidence * 100).toFixed(1)}%</p>
            <p class="description">${result.description}</p>
            <p class="hr-range">Зона пульса: ${result.hr_range}</p>
            <div class="recommendations">
                <h4>Рекомендации:</h4>
                <ul>
                    ${result.recommendations.map(r => `<li>${r}</li>`).join('')}
                </ul>
            </div>
            ${result.personalized_notes ? `
                <div class="personalized-notes">
                    <strong>⚠️ Персональные заметки:</strong>
                    <p>${result.personalized_notes}</p>
                </div>
            ` : ''}
            <button onclick="generatePlanFromClassification('${result.predicted_class}')" class="btn-primary">
                📋 Сгенерировать План Тренировки
            </button>
        </div>
        <div class="probabilities">
            <h4>Вероятности классов:</h4>
            ${Object.entries(result.probabilities).map(([cls, prob]) => `
                <div class="prob-bar">
                    <span class="prob-label">${cls}</span>
                    <div class="prob-fill" style="width: ${prob * 100}%"></div>
                    <span class="prob-value">${(prob * 100).toFixed(1)}%</span>
                </div>
            `).join('')}
        </div>
    `;

    resultCard.scrollIntoView({ behavior: 'smooth' });
}

async function generatePlanFromClassification(trainingClass) {
    try {
        const response = await fetch('/api/v1/ml/generate-plan', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                training_class: trainingClass,
                duration_weeks: 4,
                available_days: [1, 3, 5],
                preferences: {
                    max_duration: 60,
                    available_equipment: [],
                    preferred_time: 'evening'
                }
            })
        });

        await response.json();
        alert('План сгенерирован! Переходим к тренировкам...');
        window.location.href = '/training';

    } catch (error) {
        alert('Ошибка генерации плана: ' + error.message);
    }
}

// ===================== ML Plan Generation =====================

function getDayName(day) {
    const days = {
        'monday': 'Пн', 'tuesday': 'Вт', 'wednesday': 'Ср',
        'thursday': 'Чт', 'friday': 'Пт', 'saturday': 'Сб', 'sunday': 'Вс'
    };
    return days[day] || day;
}

async function generatePlan(e) {
    e.preventDefault();

    const form = e.target;
    const btn = form.querySelector('button[type="submit"]');
    if (btn) {
        btn.disabled = true;
        btn.textContent = '⏳ Генерация...';
    }

    const days = Array.from(document.querySelectorAll('input[name="days"]:checked'))
        .map(cb => parseInt(cb.value));

    const equipment = Array.from(document.querySelectorAll('input[name="equipment"]:checked'))
        .map(cb => cb.value)
        .filter(v => v !== 'none');

    const requestData = {
        training_class: document.getElementById('trainingClass').value,
        duration_weeks: parseInt(document.getElementById('durationWeeks').value),
        available_days: days,
        preferences: {
            max_duration: parseInt(document.getElementById('maxDuration').value),
            available_equipment: equipment,
            preferred_time: document.getElementById('preferredTime').value
        }
    };

    try {
        const response = await fetch('/api/v1/ml/generate-plan', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(requestData)
        });

        const plan = await response.json();
        displayPlan(plan);

    } catch (error) {
        alert('Ошибка генерации: ' + error.message);
    } finally {
        if (btn) {
            btn.disabled = false;
            btn.textContent = '✨ Сгенерировать План';
        }
    }
}

function displayPlan(plan) {
    const planResult = document.getElementById('planResult');
    const planContent = document.getElementById('planContent');
    if (!planResult || !planContent) return;

    planResult.style.display = 'block';
    planContent.innerHTML = `
        <div class="plan-preview">
            <h3>${plan.training_type_ru || plan.training_type}</h3>
            <div class="plan-stats">
                <span>⏱️ ${plan.duration_minutes} мин</span>
                <span>📊 Интенсивность: ${(plan.intensity * 100).toFixed(0)}%</span>
                <span>📅 ${plan.weekly_frequency} раз/неделю</span>
            </div>
            <h4>Структура Тренировки:</h4>
            <ul class="exercise-list">
                ${plan.session_structure?.map(ex => `
                    <li>
                        <strong>${ex.name}</strong> - ${ex.duration_minutes} мин
                        (интенсивность: ${(ex.intensity * 100).toFixed(0)}%)
                    </li>
                `).join('') || '<li>Стандартная структура тренировки</li>'}
            </ul>
            <h4>Упражнения:</h4>
            <p>${plan.exercises?.join(', ') || 'Бег, велосипед, плавание'}</p>
            <h4>Расписание на неделю:</h4>
            <div class="weekly-schedule">
                ${Object.entries(plan.weekly_schedule || {}).map(([day, activity]) => `
                    <div class="day-card ${activity === 'rest' ? 'rest' : ''}">
                        <strong>${getDayName(day)}</strong><br>
                        ${activity === 'rest' ? 'Отдых' : activity}
                    </div>
                `).join('')}
            </div>
            ${plan.notes?.length > 0 ? `
                <div class="personalized-notes">
                    <h4>⚠️ Персональные Рекомендации:</h4>
                    <ul>
                        ${plan.notes.map(note => `<li>${note}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}
            <button onclick="savePlan()" class="btn-primary">
                💾 Сохранить План
            </button>
        </div>
    `;

    planResult.scrollIntoView({ behavior: 'smooth' });
}

async function savePlan() {
    alert('План сохранён! Переходим к тренировкам...');
    window.location.href = '/training';
}
