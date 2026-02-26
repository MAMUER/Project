package com.example.project.service;

import java.util.Arrays;
import java.util.List;
import java.util.regex.Pattern;

import org.springframework.stereotype.Service;

@Service
public class PasswordValidationService {
    
    // Простые слова и комбинации для проверки
    private static final List<String> COMMON_PATTERNS = Arrays.asList(
        "123456", "password", "qwerty", "admin", "welcome", "monkey", "sunshine",
        "password1", "12345678", "abc123", "football", "baseball", "welcome1",
        "123456789", "1234567", "123123", "000000", "111111", "1234567890"
    );
    
    // Клавиатурные последовательности
    private static final List<String> KEYBOARD_SEQUENCES = Arrays.asList(
        "qwerty", "asdfgh", "zxcvbn", "qazwsx", "123qwe", "1qaz2wsx", "!qaz@wsx"
    );

    public PasswordValidationResult validatePassword(String password, String username, String firstName, String lastName) {
        PasswordValidationResult result = new PasswordValidationResult();
        
        if (password == null || password.length() < 12) {
            result.addError("Пароль должен содержать минимум 12 символов");
            return result; // Ранний возврат, если пароль null
        }
        
        if (!Pattern.compile("[A-Z]").matcher(password).find()) {
            result.addError("Пароль должен содержать хотя бы одну заглавную букву");
        }
        
        if (!Pattern.compile("[a-z]").matcher(password).find()) {
            result.addError("Пароль должен содержать хотя бы одну строчную букву");
        }
        
        if (!Pattern.compile("[0-9]").matcher(password).find()) {
            result.addError("Пароль должен содержать хотя бы одну цифру");
        }
        
        if (!Pattern.compile("[!@#]").matcher(password).find()) {
            result.addError("Пароль должен содержать хотя бы один специальный символ (!, @ или #)");
        }
        
        // Проверка на простые комбинации
        if (containsCommonPatterns(password)) {
            result.addError("Пароль слишком простой. Не используйте распространенные комбинации");
        }
        
        // ИСПРАВЛЕНО: Добавлена проверка на null
        if (containsKeyboardSequence(password.toLowerCase())) {
            result.addError("Пароль содержит простую клавиатурную последовательность");
        }
        
        // Проверка на использование личной информации
        if (containsPersonalInfo(password, username, firstName, lastName)) {
            result.addError("Пароль не должен содержать ваше имя, фамилию или имя пользователя");
        }
        
        // Проверка на повторяющиеся символы
        if (hasRepeatingCharacters(password)) {
            result.addError("Пароль содержит слишком много повторяющихся символов");
        }
        
        return result;
    }
    
    private boolean containsCommonPatterns(String password) {
        if (password == null) return false;
        String lowerPassword = password.toLowerCase();
        return COMMON_PATTERNS.stream().anyMatch(lowerPassword::contains);
    }
    
    private boolean containsKeyboardSequence(String password) {
        if (password == null) return false; // ДОБАВЛЕНО
        return KEYBOARD_SEQUENCES.stream().anyMatch(password::contains);
    }
    
    private boolean containsPersonalInfo(String password, String username, String firstName, String lastName) {
        if (password == null) return false;
        String lowerPassword = password.toLowerCase();
        return (username != null && lowerPassword.contains(username.toLowerCase())) ||
               (firstName != null && lowerPassword.contains(firstName.toLowerCase())) ||
               (lastName != null && lowerPassword.contains(lastName.toLowerCase()));
    }
    
    private boolean hasRepeatingCharacters(String password) {
        if (password == null) return false;
        // Проверяем 3+ одинаковых символа подряд
        return Pattern.compile("(.)\\1{2,}").matcher(password).find();
    }
    
    public static class PasswordValidationResult {
        private boolean valid = true;
        private final List<String> errors = new java.util.ArrayList<>();
        
        public void addError(String error) {
            errors.add(error);
            valid = false;
        }
        
        public boolean isValid() {
            return valid;
        }
        
        public List<String> getErrors() {
            return errors;
        }
        
        public String getErrorMessage() {
            return String.join(", ", errors);
        }
    }
}