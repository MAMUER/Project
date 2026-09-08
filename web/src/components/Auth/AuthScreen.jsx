import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import LoginForm from './LoginForm';
import RegisterForm from './RegisterForm';
import TwoFAForm from './TwoFAForm';
import { useAuthForm } from './useAuthForm';
import VerifyEmail from './VerifyEmail';
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
          <div className='logo-icon' aria-hidden='true'>
            💓
          </div>
          <h1>FitPulse</h1>
          <p>Ваш персональный AI-тренер</p>
        </div>

        <div className='auth-landing'>
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
          <LoginForm
            formData={formData}
            errors={errors}
            generalError={generalError}
            submitting={submitting}
            getFieldClass={getFieldClass}
            setField={setField}
            onSubmit={handleLoginSubmit}
            onSwitchMode={() => setMode('register')}
          />
        )}

        {mode === 'register' && (
          <RegisterForm
            formData={formData}
            errors={errors}
            generalError={generalError}
            passwordChecks={passwordChecks}
            submitting={submitting}
            getFieldClass={getFieldClass}
            setField={setField}
            onSubmit={handleRegister}
            updatePasswordChecks={updatePasswordChecks}
            onSwitchMode={() => setMode('login')}
          />
        )}

        {mode === 'login2fa' && (
          <TwoFAForm
            formData={formData}
            generalError={generalError}
            submitting={submitting}
            setField={setField}
            onSubmit={handleLogin2FASubmit}
            onBack={() => {
              setMode('login');
              setTwoFATempToken(null);
            }}
          />
        )}

        {mode === 'verify' && (
          <VerifyEmail
            email={formData.email}
            successMessage={successMessage}
            generalError={generalError}
            onBack={() => setMode('login')}
          />
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
