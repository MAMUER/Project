document.addEventListener('DOMContentLoaded', () => {
    // Элементы DOM
    const authView = document.getElementById('authView');
    const dashboardView = document.getElementById('dashboardView');
    const profileView = document.getElementById('profileView');
    const trainingView = document.getElementById('trainingView');
    const achievementsView = document.getElementById('achievementsView');
    
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const showRegister = document.getElementById('showRegister');
    const showLogin = document.getElementById('showLogin');
    const registerCard = document.getElementById('registerCard');
    const logoutBtn = document.getElementById('logoutBtn');
    
    const navLinks = document.querySelectorAll('.nav-menu a');
    const menuToggle = document.getElementById('menuToggle');
    const navMenu = document.getElementById('navMenu');
    
    let heartChart = null;
    
    // Мобильное меню
    if (menuToggle) {
        menuToggle.addEventListener('click', () => {
            navMenu.classList.toggle('active');
        });
    }
    
    // Навигация
    function showView(viewName) {
        authView.classList.add('hidden');
        dashboardView.classList.add('hidden');
        profileView.classList.add('hidden');
        trainingView.classList.add('hidden');
        achievementsView.classList.add('hidden');
        
        const activeView = document.getElementById(`${viewName}View`);
        if (activeView) activeView.classList.remove('hidden');
        
        // Обновляем активную ссылку
        navLinks.forEach(link => {
            const href = link.getAttribute('href');
            if (href === `/${viewName}` || (viewName === 'dashboard' && href === '/')) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
        
        // Загружаем данные для конкретного представления
        if (viewName === 'dashboard') loadDashboard();
        if (viewName === 'profile') loadProfile();
        if (viewName === 'training') loadTrainingPlans();
        if (viewName === 'achievements') loadAchievements();
    }
    
    // Переключение между входом и регистрацией
    if (showRegister) {
        showRegister.addEventListener('click', (e) => {
            e.preventDefault();
            document.querySelector('.auth-card').classList.add('hidden');
            registerCard.classList.remove('hidden');
        });
    }
    
    if (showLogin) {
        showLogin.addEventListener('click', (e) => {
            e.preventDefault();
            registerCard.classList.add('hidden');
            document.querySelector('.auth-card').classList.remove('hidden');
        });
    }
    
    // Вход
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = document.getElementById('loginEmail').value;
            const password = document.getElementById('loginPassword').value;
            
            try {
                const data = await login(email, password);
                if (data.access_token) {
                    showView('dashboard');
                }
            } catch (error) {
                alert('Ошибка входа: ' + error.message);
            }
        });
    }
    
    // Регистрация
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const fullName = document.getElementById('regFullName').value;
            const email = document.getElementById('regEmail').value;
            const password = document.getElementById('regPassword').value;
            
            try {
                const data = await register(email, password, fullName);
                if (data.user_id) {
                    alert('Регистрация успешна! Войдите в систему.');
                    registerCard.classList.add('hidden');
                    document.querySelector('.auth-card').classList.remove('hidden');
                }
            } catch (error) {
                alert('Ошибка регистрации: ' + error.message);
            }
        });
    }
    
    // Выход
    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => {
            setAuthToken(null);
            window.location.reload();
        });
    }
    
    // Загрузка дашборда
    async function loadDashboard() {
        try {
            const profile = await getProfile();
            document.getElementById('statsGrid').innerHTML = `
                <div class="stat-card"><span>Пульс</span>--</div>
                <div class="stat-card"><span>Давление</span>--/--</div>
                <div class="stat-card"><span>SpO2</span>--</div>
                <div class="stat-card"><span>Сон</span>-- ч</div>
            `;
            
            // Загружаем данные пульса для графика
            const heartData = await getBiometricRecords('heart_rate', null, null, 30);
            if (heartData.records && heartData.records.length > 0) {
                const labels = heartData.records.map(r => new Date(r.timestamp).toLocaleDateString());
                const values = heartData.records.map(r => r.value);
                
                if (heartChart) heartChart.destroy();
                const ctx = document.getElementById('heartChart').getContext('2d');
                heartChart = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: labels.reverse(),
                        datasets: [{
                            label: 'Пульс (уд/мин)',
                            data: values.reverse(),
                            borderColor: '#1abc9c',
                            tension: 0.3
                        }]
                    }
                });
            }
            
            document.getElementById('todayWorkout').innerHTML = `
                <h3>Разминка</h3>
                <p>10 минут легкой кардио</p>
                <h3>Основная часть</h3>
                <p>Приседания: 3x15</p>
                <p>Отжимания: 3x12</p>
                <p>Планка: 3x30 сек</p>
                <button onclick="completeWorkout('today')">Выполнено</button>
            `;
            
        } catch (error) {
            console.error('Failed to load dashboard:', error);
        }
    }
    
    // Загрузка профиля
    async function loadProfile() {
        try {
            const profile = await getProfile();
            document.getElementById('profileFullName').value = profile.full_name || '';
            document.getElementById('profileAge').value = profile.age || '';
            document.getElementById('profileGender').value = profile.gender || '';
            document.getElementById('profileHeight').value = profile.height_cm || '';
            document.getElementById('profileWeight').value = profile.weight_kg || '';
            document.getElementById('profileFitnessLevel').value = profile.fitness_level || '';
            
            const goalsCheckboxes = document.querySelectorAll('#profileForm input[type="checkbox"]');
            if (profile.goals) {
                goalsCheckboxes.forEach(cb => {
                    cb.checked = profile.goals.includes(cb.value);
                });
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
        }
    }
    
    // Сохранение профиля
    const profileForm = document.getElementById('profileForm');
    if (profileForm) {
        profileForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const goals = Array.from(document.querySelectorAll('#profileForm input[type="checkbox"]:checked'))
                .map(cb => cb.value);
            
            const profileData = {
                age: parseInt(document.getElementById('profileAge').value) || 0,
                gender: document.getElementById('profileGender').value,
                height_cm: parseInt(document.getElementById('profileHeight').value) || 0,
                weight_kg: parseFloat(document.getElementById('profileWeight').value) || 0,
                fitness_level: document.getElementById('profileFitnessLevel').value,
                goals: goals
            };
            
            try {
                await updateProfile(profileData);
                alert('Профиль сохранен');
            } catch (error) {
                alert('Ошибка сохранения: ' + error.message);
            }
        });
    }
    
    // Загрузка программ тренировок
    async function loadTrainingPlans() {
        try {
            const data = await getTrainingPlans();
            const container = document.getElementById('trainingPlans');
            if (data.plans && data.plans.length > 0) {
                container.innerHTML = data.plans.map(plan => `
                    <div class="workout-card">
                        <h3>Программа от ${new Date(plan.generated_at).toLocaleDateString()}</h3>
                        <p>Статус: ${plan.status}</p>
                        <pre>${JSON.stringify(plan.plan_data, null, 2)}</pre>
                    </div>
                `).join('');
            } else {
                container.innerHTML = '<p>Нет активных программ. Сгенерируйте новую!</p>';
            }
        } catch (error) {
            console.error('Failed to load training plans:', error);
        }
    }
    
    // Генерация программы
    const generateBtn = document.getElementById('generatePlanBtn');
    if (generateBtn) {
        generateBtn.addEventListener('click', async () => {
            try {
                const result = await generateTrainingPlan(4, [1,3,5]);
                alert('Программа сгенерирована!');
                loadTrainingPlans();
            } catch (error) {
                alert('Ошибка генерации: ' + error.message);
            }
        });
    }
    
    // Загрузка достижений
    async function loadAchievements() {
        try {
            const data = await getAchievements();
            const container = document.getElementById('achievementsList');
            if (data.achievements && data.achievements.length > 0) {
                container.innerHTML = data.achievements.map(ach => `
                    <div class="achievement-card">
                        <h3>${ach.name}</h3>
                        <p>${ach.description}</p>
                    </div>
                `).join('');
            } else {
                container.innerHTML = '<p>Пока нет достижений. Продолжайте тренироваться!</p>';
            }
        } catch (error) {
            console.error('Failed to load achievements:', error);
        }
    }
    
    // Проверяем, авторизован ли пользователь
    if (authToken) {
        showView('dashboard');
    }
    
    // Навигация по ссылкам
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const href = link.getAttribute('href');
            let view = href === '/' ? 'dashboard' : href.slice(1);
            showView(view);
            if (navMenu.classList.contains('active')) {
                navMenu.classList.remove('active');
            }
        });
    });
});