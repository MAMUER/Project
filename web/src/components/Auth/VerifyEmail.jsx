export default function VerifyEmail({ email, successMessage, generalError, onBack }) {
  return (
    <div className='auth-form verify-form'>
      <div className='verify-icon' aria-hidden='true'>
        📧
      </div>
      <h2>Проверьте почту</h2>
      <p className='verify-text'>
        Мы отправили письмо на{' '}
        <strong>{email || 'ваш email'}</strong>
      </p>
      <p className='verify-text'>
        Перейдите по ссылке из письма, чтобы подтвердить email и войти.
      </p>
      {successMessage && (
        <div className='auth-success'>{successMessage}</div>
      )}
      {generalError && <div className='auth-error'>{generalError}</div>}
      <p className='auth-switch'>
        <button type='button' className='link-button' onClick={onBack}>
          ← Вернуться ко входу
        </button>
      </p>
    </div>
  );
}
