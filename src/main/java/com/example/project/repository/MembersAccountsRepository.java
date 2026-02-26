package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import com.example.project.model.Accounts.MembersAccounts;

@Repository
public interface MembersAccountsRepository extends JpaRepository<MembersAccounts, String> {

    @Query(value = "SELECT m.username FROM members_accounts m", nativeQuery = true)
    Set<String> getUsernames();

    @Query(value = "SELECT m.password FROM members_accounts m", nativeQuery = true)
    Set<String> getPasswords();

    @Query(value = "SELECT m.user_role FROM members_accounts m", nativeQuery = true)
    Set<String> getUserRoles();
    
    Optional<MembersAccounts> findByMemberIdMember(Integer memberId);
    
    Set<MembersAccounts> findByUserRole(String userRole);
    
    Set<MembersAccounts> findByLastLoginBefore(LocalDate date);
    
    Set<MembersAccounts> findByAccountCreationDateBetween(LocalDate start, LocalDate end);
}