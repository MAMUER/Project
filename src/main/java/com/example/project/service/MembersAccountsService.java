package com.example.project.service;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.accounts.MembersAccounts;
import com.example.project.repository.MembersAccountsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class MembersAccountsService {
    private final MembersAccountsRepository membersAccountsRepository;

    @SuppressWarnings("null")
    public MembersAccounts getMemberAccount(String username) {
        return membersAccountsRepository.findById(username).orElse(null);
    }

    public MembersAccounts getMemberAccountByMemberId(Integer memberId) {
        return membersAccountsRepository.findByMemberIdMember(memberId).orElse(null);
    }

    public Set<String> getAllUsernames() {
        return membersAccountsRepository.getUsernames();
    }

    public Set<String> getAllPasswords() {
        return membersAccountsRepository.getPasswords();
    }

    public Set<String> getAllUserRoles() {
        return membersAccountsRepository.getUserRoles();
    }

    public Set<MembersAccounts> getAccountsByRole(String userRole) {
        return membersAccountsRepository.findByUserRole(userRole);
    }

    public Set<MembersAccounts> getInactiveAccounts(LocalDate cutoffDate) {
        return membersAccountsRepository.findByLastLoginBefore(cutoffDate);
    }

    @SuppressWarnings("null")
    public void saveMemberAccount(MembersAccounts account) {
        membersAccountsRepository.save(account);
    }
}