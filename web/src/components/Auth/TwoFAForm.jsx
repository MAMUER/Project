import { useState } from 'react';

export default function TwoFAForm({ formData, generalError, submitting, setField, onSubmit, onBack }) {
  const [backupCode, setBackupCode] = useState('');

  const handleBackupClick = () => {
    setField('backupCode', formData.totpCode);
    setBackupCode(formData.totpCode);
  };

  return (
    <form className='auth-form' onSubmit={onSubmit} noValidate aria-label='Двухфакторная аутентификация'>
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
        <div className='field-error' id='login2fa-error' role='alert'>
          {generalError || ''}
        </div>
      </div>
      <div className='field'>
        <label htmlFor='login-backup-code'>Резервный код xxxx-xxxx</label>
        <input
          id='login-backup-code'
          type='text'
          placeholder='Резервный код xxxx-xxxx'
          value={backupCode}
          onChange={(e) => {
            setBackupCode(e.target.value);
            setField('backupCode', e.target.value);
          }}
          maxLength={9}
          aria-invalid={false}
        />
        <div className='field-error' role='alert'></div>
      </div>
      <div className='auth-error' role='alert' aria-live='assertive'>
        {generalError}
      </div>
      <button type='submit' className='btn-primary' disabled={submitting}>
        {submitting ? 'Вход...' : 'Войти'}
      </button>
      <button type='button' className='btn-secondary' onClick={handleBackupClick}>
        Использовать резервный код
      </button>
      <p className='auth-switch'>
        <button type='button' className='link-button' onClick={onBack}>
          ← Вернуться ко входу
        </button>
      </p>
    </form>
  );
}
