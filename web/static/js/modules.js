// FitPulse New Modules — Doctor, Devices, Diet, Training
// Mobile web app UI logic

const AppModules = (() => {
    // ===== State =====
    let currentUser = null;
    let selectedDoctor = null;
    let selectedDevice = null;

    // ===== Device Module =====
    const DeviceModule = {
        devices: [
            { type: 'apple_watch', name: 'Apple Watch', icon: '⌚', capabilities: 'Пульс, ЭКГ, SpO₂, Сон' },
            { type: 'samsung_galaxy_watch', name: 'Samsung Galaxy Watch', icon: '⌚', capabilities: 'Пульс, ЭКГ, SpO₂, Температура' },
            { type: 'huawei_watch_d2', name: 'Huawei Watch D2', icon: '⌚', capabilities: 'Пульс, Давление, ЭКГ, SpO₂' },
            { type: 'amazfit_trex3', name: 'Amazfit T-Rex 3', icon: '⌚', capabilities: 'Пульс, SpO₂, Сон' }
        ],

        init() {
            this.renderDeviceSelector();
            this.bindEvents();
        },

        renderDeviceSelector() {
            const container = document.getElementById('deviceSelector');
            if (!container) return;

            container.innerHTML = this.devices.map(d => `
                <div class="device-option" data-type="${d.type}">
                    <div class="device-icon">${d.icon}</div>
                    <div class="device-name">${d.name}</div>
                    <div class="device-capabilities">${d.capabilities}</div>
                </div>
            `).join('');
        },

        bindEvents() {
            document.addEventListener('click', (e) => {
                const deviceOption = e.target.closest('.device-option');
                if (deviceOption) {
                    document.querySelectorAll('.device-option').forEach(el => el.classList.remove('selected'));
                    deviceOption.classList.add('selected');
                    selectedDevice = deviceOption.dataset.type;
                }
            });
        },

        async connectDevice(deviceType, userId) {
            try {
                const resp = await fetch('/api/v1/devices/register', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ device_type: deviceType, user_id: userId })
                });
                return await resp.json();
            } catch (err) {
                console.error('Device connection failed:', err);
                throw err;
            }
        }
    };

    // ===== Doctor Module =====
    const DoctorModule = {
        async loadDoctors() {
            try {
                const doctors = await window.doctorAPI.listDoctors();
                this.renderDoctorsList(doctors.doctors || []);
            } catch (err) {
                console.error('Failed to load doctors:', err);
            }
        },

        renderDoctorsList(doctors) {
            const container = document.getElementById('doctorsList');
            if (!container) return;

            if (doctors.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <div class="empty-state-icon">👨‍⚕️</div>
                        <div class="empty-state-text">Врачи пока недоступны</div>
                    </div>
                `;
                return;
            }

            container.innerHTML = doctors.map(d => `
                <div class="doctor-card" data-id="${d.id}">
                    <div class="doctor-avatar">👨‍⚕️</div>
                    <div class="doctor-info">
                        <div class="doctor-name">${d.full_name}</div>
                        <div class="doctor-specialty">${d.specialty}</div>
                    </div>
                    <div class="doctor-rating">
                        ⭐ ${d.rating}
                        <span>(${d.consultation_count})</span>
                    </div>
                </div>
            `).join('');
        },

        async openChat(doctorId) {
            selectedDoctor = doctorId;
            this.showChatView();
            this.loadChatHistory();
        },

        async showChatView() {
            const view = document.getElementById('doctorChatView');
            if (view) {
                view.classList.add('active');
            }
        },

        async loadChatHistory() {
            if (!currentUser || !selectedDoctor) return;

            try {
                const history = await window.doctorAPI.getChatHistory(currentUser.id, selectedDoctor);
                this.renderMessages(history.messages || []);
            } catch (err) {
                console.error('Failed to load chat:', err);
            }
        },

        renderMessages(messages) {
            const container = document.getElementById('chatMessages');
            if (!container) return;

            if (messages.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <div class="empty-state-icon">💬</div>
                        <div class="empty-state-text">Начните диалог с врачом</div>
                    </div>
                `;
                return;
            }

            container.innerHTML = messages.map(m => `
                <div class="chat-message ${m.sender_type}">
                    <div class="message-text">${m.message}</div>
                    <div class="chat-time">${new Date(m.created_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}</div>
                </div>
            `).join('');

            container.scrollTop = container.scrollHeight;
        },

        async sendMessage(text) {
            if (!currentUser || !selectedDoctor || !text.trim()) return;

            try {
                await window.doctorAPI.sendMessage(
                    currentUser.id,
                    selectedDoctor,
                    currentUser.id,
                    'user',
                    text
                );
                this.loadChatHistory();
            } catch (err) {
                console.error('Failed to send message:', err);
                this.showToast('Ошибка отправки сообщения', 'error');
            }
        },

        async loadPrescriptions() {
            if (!currentUser) return;

            try {
                const prescriptions = await window.doctorAPI.getPrescriptions(currentUser.id);
                this.renderPrescriptions(prescriptions.prescriptions || []);
            } catch (err) {
                console.error('Failed to load prescriptions:', err);
            }
        },

        renderPrescriptions(prescriptions) {
            const container = document.getElementById('prescriptionsList');
            if (!container) return;

            if (prescriptions.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <div class="empty-state-icon">📋</div>
                        <div class="empty-state-text">Нет назначений от врача</div>
                    </div>
                `;
                return;
            }

            container.innerHTML = prescriptions.map(p => `
                <div class="prescription-card priority-${p.priority}">
                    <div class="prescription-title">${p.title}</div>
                    <div class="prescription-desc">${p.description || 'Без описания'}</div>
                    <div class="prescription-meta">
                        <span>${p.prescription_type}</span>
                        <span>${new Date(p.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                </div>
            `).join('');
        }
    };

    // ===== Diet Module =====
    const DietModule = {
        async loadDietPlan() {
            const container = document.getElementById('dietPlanContainer');
            if (!container) return;

            try {
                // TODO: API call to get diet plan
                container.innerHTML = `
                    <div class="diet-summary">
                        <div class="diet-calories">2,200</div>
                        <div class="diet-label">калорий в день</div>
                        <div class="diet-macros">
                            <div class="macro-item">
                                <div class="macro-value">165г</div>
                                <div class="macro-label">Белки</div>
                            </div>
                            <div class="macro-item">
                                <div class="macro-value">247г</div>
                                <div class="macro-label">Углеводы</div>
                            </div>
                            <div class="macro-item">
                                <div class="macro-value">61г</div>
                                <div class="macro-label">Жиры</div>
                            </div>
                        </div>
                    </div>
                    <div class="meal-card">
                        <div class="meal-time">🌅 08:00</div>
                        <div class="meal-name">Овсянка с фруктами</div>
                        <div class="meal-details">350 ккал • 12г белка • 60г углеводов • 8г жиров</div>
                    </div>
                    <div class="meal-card">
                        <div class="meal-time">☀️ 13:00</div>
                        <div class="meal-name">Курица с рисом</div>
                        <div class="meal-details">550 ккал • 40г белка • 60г углеводов • 15г жиров</div>
                    </div>
                    <div class="meal-card">
                        <div class="meal-time">🌙 19:00</div>
                        <div class="meal-name">Индейка с овощами</div>
                        <div class="meal-details">400 ккал • 35г белка • 25г углеводов • 18г жиров</div>
                    </div>
                `;
            } catch (err) {
                console.error('Failed to load diet plan:', err);
            }
        }
    };

    // ===== Training Plan Module =====
    const TrainingModule = {
        async loadTrainingPlan() {
            const container = document.getElementById('trainingPlanContainer');
            if (!container) return;

            try {
                container.innerHTML = `
                    <div class="training-plan-card">
                        <div class="plan-day-header">
                            <div class="plan-day-name">Понедельник</div>
                            <div class="plan-day-type">Выносливость</div>
                        </div>
                        <div class="exercise-item">
                            <div class="exercise-number">1</div>
                            <div class="exercise-details">
                                <div class="exercise-name">Ходьба на беговой дорожке</div>
                                <div class="exercise-meta">10 мин • Разминка</div>
                            </div>
                        </div>
                        <div class="exercise-item">
                            <div class="exercise-number">2</div>
                            <div class="exercise-details">
                                <div class="exercise-name">Жим лёжа</div>
                                <div class="exercise-meta">4×10 • 90 сек отдых</div>
                            </div>
                        </div>
                        <div class="exercise-item">
                            <div class="exercise-number">3</div>
                            <div class="exercise-details">
                                <div class="exercise-name">Становая тяга</div>
                                <div class="exercise-meta">4×8 • 120 сек отдых</div>
                            </div>
                        </div>
                        <div class="exercise-item">
                            <div class="exercise-number">4</div>
                            <div class="exercise-details">
                                <div class="exercise-name">Фоам-роллинг</div>
                                <div class="exercise-meta">10 мин • Заминка</div>
                            </div>
                        </div>
                    </div>
                `;
            } catch (err) {
                console.error('Failed to load training plan:', err);
            }
        }
    };

    // ===== Toast Notifications =====
    function showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, 3000);
    }

    // ===== Init =====
    function init(user) {
        currentUser = user;
        DeviceModule.init();
        DoctorModule.loadDoctors();
        DoctorModule.loadPrescriptions();
        DietModule.loadDietPlan();
        TrainingModule.loadTrainingPlan();
    }

    return {
        init,
        DeviceModule,
        DoctorModule,
        DietModule,
        TrainingModule,
        showToast
    };
})();

// Export to global scope
window.AppModules = AppModules;
