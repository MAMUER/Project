package com.example.project.repository;


import java.time.LocalDate;
import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.Accounts.StaffAccounts;
@Repository
public interface StaffAccountsRepository extends JpaRepository<StaffAccounts, String> {
    
    Optional<StaffAccounts> findByStaffIdStaff(Integer staffId);
    
    Set<StaffAccounts> findByUserRole(String userRole);
    
    Set<StaffAccounts> findByLastLoginBefore(LocalDate date);
    
    Set<StaffAccounts> findByAccountCreationDateBetween(LocalDate start, LocalDate end);
}