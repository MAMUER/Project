export default function LoginForm({
  formData,
  errors,
  generalError,
  submitting,
  getFieldClass,
  setField,
  onSubmit,
  onSwitchMode,
}) {
  return (
    <form
      className='auth-form'
      onSubmit={onSubmit}
      noValidate
      aria-label='Форма входа'
    >
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
        <div className='field-error' id='login-email-error' role='alert'>
          {errors.email || ''}
        </div>
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
          aria-describedby={
            errors.password ? 'login-password-error' : undefined
          }
        />
        <div className='field-error' id='login-password-error' role='alert'>
          {errors.password || ''}
        </div>
      </div>
      <div className='auth-error hidden' role='alert' aria-live='assertive'>
        {generalError}
      </div>
      <button type='submit' className='btn-primary' disabled={submitting}>
        {submitting ? 'Вход...' : 'Войти'}
      </button>
      <p className='auth-switch'>
        Нет аккаунта?{' '}
        <button type='button' className='link-button' onClick={onSwitchMode}>
          Создать
        </button>
      </p>
    </form>
  );
}
