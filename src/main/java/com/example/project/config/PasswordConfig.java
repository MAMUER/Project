package com.example.project.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCrypt;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
public class PasswordConfig {

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new PasswordEncoder() {
            @Override
            public String encode(CharSequence rawPassword) {
                return BCrypt.hashpw(rawPassword.toString(), BCrypt.gensalt());
            }

            @Override
            public boolean matches(CharSequence rawPassword, String encodedPassword) {
                // Заменяем $2y$ на $2a$ для совместимости с Spring Security
                if (encodedPassword.startsWith("$2y$")) {
                    encodedPassword = "$2a$" + encodedPassword.substring(4);
                }
                return BCrypt.checkpw(rawPassword.toString(), encodedPassword);
            }
        };
    }
}