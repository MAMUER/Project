package com.example.project.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
@EnableWebSecurity
@EnableMethodSecurity()
public class SpringSecurityConfig {

    private final CustomAuthenticationSuccessHandler authenticationSuccessHandler;

    public SpringSecurityConfig(CustomAuthenticationSuccessHandler authenticationSuccessHandler) {
        this.authenticationSuccessHandler = authenticationSuccessHandler;
    }

    @Bean
    SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
            .authorizeHttpRequests(auth -> auth
                // Публичные пути
                .requestMatchers("/public/**", "/login", "/registration", "/css/**",
                               "/js/**", "/diagnose", "/direct-diagnose", "/check-password",
                               "/error", "/error/**")
                .permitAll()
                // Админские пути - только для STAFF
                .requestMatchers("/admin/**").hasAuthority("STAFF")
                // Все остальные запросы требуют аутентификации
                .anyRequest().authenticated())
            .formLogin(form -> form
                .loginPage("/login")
                .successHandler(authenticationSuccessHandler)
                .permitAll())
            .logout(logout -> logout
                .logoutSuccessUrl("/login?logout")
                .permitAll())
            .exceptionHandling(exceptions -> exceptions
                .accessDeniedPage("/error/403"))
            .csrf(csrf -> csrf.ignoringRequestMatchers("/check-password"));

        return http.build();
    }
}