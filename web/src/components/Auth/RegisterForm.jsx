import { useAuthForm } from './useAuthForm';
import './Auth.css';

export default function RegisterForm({ formData, errors, generalError, passwordChecks, submitting, getFieldClass, setField, onSubmit, updatePasswordChecks, onSwitchMode }) {
  return (
    <form className='auth-form' onSubmit={onSubmit} noValidate aria-label='Форма регистрации'>
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
        <div className='field-error' id='register-name-error' role='alert'>
          {errors.name || ''}
        </div>
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
        <div className='field-error' id='register-email-error' role='alert'>
          {errors.email || ''}
        </div>
      </div>
      <div className='field'>
        <label htmlFor='register-password'>
          Пароль (мин. 8 символов)
        </label>
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
        <div className='field-error' id='register-password-error' role='alert'>
          {errors.password || ''}
        </div>
        <div
          className={`password-hint ${formData.password ? '' : 'hidden'}`}
          id='password-hints'
        >
          <span className={`hint-item ${passwordChecks.length ? 'pass' : ''}`}>
            {passwordChecks.length ? '✓' : '✗'} 8+ символов
          </span>
          <span className={`hint-item ${passwordChecks.upper ? 'pass' : ''}`}>
            {passwordChecks.upper ? '✓' : '✗'} Заглавная буква
          </span>
          <span className={`hint-item ${passwordChecks.lower ? 'pass' : ''}`}>
            {passwordChecks.lower ? '✓' : '✗'} Строчная буква
          </span>
          <span className={`hint-item ${passwordChecks.digit ? 'pass' : ''}`}>
            {passwordChecks.digit ? '✓' : '✗'} Цифра
          </span>
        </div>
      </div>
      <div className='auth-error hidden'>{generalError}</div>
      <button
        type='submit'
        className='btn-primary'
        disabled={submitting || Object.values(passwordChecks).some((v) => !v)}
      >
        {submitting ? 'Создание...' : 'Создать аккаунт'}
      </button>
      <p className='auth-switch'>
        Уже есть?{' '}
        <button type='button' className='link-button' onClick={onSwitchMode}>
          Войти
        </button>
      </p>
    </form>
  );
}
