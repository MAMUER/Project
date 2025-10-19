package com.example.project.service;

import java.util.ArrayList;

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
        return new User(account.username(), account.password(), new ArrayList<>());
    }

    @Transactional
    public UserDetails loadUserById(Integer userId) throws UsernameNotFoundException {
        AccountService.AccountInfo account = accountService.getAccountInfoById(userId);
        if (account == null) {
            throw new UsernameNotFoundException("User not found with id: " + userId);
        }
        return new User(account.username(), account.password(), new ArrayList<>());
    }

    public Integer getUserId(String username) {
        return accountService.getIdByUsername(username);
    }

    public String getUserRole(String username) {
        String role = accountService.getRoleByUsername(username);
        if (role == null)
            return null;

        switch (role.toUpperCase()) {
            case "MEMBER":
                return "member";
            case "TRAINER":
                return "trainer";
            case "STAFF":
                return "staff";
            default:
                return null;
        }
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