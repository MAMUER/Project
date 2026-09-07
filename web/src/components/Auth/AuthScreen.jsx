import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuthForm } from './useAuthForm';
import './Auth.css';

export default function AuthScreen({ searchParams: searchParamsProp }) {
  const [routerSearchParams] = useSearchParams();
  const searchParams = searchParamsProp || routerSearchParams;
  const [mode, setMode] = useState('login');
  const [successMessage, setSuccessMessage] = useState('');
  const [twoFATempToken, setTwoFATempToken] = useState(null);

  const {
    formData,
    errors,
    generalError,
    passwordChecks,
    submitting,
    setField,
    getFieldClass,
    handleLogin,
    handleRegister,
    handleLogin2FA,
    updatePasswordChecks,
  } = useAuthForm({
    searchParams,
    onModeChange: setMode,
    onSuccessMessage: setSuccessMessage,
  });

  const handleLoginSubmit = async (e) => {
    const result = await handleLogin(e);
    if (result?.requires2FA) {
      setTwoFATempToken(result.tempToken);
    }
  };

  const handleLogin2FASubmit = (e) => {
    handleLogin2FA(e, twoFATempToken);
  };

  return (
    <div className='auth-screen'>
      <div className='auth-container'>
        <div className='auth-logo'>
          <div className='logo-icon' aria-hidden='true'>💓</div>
          <h1>FitPulse</h1>
          <p>Ваш персональный AI-тренер</p>
        </div>

        <div className='auth-landing' aria-label='О платформе FitPulse'>
          <p style={{ marginBottom: 12 }}>
            FitPulse — это открытая платформа для фитнес- и health-трекинга. Мы
            помогаем отслеживать пульс, SpO2, шаги, сон и тренировки,
            синхронизировать данные с носимых устройств и получать
            персонализированные insights.
          </p>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
              gap: 10,
              textAlign: 'left',
              marginTop: 18,
            }}
            aria-label='Возможности платформы'
          >
            {[
              '📊 Биометрия и активность',
              '⌚ Все основные бренды',
              '🤖 AI-планы тренировок',
              '🔒 End-to-end защита',
            ].map((feature) => (
              <div
                key={feature}
                style={{
                  background: 'var(--bg-card)',
                  borderRadius: 'var(--radius-md)',
                  padding: '12px 14px',
                  fontSize: 13,
                }}
              >
                {feature}
              </div>
            ))}
          </div>
          <p
            style={{
              marginTop: 16,
              fontSize: 13,
              color: 'var(--text-tertiary)',
            }}
          >
            Мы собираем только данные, необходимые для работы сервиса: учётные
            записи, биометрию с устройств, технические логи. Вы можете запросить
            копию или удаление данных в любой момент.
          </p>
        </div>

        {mode === 'login' && (
          <form className='auth-form' onSubmit={handleLoginSubmit} noValidate aria-label='Форма входа'>
            <div className='field'>
              <label htmlFor='login-email'>Email</label>
              <input
                id='login-email'
                type='email'
                placeholder='Email'
                value={formData.email}
                onChange={(e) => setField('email', e.target.value)}
                autoComplete='email'
                required
                maxLength={254}
                inputMode='email'
                className={getFieldClass('email')}
                aria-invalid={!!errors.email}
                aria-describedby={errors.email ? 'login-email-error' : undefined}
              />
              <div className='field-error' id='login-email-error' role='alert'>{errors.email || ''}</div>
            </div>
            <div className='field'>
              <label htmlFor='login-password'>Пароль</label>
              <input
                id='login-password'
                type='password'
                placeholder='Пароль'
                value={formData.password}
                onChange={(e) => setField('password', e.target.value)}
                autoComplete='current-password'
                required
                maxLength={128}
                className={errors.password ? 'invalid' : ''}
                aria-invalid={!!errors.password}
                aria-describedby={errors.password ? 'login-password-error' : undefined}
              />
              <div className='field-error' id='login-password-error' role='alert'>{errors.password || ''}</div>
            </div>
            <div className='auth-error hidden' role='alert' aria-live='assertive'>{generalError}</div>
            <button type='submit' className='btn-primary' disabled={submitting}>
              {submitting ? 'Вход...' : 'Войти'}
            </button>
            <p className='auth-switch'>
              Нет аккаунта?{' '}
              <button
                type='button'
                className='link-button'
                onClick={() => setMode('register')}
              >
                Создать
              </button>
            </p>
          </form>
        )}

        {mode === 'register' && (
          <form className='auth-form' onSubmit={handleRegister} noValidate aria-label='Форма регистрации'>
            <div className='field'>
              <label htmlFor='register-name'>Имя</label>
              <input
                id='register-name'
                type='text'
                placeholder='Имя'
                value={formData.name}
                onChange={(e) => setField('name', e.target.value)}
                autoComplete='name'
                required
                maxLength={100}
                minLength={2}
                className={getFieldClass('name')}
                aria-invalid={!!errors.name}
                aria-describedby={errors.name ? 'register-name-error' : undefined}
              />
              <div className='field-error' id='register-name-error' role='alert'>{errors.name || ''}</div>
            </div>
            <div className='field'>
              <label htmlFor='register-email'>Email</label>
              <input
                id='register-email'
                type='email'
                placeholder='Email'
                value={formData.email}
                onChange={(e) => setField('email', e.target.value)}
                autoComplete='email'
                required
                maxLength={254}
                inputMode='email'
                className={getFieldClass('email')}
                aria-invalid={!!errors.email}
                aria-describedby={errors.email ? 'register-email-error' : undefined}
              />
              <div className='field-error' id='register-email-error' role='alert'>{errors.email || ''}</div>
            </div>
            <div className='field'>
              <label htmlFor='register-password'>Пароль (мин. 8 символов)</label>
              <input
                id='register-password'
                type='password'
                placeholder='Пароль (мин. 8 символов)'
                value={formData.password}
                onChange={(e) => {
                  setField('password', e.target.value);
                  updatePasswordChecks(e.target.value);
                }}
                autoComplete='new-password'
                required
                minLength={8}
                maxLength={128}
                className={errors.password ? 'invalid' : ''}
                aria-invalid={!!errors.password}
                aria-describedby={errors.password ? 'register-password-error' : 'password-hints'}
              />
              <div className='field-error' id='register-password-error' role='alert'>{errors.password || ''}</div>
              <div
                className={`password-hint ${formData.password ? '' : 'hidden'}`}
                id='password-hints'
              >
                <span
                  className={`hint-item ${passwordChecks.length ? 'pass' : ''}`}
                >
                  {passwordChecks.length ? '✓' : '✗'} 8+ символов
                </span>
                <span
                  className={`hint-item ${passwordChecks.upper ? 'pass' : ''}`}
                >
                  {passwordChecks.upper ? '✓' : '✗'} Заглавная буква
                </span>
                <span
                  className={`hint-item ${passwordChecks.lower ? 'pass' : ''}`}
                >
                  {passwordChecks.lower ? '✓' : '✗'} Строчная буква
                </span>
                <span
                  className={`hint-item ${passwordChecks.digit ? 'pass' : ''}`}
                >
                  {passwordChecks.digit ? '✓' : '✗'} Цифра
                </span>
              </div>
            </div>
            <div className='auth-error hidden'>{generalError}</div>
            <button
              type='submit'
              className='btn-primary'
              disabled={
                submitting || Object.values(passwordChecks).some((v) => !v)
              }
            >
              {/* istanbul ignore next */}
              {submitting ? 'Создание...' : 'Создать аккаунт'}
            </button>
            <p className='auth-switch'>
              Уже есть?{' '}
              <button
                type='button'
                className='link-button'
                onClick={() => setMode('login')}
              >
                Войти
              </button>
            </p>
          </form>
        )}

        {mode === 'login2fa' && (
          <form
            className='auth-form'
            onSubmit={handleLogin2FASubmit}
            noValidate
            aria-label='Двухфакторная аутентификация'
          >
            <h2>Двухфакторная аутентификация</h2>
            <p className='verify-text'>
              Введите код из приложения-аутентификатора.
            </p>
            <div className='field'>
              <label htmlFor='login-totp-code'>6-значный код</label>
              <input
                id='login-totp-code'
                type='text'
                placeholder='6-значный код'
                value={formData.totpCode}
                onChange={(e) => setField('totpCode', e.target.value)}
                maxLength={6}
                inputMode='numeric'
                autoComplete='one-time-code'
                aria-invalid={!!generalError}
                aria-describedby={generalError ? 'login2fa-error' : undefined}
              />
              <div className='field-error' id='login2fa-error' role='alert'>{generalError || ''}</div>
            </div>
            <div className='field'>
              <label htmlFor='login-backup-code'>Резервный код xxxx-xxxx</label>
              <input
                id='login-backup-code'
                type='text'
                placeholder='Резервный код xxxx-xxxx'
                value={formData.backupCode}
                onChange={(e) => setField('backupCode', e.target.value)}
                maxLength={9}
                aria-invalid={false}
              />
              <div className='field-error' role='alert'></div>
            </div>
            <div className='auth-error' role='alert' aria-live='assertive'>{generalError}</div>
            <button type='submit' className='btn-primary' disabled={submitting}>
              {submitting ? 'Вход...' : 'Войти'}
            </button>
            <button
              type='button'
              className='btn-secondary'
              onClick={() => setField('backupCode', formData.totpCode)}
            >
              Использовать резервный код
            </button>
            <p className='auth-switch'>
              <button
                type='button'
                className='link-button'
                onClick={() => {
                  setMode('login');
                  setTwoFATempToken(null);
                }}
              >
                ← Вернуться ко входу
              </button>
            </p>
          </form>
        )}

        {/* istanbul ignore next */}
        {mode === 'verify' && (
          <div className='auth-form verify-form'>
            <div className='verify-icon' aria-hidden='true'>📧</div>
            <h2>Проверьте почту</h2>
            <p className='verify-text'>
              Мы отправили письмо на{' '}
              <strong>{formData.email || 'ваш email'}</strong>
            </p>
            <p className='verify-text'>
              Перейдите по ссылке из письма, чтобы подтвердить email и войти.
            </p>
            {/* istanbul ignore next */}
            {successMessage && (
              <div className='auth-success'>{successMessage}</div>
            )}
            {generalError && <div className='auth-error'>{generalError}</div>}
            <p className='auth-switch'>
              <button
                type='button'
                className='link-button'
                onClick={() => setMode('login')}
              >
                ← Вернуться ко входу
              </button>
            </p>
          </div>
        )}

        <div
          style={{
            textAlign: 'center',
            color: 'var(--text-tertiary)',
            fontSize: 13,
            display: 'flex',
            gap: 16,
            justifyContent: 'center',
            flexWrap: 'wrap',
            padding: '0 24px 24px',
          }}
        >
          <a
            href='/privacy'
            style={{ color: 'inherit', textDecoration: 'none' }}
          >
            Политика конфиденциальности
          </a>
          <span>•</span>
          <a href='/terms' style={{ color: 'inherit', textDecoration: 'none' }}>
            Пользовательское соглашение
          </a>
        </div>
      </div>
    </div>
  );
}
