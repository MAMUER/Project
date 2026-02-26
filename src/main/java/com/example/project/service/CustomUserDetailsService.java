package com.example.project.service;

import java.util.Collections;

import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CustomUserDetailsService implements UserDetailsService {
    private final AccountService accountService;

    @Override
    @Transactional
    public UserDetails loadUserByUsername(String username) throws UsernameNotFoundException {
        AccountService.AccountInfo account = accountService.getAccountInfo(username);
        if (account == null) {
            throw new UsernameNotFoundException("User not found with username: " + username);
        }
        
        // ИСПРАВЛЕНО: Добавляем authorities на основе роли из БД
        var authorities = Collections.singletonList(new SimpleGrantedAuthority(account.role()));
        return new User(account.username(), account.password(), authorities);
    }

    @Transactional
    public UserDetails loadUserById(Integer userId) throws UsernameNotFoundException {
        AccountService.AccountInfo account = accountService.getAccountInfoById(userId);
        if (account == null) {
            throw new UsernameNotFoundException("User not found with id: " + userId);
        }
        
        // ИСПРАВЛЕНО: Добавляем authorities
        var authorities = Collections.singletonList(new SimpleGrantedAuthority(account.role()));
        return new User(account.username(), account.password(), authorities);
    }

    public Integer getUserId(String username) {
        return accountService.getIdByUsername(username);
    }

    public String getUserRole(String username) {
        String role = accountService.getRoleByUsername(username);
        if (role == null)
            return null;

        return switch (role.toUpperCase()) {
            case "MEMBER" -> "member";
            case "TRAINER" -> "trainer";
            case "STAFF" -> "staff";
            default -> null;
        };
    }

    public String getNormalizedRole(String username) {
        return getUserRole(username);
    }

    // Дополнительные полезные методы
    public boolean isMember(String username) {
        return "member".equals(getUserRole(username));
    }

    public boolean isTrainer(String username) {
        return "trainer".equals(getUserRole(username));
    }

    public boolean isStaff(String username) {
        return "staff".equals(getUserRole(username));
    }
}