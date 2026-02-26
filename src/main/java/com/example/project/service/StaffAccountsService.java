package com.example.project.service;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.accounts.StaffAccounts;
import com.example.project.repository.StaffAccountsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class StaffAccountsService {
    private final StaffAccountsRepository staffAccountsRepository;

    @SuppressWarnings("null")
    public StaffAccounts getStaffAccount(String username) {
        return staffAccountsRepository.findById(username).orElse(null);
    }

    public StaffAccounts getStaffAccountByStaffId(Integer staffId) {
        return staffAccountsRepository.findByStaffIdStaff(staffId).orElse(null);
    }

    public Set<StaffAccounts> getAccountsByRole(String userRole) {
        return staffAccountsRepository.findByUserRole(userRole);
    }

    public Set<StaffAccounts> getInactiveAccounts(LocalDate cutoffDate) {
        return staffAccountsRepository.findByLastLoginBefore(cutoffDate);
    }

    @SuppressWarnings("null")
    public StaffAccounts saveStaffAccount(StaffAccounts account) {
        return staffAccountsRepository.save(account);
    }
}