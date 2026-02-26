package com.example.project.config;

import org.springframework.security.core.Authentication;
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;
import org.springframework.stereotype.Component;

import com.example.project.service.AccountService;

import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.time.LocalDate;

@Component
public class CustomAuthenticationSuccessHandler implements AuthenticationSuccessHandler {

    private final AccountService accountService;

    public CustomAuthenticationSuccessHandler(AccountService accountService) {
        this.accountService = accountService;
    }

    @Override
    public void onAuthenticationSuccess(HttpServletRequest request, HttpServletResponse response,
            Authentication authentication) throws IOException {

        String username = authentication.getName();
        System.out.println("=== AUTH SUCCESS for user: " + username + " ===");

        // Обновляем дату последнего входа
        accountService.updateLastLogin(username);
        System.out.println(" Updated last_login to: " + LocalDate.now());

        // Получаем информацию об аккаунте
        AccountService.AccountInfo accountInfo = accountService.getAccountInfo(username);

        if (accountInfo != null) {
            String role = accountInfo.role().toLowerCase();
            Integer userId = accountInfo.id();

            String targetUrl = "/profile/" + role + "/" + userId;
            System.out.println("🔀 Redirecting to: " + targetUrl);

            response.sendRedirect(targetUrl);
        } else {
            // Если не нашли пользователя, перенаправляем на страницу ошибки
            response.sendRedirect("/login?error");
        }
    }
}