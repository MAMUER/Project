document.addEventListener('DOMContentLoaded', () => {
    console.log('[APP] Loaded');

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
    const verifyForm = document.getElementById('verifyForm');
    const pageTitle = document.getElementById('pageTitle');

    const viewTitles = {
        dashboard: 'Обзор',
        profile: 'Профиль',
        training: 'Тренировки',
        devices: 'Устройства',
        doctor: 'Врач',
        diet: 'Диета',
        ml: 'AI Анализ',
    };

    // ===== Init =====
    function init() {
        console.log('[APP] Init, authToken:', authToken ? 'present' : 'null');

        // Проверяем токен подтверждения в URL (из письма)
        const urlParams = new URLSearchParams(window.location.search);
        const confirmToken = urlParams.get('token');
        if (confirmToken) {
            console.log('[APP] Found confirm token in URL, auto-confirming...');
            showAuthScreen();
            // Показываем форму подтверждения с токеном
            loginForm.classList.add('hidden');
            registerForm.classList.add('hidden');
            if (verifyForm) {
                verifyForm.classList.remove('hidden');
                document.getElementById('verifyToken').value = confirmToken;
                // Автоматически подтверждаем
                autoConfirmEmail(confirmToken);
            }
            return;
        }

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
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
        if (verifyForm) verifyForm.classList.add('hidden');
        clearErrors();
        console.log('[APP] Auth screen shown');
    }

    function showVerification(email, message, userId) {
        console.log('[APP] Show verification for:', email);
        loginForm.classList.add('hidden');
        registerForm.classList.add('hidden');
        if (verifyForm) verifyForm.classList.remove('hidden');

        document.getElementById('verifyEmail').textContent = email;

        // Show dev token if present in message
        const tokenMatch = message.match(/token \(dev only\):\s*([a-f0-9]+)/i);
        const devSection = document.getElementById('devTokenSection');
        if (tokenMatch && devSection) {
            devSection.classList.remove('hidden');
            document.getElementById('devToken').textContent = tokenMatch[1];
        } else if (devSection) {
            devSection.classList.add('hidden');
        }

        // Reset confirm state
        const confirmErr = document.getElementById('confirmError');
        const confirmOk = document.getElementById('confirmSuccess');
        if (confirmErr) { confirmErr.textContent = ''; confirmErr.classList.add('hidden'); }
        if (confirmOk) confirmOk.classList.add('hidden');
        const tokenInput = document.getElementById('verifyToken');
        if (tokenInput) tokenInput.value = '';
    }

    function copyToken() {
        const token = document.getElementById('devToken')?.textContent;
        if (token) {
            navigator.clipboard.writeText(token).then(() => {
                showToast('Токен скопирован!', 'success');
            });
        }
    }
    window.copyToken = copyToken;

    async function confirmEmail(token) {
        console.log('[AUTH] Confirming email with token:', token);
        const response = await fetch('/api/v1/auth/confirm', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token })
        });

        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.message || data.error || 'Ошибка подтверждения');
        }
        return data;
    }

    async function autoConfirmEmail(token) {
        const confirmErr = document.getElementById('confirmError');
        const confirmOk = document.getElementById('confirmSuccess');
        const btn = document.getElementById('confirmBtn');

        if (btn) { btn.disabled = true; btn.textContent = 'Подтверждение...'; }
        if (confirmErr) confirmErr.classList.add('hidden');
        if (confirmOk) confirmOk.classList.add('hidden');

        try {
            await confirmEmail(token);
            if (confirmOk) confirmOk.classList.remove('hidden');
            showToast('Email подтверждён! Переход ко входу...', 'success');
            setTimeout(() => {
                loginForm.classList.remove('hidden');
                if (verifyForm) verifyForm.classList.add('hidden');
            }, 2000);
        } catch (err) {
            console.error('[AUTH] Auto-confirm failed:', err);
            if (confirmErr) {
                confirmErr.textContent = 'Ошибка: ' + err.message;
                confirmErr.classList.remove('hidden');
            }
        } finally {
            if (btn) { btn.disabled = false; btn.textContent = 'Подтвердить email'; }
        }
    }

    function showMainApp() {
        authScreen.classList.remove('active');
        mainScreen.classList.add('active');
        mainScreen.classList.remove('hidden');
        switchView('dashboard');
        console.log('[APP] Main app shown');
    }

    // ===== Validation =====
    const validators = {
        email: (v) => {
            if (!v) return 'Введите email';
            if (v.length > 254) return 'Email слишком длинный';
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
            if (v.length < 2) return 'Минимум 2 символа';
            if (v.length > 100) return 'Максимум 100 символов';
            if (!/^[A-Za-zА-Яа-яЁё\s\-]+$/.test(v)) return 'Только буквы';
            return '';
        },
    };

    function setFieldError(input, errorEl, msg) {
        if (!input) return;
        input.classList.toggle('invalid', !!msg);
        input.classList.toggle('valid', !msg && input.value.length > 0);
        if (errorEl) errorEl.textContent = msg || '';
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
        // --- Fields ---
        const loginEmail = document.getElementById('loginEmail');
        const loginPassword = document.getElementById('loginPassword');
        const loginEmailErr = document.getElementById('loginEmailError');
        const loginPassErr = document.getElementById('loginPasswordError');
        const loginErrEl = document.getElementById('loginError');

        const regName = document.getElementById('regName');
        const regEmail = document.getElementById('regEmail');
        const regPassword = document.getElementById('regPassword');
        const regNameErr = document.getElementById('regNameError');
        const regEmailErr = document.getElementById('regEmailError');
        const regPassErr = document.getElementById('regPasswordError');
        const regErrEl = document.getElementById('registerError');

        console.log('[APP] Elements:', { loginForm: !!loginForm, registerForm: !!registerForm, loginEmail: !!loginEmail });

        // Login field validation
        if (loginEmail) loginEmail.addEventListener('input', () => {
            setFieldError(loginEmail, loginEmailErr, validators.email(loginEmail.value));
            if (loginErrEl) loginErrEl.classList.add('hidden');
        });
        if (loginPassword) loginPassword.addEventListener('input', () => {
            setFieldError(loginPassword, loginPassErr, validators.loginPassword(loginPassword.value));
            if (loginErrEl) loginErrEl.classList.add('hidden');
        });

        // Register field validation
        if (regName) regName.addEventListener('input', () => {
            setFieldError(regName, regNameErr, validators.name(regName.value));
            if (regErrEl) regErrEl.classList.add('hidden');
        });
        if (regEmail) regEmail.addEventListener('input', () => {
            setFieldError(regEmail, regEmailErr, validators.email(regEmail.value));
            if (regErrEl) regErrEl.classList.add('hidden');
        });
        if (regPassword) regPassword.addEventListener('input', () => {
            const result = validators.password(regPassword.value);
            const err = typeof result === 'object' ? result.error : result;
            setFieldError(regPassword, regPassErr, err);
            updatePasswordHint(typeof result === 'object' ? result : null);
            if (regErrEl) regErrEl.classList.add('hidden');
        });

        // Auth toggle
        document.getElementById('toRegister')?.addEventListener('click', e => {
            e.preventDefault();
            console.log('[APP] Switch to register');
            loginForm.classList.add('hidden');
            registerForm.classList.remove('hidden');
            clearErrors();
        });
        document.getElementById('toLogin')?.addEventListener('click', e => {
            e.preventDefault();
            console.log('[APP] Switch to login');
            registerForm.classList.add('hidden');
            loginForm.classList.remove('hidden');
            if (verifyForm) verifyForm.classList.add('hidden');
            clearErrors();
        });

        // Back to login from verify screen
        document.getElementById('backToLogin')?.addEventListener('click', e => {
            e.preventDefault();
            console.log('[APP] Back to login from verify');
            loginForm.classList.remove('hidden');
            if (verifyForm) verifyForm.classList.add('hidden');
            clearErrors();
        });

        // Confirm email
        document.getElementById('confirmBtn')?.addEventListener('click', async (e) => {
            e.preventDefault();
            const token = document.getElementById('verifyToken').value.trim();
            const confirmErr = document.getElementById('confirmError');
            const confirmOk = document.getElementById('confirmSuccess');
            const btn = document.getElementById('confirmBtn');

            if (!token) {
                if (confirmErr) { confirmErr.textContent = 'Вставьте токен'; confirmErr.classList.remove('hidden'); }
                return;
            }

            if (btn) { btn.disabled = true; btn.textContent = 'Подтверждение...'; }
            if (confirmErr) confirmErr.classList.add('hidden');
            if (confirmOk) confirmOk.classList.add('hidden');

            try {
                await confirmEmail(token);
                if (confirmOk) confirmOk.classList.remove('hidden');
                showToast('Email подтверждён! Теперь войдите.', 'success');
                setTimeout(() => {
                    loginForm.classList.remove('hidden');
                    if (verifyForm) verifyForm.classList.add('hidden');
                }, 2000);
            } catch (err) {
                if (confirmErr) { confirmErr.textContent = err.message; confirmErr.classList.remove('hidden'); }
            } finally {
                if (btn) { btn.disabled = false; btn.textContent = 'Подтвердить email'; }
            }
        });

        // ===== LOGIN SUBMIT =====
        if (loginForm) {
            loginForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                console.log('[LOGIN] Submit!');
                const email = loginEmail.value.trim();
                const password = loginPassword.value;

                const emailErr = validators.email(email);
                const passErr = validators.loginPassword(password);
                setFieldError(loginEmail, loginEmailErr, emailErr);
                setFieldError(loginPassword, loginPassErr, passErr);

                if (emailErr || passErr) {
                    console.log('[LOGIN] Validation failed');
                    if (loginErrEl) { loginErrEl.textContent = 'Проверьте введённые данные'; loginErrEl.classList.remove('hidden'); }
                    return;
                }

                const btn = document.getElementById('loginBtn');
                if (btn) { btn.disabled = true; btn.textContent = 'Вход...'; }

                try {
                    console.log('[LOGIN] Calling API for:', email);
                    const data = await login(email, password);
                    console.log('[LOGIN] Got response type:', typeof data, 'keys:', data ? Object.keys(data) : 'null');
                    if (data && typeof data === 'object' && data.access_token) {
                        setAuthToken(data.access_token);
                        showMainApp();
                    } else {
                        console.error('[LOGIN] Unexpected response:', data);
                        const msg = data && data.message ? data.message : 'Сервер вернул неожиданный ответ. Попробуйте войти позже.';
                        throw new Error(msg);
                    }
                } catch (err) {
                    console.error('[LOGIN] Error:', err);
                    if (loginErrEl) { loginErrEl.textContent = err.message; loginErrEl.classList.remove('hidden'); }
                } finally {
                    if (btn) { btn.disabled = false; btn.textContent = 'Войти'; }
                }
            });
        }

        // ===== REGISTER SUBMIT =====
        if (registerForm) {
            registerForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                console.log('[REGISTER] Submit!');
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
                    console.log('[REGISTER] Validation failed');
                    if (regErrEl) { regErrEl.textContent = 'Проверьте введённые данные'; regErrEl.classList.remove('hidden'); }
                    return;
                }

                const btn = document.getElementById('registerBtn');
                if (btn) { btn.disabled = true; btn.textContent = 'Создание...'; }

                try {
                    console.log('[REGISTER] Calling API for:', email, name);
                    const data = await register(email, password, name);
                    console.log('[REGISTER] Got response:', data);

                    // Show verification screen
                    showVerification(email, data.message || '', data.user_id);
                } catch (err) {
                    console.error('[REGISTER] Error:', err);
                    if (regErrEl) { regErrEl.textContent = err.message; regErrEl.classList.remove('hidden'); regErrEl.style.color = ''; }
                } finally {
                    if (btn) { btn.disabled = false; btn.textContent = 'Создать аккаунт'; }
                }
            });
        }

        // Logout
        document.getElementById('logoutBtn')?.addEventListener('click', async () => {
            await logout();
            setAuthToken(null);
            showAuthScreen();
        });

        // Tab bar
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => switchView(tab.dataset.view));
        });

        // Generate plan
        document.getElementById('generatePlanBtn')?.addEventListener('click', generatePlan);
        document.getElementById('dashGenerateBtn')?.addEventListener('click', generatePlan);

        // ML classify
        document.getElementById('mlClassifyBtn')?.addEventListener('click', mlClassify);

        // Profile save
        document.getElementById('profileForm')?.addEventListener('submit', saveProfile);
    }

    function clearErrors() {
        ['loginError', 'registerError', 'authError'].forEach(id => {
            const el = document.getElementById(id);
            if (el) { el.textContent = ''; el.classList.add('hidden'); el.style.color = ''; }
        });
        ['loginEmailError', 'loginPasswordError', 'regNameError', 'regEmailError', 'regPasswordError'].forEach(id => {
            const el = document.getElementById(id);
            if (el) el.textContent = '';
        });
    }

    // ===== Navigation =====
    function switchView(viewName) {
        state.currentView = viewName;
        document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
        const targetView = document.getElementById(`${viewName}View`);
        if (targetView) targetView.classList.add('active');

        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        const activeTab = document.querySelector(`.tab[data-view="${viewName}"]`);
        if (activeTab) activeTab.classList.add('active');

        pageTitle.textContent = viewTitles[viewName] || 'FitPulse';

        if (viewName === 'dashboard') loadDashboard();
        if (viewName === 'profile') loadProfile();
        if (viewName === 'training') loadTrainingPlans();
        if (viewName === 'ml') loadMLView();
        if (viewName === 'devices') initDevicesView();
        if (viewName === 'doctor') initDoctorView();
        if (viewName === 'diet') initDietView();
    }

    // ===== Dashboard =====
    async function loadDashboard() {
        try {
            const [hrData, spo2Data] = await Promise.allSettled([
                getBiometricRecords('heart_rate', null, null, 10),
                getBiometricRecords('spo2', null, null, 5),
            ]);
            if (hrData.status === 'fulfilled' && hrData.value.records?.length > 0) {
                document.getElementById('hrValue').textContent = Math.round(hrData.value.records[0].value);
            }
            if (spo2Data.status === 'fulfilled' && spo2Data.value.records?.length > 0) {
                document.getElementById('spo2Value').textContent = Math.round(spo2Data.value.records[0].value);
            }

            // Chart
            if (hrData.status === 'fulfilled' && hrData.value.records?.length > 1) {
                const records = hrData.value.records.slice(0, 20).reverse();
                const labels = records.map(r => new Date(r.timestamp).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }));
                const values = records.map(r => r.value);

                if (state.heartChart) state.heartChart.destroy();
                const ctx = document.getElementById('heartChart')?.getContext('2d');
                if (ctx) {
                    state.heartChart = new Chart(ctx, {
                        type: 'line',
                        data: {
                            labels,
                            datasets: [{
                                data: values, borderColor: '#ff375f', backgroundColor: 'rgba(255,55,95,0.1)',
                                fill: true, tension: 0.4, pointRadius: 0, borderWidth: 2.5,
                            }]
                        },
                        options: {
                            responsive: true, maintainAspectRatio: false,
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
                    return `<div class="plan-card"><h4>📋 Программа от ${date}</h4>
                        <div class="plan-meta"><span>Статус: ${plan.status}</span><span>${plan.classification_class || '—'}</span></div></div>`;
                }).join('');
            } else {
                container.innerHTML = `<div class="empty-state"><div class="empty-icon">🏃</div>
                    <h3>Нет активных программ</h3><p>AI создаст персональный план</p></div>`;
            }
        } catch (err) { console.error('Training plans load failed:', err); }
    }

    async function generatePlan() {
        try {
            showToast('Генерация плана...', 'success');
            const result = await apiRequest('/ml/generate-plan', {
                method: 'POST',
                body: JSON.stringify({ training_class: 'endurance_e1e2', duration_weeks: 4, available_days: [1, 3, 5], preferences: { max_duration: 60 } })
            });
            if (result.training_type) {
                showToast(`✅ ${result.training_type_ru || result.training_type}`, 'success');
                if (state.currentView === 'training') loadTrainingPlans();
                if (state.currentView === 'dashboard') {
                    const el = document.getElementById('todayWorkout');
                    if (el) {
                        const exercises = result.exercises || ['Разминка 10 мин', 'Основная часть 30 мин', 'Заминка 10 мин'];
                        el.innerHTML = `<h4 style="margin-bottom:12px;font-size:17px;">${result.training_type_ru || 'Тренировка'}</h4>
                            <div style="color:var(--text-secondary);font-size:14px;margin-bottom:12px;">${result.duration_minutes || 45} мин</div>
                            ${exercises.map(ex => `<div class="workout-item"><span class="workout-exercise">${ex}</span></div>`).join('')}`;
                    }
                }
            }
        } catch (err) { showToast('Ошибка: ' + err.message, 'error'); }
    }

    // ===== ML =====
    async function loadMLView() {}

    async function mlClassify() {
        try {
            const container = document.getElementById('mlResult');
            container.innerHTML = '<div style="text-align:center;padding:40px;color:var(--text-secondary);">Анализ...</div>';
            const result = await apiRequest('/ml/classify', { method: 'POST', body: '{}' });
            const classRu = result.predicted_class_ru || result.predicted_class || 'Не определено';
            const confidence = result.confidence ? Math.round(result.confidence * 100) : 0;
            container.innerHTML = `<div class="ml-classification">
                <div class="class-label">Ваше состояние</div>
                <div class="class-name">${classRu}</div>
                <div class="confidence">Уверенность: ${confidence}%</div>
                ${result.description ? `<p style="margin-top:12px;font-size:15px;color:var(--text-secondary);">${result.description}</p>` : ''}</div>`;
        } catch (err) {
            document.getElementById('mlResult').innerHTML = `<div class="empty-state"><div class="empty-icon">⚠️</div>
                <h3>Не удалось проанализировать</h3><p>${err.message}</p></div>`;
        }
    }

    // ===== Devices View =====
    function initDevicesView() {
        if (window.AppModules) {
            window.AppModules.DeviceModule.init();
        }
    }

    // ===== Doctor View =====
    function initDoctorView() {
        if (window.AppModules) {
            window.AppModules.DoctorModule.loadDoctors();
            window.AppModules.DoctorModule.loadPrescriptions();
            bindDoctorTabs();
        }
    }

    // ===== Diet View =====
    function initDietView() {
        if (window.AppModules) {
            window.AppModules.DietModule.loadDietPlan();
        }
    }

    // ===== Doctor Tabs =====
    function bindDoctorTabs() {
        document.querySelectorAll('#doctorTabs .tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('#doctorTabs .tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                const target = tab.dataset.tab;
                document.querySelectorAll('.doctor-tab-content').forEach(c => c.classList.remove('active'));
                document.getElementById(target)?.classList.add('active');
            });
        });

        // Chat send button
        document.getElementById('chatSendBtn')?.addEventListener('click', () => {
            const input = document.getElementById('chatInput');
            if (input && input.value.trim() && window.AppModules) {
                window.AppModules.DoctorModule.sendMessage(input.value.trim());
                input.value = '';
            }
        });

        document.getElementById('chatInput')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && e.target.value.trim() && window.AppModules) {
                window.AppModules.DoctorModule.sendMessage(e.target.value.trim());
                e.target.value = '';
            }
        });
    }

    // ===== Toast =====
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
