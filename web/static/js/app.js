document.addEventListener('DOMContentLoaded', () => {
    // ===== State =====
    const state = {
        currentView: 'dashboard',
        heartChart: null,
    };

    // ===== DOM Elements =====
    const authScreen = document.getElementById('authScreen');
    const mainScreen = document.getElementById('mainScreen');
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const authError = document.getElementById('authError');
    const pageTitle = document.getElementById('pageTitle');

    const viewTitles = {
        dashboard: 'Обзор',
        profile: 'Профиль',
        training: 'Тренировки',
        devices: 'Устройства',
        ml: 'AI Анализ',
    };

    // ===== Init =====
    function init() {
        if (authToken) {
            showMainApp();
        } else {
            showAuthScreen();
        }

        bindEvents();
    }

    function showAuthScreen() {
        authScreen.classList.add('active');
        mainScreen.classList.remove('active');
        mainScreen.classList.add('hidden');
    }

    function showMainApp() {
        authScreen.classList.remove('active');
        mainScreen.classList.add('active');
        mainScreen.classList.remove('hidden');
        switchView('dashboard');
    }

    // ===== Validation =====
    const validators = {
        email: (v) => {
            if (!v) return 'Введите email';
            if (v.length > 254) return 'Email слишком длинный (макс. 254 символа)';
            if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return 'Некорректный формат email';
            return '';
        },
        loginPassword: (v) => {
            if (!v) return 'Введите пароль';
            return '';
        },
        password: (v) => {
            const checks = {
                length: v.length >= 8,
                upper: /[A-ZА-ЯЁ]/.test(v),
                lower: /[a-zа-яё]/.test(v),
                digit: /\d/.test(v),
            };
            if (!v) return { error: 'Введите пароль', checks };
            if (!checks.length) return { error: 'Минимум 8 символов', checks };
            return { error: '', checks };
        },
        name: (v) => {
            if (!v) return 'Введите имя';
            if (v.length < 2) return 'Имя слишком короткое (мин. 2 символа)';
            if (v.length > 100) return 'Имя слишком длинное (макс. 100 символов)';
            if (!/^[A-Za-zА-Яа-яЁё\s\-]+$/.test(v)) return 'Имя может содержать только буквы';
            return '';
        },
    };

    function setFieldError(input, errorEl, msg) {
        input.classList.toggle('invalid', !!msg);
        input.classList.toggle('valid', !msg && input.value.length > 0);
        if (errorEl) errorEl.textContent = msg;
    }

    function updatePasswordHint(result) {
        const hint = document.getElementById('passwordHint');
        if (!hint) return;
        hint.classList.toggle('hidden', !result || !result.checks);
        if (!result) return;

        const items = {
            hintLength: result.checks.length,
            hintUpper: result.checks.upper,
            hintLower: result.checks.lower,
            hintDigit: result.checks.digit,
        };

        for (const [id, pass] of Object.entries(items)) {
            const el = document.getElementById(id);
            if (el) {
                el.classList.toggle('pass', pass);
                el.textContent = (pass ? '✓ ' : '✗ ') + el.textContent.slice(2);
            }
        }

        const btn = document.getElementById('registerBtn');
        if (btn) btn.disabled = !!result.error;
    }

    // ===== Events =====
    function bindEvents() {
        // --- Login validation ---
        const loginEmail = document.getElementById('loginEmail');
        const loginPassword = document.getElementById('loginPassword');
        const loginEmailErr = document.getElementById('loginEmailError');
        const loginPassErr = document.getElementById('loginPasswordError');
        const loginErr = document.getElementById('loginError');

        loginEmail?.addEventListener('input', () => {
            const err = validators.email(loginEmail.value);
            setFieldError(loginEmail, loginEmailErr, err);
            loginErr?.classList.add('hidden');
        });

        loginPassword?.addEventListener('input', () => {
            const err = validators.loginPassword(loginPassword.value);
            setFieldError(loginPassword, loginPassErr, err);
            loginErr?.classList.add('hidden');
        });

        // --- Registration validation ---
        const regName = document.getElementById('regName');
        const regEmail = document.getElementById('regEmail');
        const regPassword = document.getElementById('regPassword');
        const regNameErr = document.getElementById('regNameError');
        const regEmailErr = document.getElementById('regEmailError');
        const regPassErr = document.getElementById('regPasswordError');
        const regErr = document.getElementById('registerError');

        regName?.addEventListener('input', () => {
            const err = validators.name(regName.value);
            setFieldError(regName, regNameErr, err);
            regErr?.classList.add('hidden');
        });

        regEmail?.addEventListener('input', () => {
            const err = validators.email(regEmail.value);
            setFieldError(regEmail, regEmailErr, err);
            regErr?.classList.add('hidden');
        });

        regPassword?.addEventListener('input', () => {
            const result = validators.password(regPassword.value);
            const err = typeof result === 'object' ? result.error : result;
            setFieldError(regPassword, regPassErr, err);
            updatePasswordHint(typeof result === 'object' ? result : null);
            regErr?.classList.add('hidden');
        });

        // Auth toggle
        document.getElementById('toRegister')?.addEventListener('click', e => {
            e.preventDefault();
            loginForm.classList.add('hidden');
            registerForm.classList.remove('hidden');
            hideAllAuthErrors();
        });

        document.getElementById('toLogin')?.addEventListener('click', e => {
            e.preventDefault();
            registerForm.classList.add('hidden');
            loginForm.classList.remove('hidden');
            hideAllAuthErrors();
        });

        // Login submit
        loginForm.addEventListener('submit', async e => {
            e.preventDefault();
            const email = loginEmail.value.trim();
            const password = loginPassword.value;

            const emailErr = validators.email(email);
            const passErr = validators.loginPassword(password);
            setFieldError(loginEmail, loginEmailErr, emailErr);
            setFieldError(loginPassword, loginPassErr, passErr);

            if (emailErr || passErr) {
                if (loginErr) { loginErr.textContent = 'Проверьте введённые данные'; loginErr.classList.remove('hidden'); }
                return;
            }

            try {
                const data = await login(email, password);
                if (data.access_token) {
                    setAuthToken(data.access_token);
                    showMainApp();
                }
            } catch (err) {
                if (loginErr) { loginErr.textContent = err.message || 'Неверный email или пароль'; loginErr.classList.remove('hidden'); }
            }
        });

        // Register submit
        registerForm.addEventListener('submit', async e => {
            e.preventDefault();
            const name = regName.value.trim();
            const email = regEmail.value.trim();
            const password = regPassword.value;

            const nameErr = validators.name(name);
            const emailErr = validators.email(email);
            const passResult = validators.password(password);
            const passErr = typeof passResult === 'object' ? passResult.error : passResult;

            setFieldError(regName, regNameErr, nameErr);
            setFieldError(regEmail, regEmailErr, emailErr);
            setFieldError(regPassword, regPassErr, passErr);

            if (nameErr || emailErr || passErr) {
                if (regErr) { regErr.textContent = 'Проверьте введённые данные'; regErr.classList.remove('hidden'); }
                return;
            }

            try {
                await register(email, password, name);
                if (regErr) { regErr.textContent = '✅ Аккаунт создан! Теперь войдите.'; regErr.classList.remove('hidden'); }
                registerForm.classList.add('hidden');
                loginForm.classList.remove('hidden');
                // Reset form
                regName.value = '';
                regEmail.value = '';
                regPassword.value = '';
                [regName, regEmail, regPassword].forEach(el => {
                    el.classList.remove('valid', 'invalid');
                });
                updatePasswordHint(null);
            } catch (err) {
                if (regErr) { regErr.textContent = err.message || 'Ошибка регистрации'; regErr.classList.remove('hidden'); }
            }
        });

        // Logout
        document.getElementById('logoutBtn')?.addEventListener('click', async () => {
            await logout();
            setAuthToken(null);
            showAuthScreen();
        });

        // Tab bar navigation
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => {
                const view = tab.dataset.view;
                switchView(view);
            });
        });

        // Generate plan buttons
        document.getElementById('generatePlanBtn')?.addEventListener('click', generatePlan);
        document.getElementById('dashGenerateBtn')?.addEventListener('click', generatePlan);

        // ML classify
        document.getElementById('mlClassifyBtn')?.addEventListener('click', mlClassify);

        // Profile save
        document.getElementById('profileForm')?.addEventListener('submit', saveProfile);
    }

    // ===== Navigation =====
    function switchView(viewName) {
        state.currentView = viewName;

        // Update views
        document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
        const targetView = document.getElementById(`${viewName}View`);
        if (targetView) targetView.classList.add('active');

        // Update tabs
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        const activeTab = document.querySelector(`.tab[data-view="${viewName}"]`);
        if (activeTab) activeTab.classList.add('active');

        // Update title
        pageTitle.textContent = viewTitles[viewName] || 'FitPulse';

        // Load data
        if (viewName === 'dashboard') loadDashboard();
        if (viewName === 'profile') loadProfile();
        if (viewName === 'training') loadTrainingPlans();
        if (viewName === 'ml') loadMLView();
    }

    // ===== Dashboard =====
    async function loadDashboard() {
        try {
            // Load biometrics in parallel
            const [hrData, spo2Data] = await Promise.allSettled([
                getBiometricRecords('heart_rate', null, null, 10),
                getBiometricRecords('spo2', null, null, 5),
            ]);

            // Heart Rate
            if (hrData.status === 'fulfilled' && hrData.value.records?.length > 0) {
                const latest = hrData.value.records[0];
                document.getElementById('hrValue').textContent = Math.round(latest.value);
            }

            // SpO2
            if (spo2Data.status === 'fulfilled' && spo2Data.value.records?.length > 0) {
                const latest = spo2Data.value.records[0];
                document.getElementById('spo2Value').textContent = Math.round(latest.value);
            }

            // Chart
            if (hrData.status === 'fulfilled' && hrData.value.records?.length > 1) {
                const records = hrData.value.records.slice(0, 20).reverse();
                const labels = records.map(r => {
                    const d = new Date(r.timestamp);
                    return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
                });
                const values = records.map(r => r.value);

                if (state.heartChart) state.heartChart.destroy();
                const ctx = document.getElementById('heartChart')?.getContext('2d');
                if (ctx) {
                    state.heartChart = new Chart(ctx, {
                        type: 'line',
                        data: {
                            labels,
                            datasets: [{
                                data: values,
                                borderColor: '#ff375f',
                                backgroundColor: 'rgba(255, 55, 95, 0.1)',
                                fill: true,
                                tension: 0.4,
                                pointRadius: 0,
                                borderWidth: 2.5,
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: { legend: { display: false } },
                            scales: {
                                x: { display: true, grid: { display: false }, ticks: { color: '#636366', maxTicksLimit: 6, font: { size: 11 } } },
                                y: { display: true, grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#636366', font: { size: 11 } } }
                            }
                        }
                    });
                }
            }

            // AI recommendation
            try {
                const classifyRes = await apiRequest('/ml/classify', { method: 'POST', body: '{}' });
                if (classifyRes.predicted_class_ru) {
                    document.getElementById('aiRecommendation').textContent = classifyRes.predicted_class_ru;
                    document.getElementById('aiDescription').textContent =
                        `Уверенность: ${Math.round(classifyRes.confidence * 100)}% | ${classifyRes.description || ''}`;
                }
            } catch {
                document.getElementById('aiRecommendation').textContent = 'Нужно больше данных';
                document.getElementById('aiDescription').textContent = 'Добавьте биометрические данные для AI-анализа';
            }

        } catch (err) {
            console.error('Dashboard load failed:', err);
        }
    }

    // ===== Profile =====
    async function loadProfile() {
        try {
            const profile = await getProfile();
            const p = profile.profile || profile;

            document.getElementById('profName').value = p.full_name || '';
            document.getElementById('profAge').value = p.age || '';
            document.getElementById('profGender').value = p.gender || '';
            document.getElementById('profHeight').value = p.height_cm || '';
            document.getElementById('profWeight').value = p.weight_kg || '';
            document.getElementById('profFitness').value = p.fitness_level || '';
            document.getElementById('profNutrition').value = p.nutrition || '';
            document.getElementById('profSleep').value = p.sleep_hours || '';

            document.querySelectorAll('.goal-chip input[type="checkbox"]').forEach(cb => {
                cb.checked = p.goals?.includes(cb.value) || false;
            });
        } catch (err) {
            console.error('Profile load failed:', err);
        }
    }

    async function saveProfile(e) {
        e.preventDefault();
        const goals = Array.from(document.querySelectorAll('.goal-chip input:checked')).map(cb => cb.value);

        const data = {
            age: parseInt(document.getElementById('profAge').value) || null,
            gender: document.getElementById('profGender').value || null,
            height_cm: parseInt(document.getElementById('profHeight').value) || null,
            weight_kg: parseFloat(document.getElementById('profWeight').value) || null,
            fitness_level: document.getElementById('profFitness').value || null,
            nutrition: document.getElementById('profNutrition').value || null,
            sleep_hours: parseFloat(document.getElementById('profSleep').value) || null,
            goals,
        };

        try {
            await updateProfile(data);
            showToast('Профиль сохранён', 'success');
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
        }
    }

    // ===== Training =====
    async function loadTrainingPlans() {
        try {
            const data = await getTrainingPlans();
            const container = document.getElementById('plansList');
            const plans = data.plans || [];

            if (plans.length > 0) {
                container.innerHTML = plans.map(plan => {
                    const date = new Date(plan.generated_at).toLocaleDateString('ru-RU');
                    return `
                        <div class="plan-card">
                            <h4>📋 Программа от ${date}</h4>
                            <div class="plan-meta">
                                <span>Статус: ${plan.status}</span>
                                <span>${plan.classification_class || '—'}</span>
                            </div>
                        </div>
                    `;
                }).join('');
            } else {
                container.innerHTML = `
                    <div class="empty-state">
                        <div class="empty-icon">🏃</div>
                        <h3>Нет активных программ</h3>
                        <p>AI создаст персональный план на основе ваших данных</p>
                    </div>
                `;
            }
        } catch (err) {
            console.error('Training plans load failed:', err);
        }
    }

    async function generatePlan() {
        try {
            showToast('Генерация плана...', 'success');
            const result = await apiRequest('/ml/generate-plan', {
                method: 'POST',
                body: JSON.stringify({
                    training_class: 'endurance_e1e2',
                    duration_weeks: 4,
                    available_days: [1, 3, 5],
                    preferences: { max_duration: 60 }
                })
            });

            if (result.training_type) {
                showToast(`✅ ${result.training_type_ru || result.training_type}`, 'success');
                if (state.currentView === 'training') loadTrainingPlans();
                if (state.currentView === 'dashboard') {
                    const el = document.getElementById('todayWorkout');
                    if (el) {
                        const exercises = result.exercises || ['Разминка 10 мин', 'Основная часть 30 мин', 'Заминка 10 мин'];
                        el.innerHTML = `
                            <h4 style="margin-bottom:12px;font-size:17px;">${result.training_type_ru || 'Тренировка'}</h4>
                            <div style="color:var(--text-secondary);font-size:14px;margin-bottom:12px;">
                                ${result.duration_minutes || 45} мин · Интенсивность ${Math.round((result.intensity || 0.6) * 100)}%
                            </div>
                            ${exercises.map(ex => `
                                <div class="workout-item">
                                    <span class="workout-exercise">${ex}</span>
                                </div>
                            `).join('')}
                        `;
                    }
                }
            }
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
        }
    }

    // ===== ML Analysis =====
    async function loadMLView() {
        // Show last result if available
    }

    async function mlClassify() {
        try {
            const container = document.getElementById('mlResult');
            container.innerHTML = '<div style="text-align:center;padding:40px;color:var(--text-secondary);">Анализ...</div>';

            const result = await apiRequest('/ml/classify', { method: 'POST', body: '{}' });

            const classRu = result.predicted_class_ru || result.predicted_class || 'Не определено';
            const confidence = result.confidence ? Math.round(result.confidence * 100) : 0;

            container.innerHTML = `
                <div class="ml-classification">
                    <div class="class-label">Ваше состояние</div>
                    <div class="class-name">${classRu}</div>
                    <div class="confidence">Уверенность: ${confidence}%</div>
                    ${result.description ? `<p style="margin-top:12px;font-size:15px;color:var(--text-secondary);">${result.description}</p>` : ''}
                    ${result.recommendations?.length ? `
                        <div style="margin-top:16px;text-align:left;">
                            ${result.recommendations.map(r => `<p style="padding:8px 0;border-bottom:0.5px solid rgba(255,255,255,0.1);font-size:15px;">• ${r}</p>`).join('')}
                        </div>
                    ` : ''}
                </div>
            `;
        } catch (err) {
            document.getElementById('mlResult').innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">⚠️</div>
                    <h3>Не удалось проанализировать</h3>
                    <p>${err.message}</p>
                </div>
            `;
        }
    }

    // ===== Helpers =====
    function hideAllAuthErrors() {
        ['loginError', 'registerError', 'authError'].forEach(id => {
            const el = document.getElementById(id);
            if (el) { el.textContent = ''; el.classList.add('hidden'); }
        });
        ['loginEmailError', 'loginPasswordError', 'regNameError', 'regEmailError', 'regPasswordError'].forEach(id => {
            const el = document.getElementById(id);
            if (el) el.textContent = '';
        });
    }

    function showToast(msg, type = 'success') {
        const existing = document.querySelector('.toast');
        if (existing) existing.remove();

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = msg;
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), 3000);
    }

    // ===== Start =====
    init();
});
