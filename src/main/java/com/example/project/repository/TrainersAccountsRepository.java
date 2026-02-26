package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.accounts.TrainersAccounts;

@Repository
public interface TrainersAccountsRepository extends JpaRepository<TrainersAccounts, String> {

    Optional<TrainersAccounts> findByTrainerIdTrainer(Integer trainerId);

    Set<TrainersAccounts> findByUserRole(String userRole);

    Set<TrainersAccounts> findByLastLoginBefore(LocalDate date);

    Set<TrainersAccounts> findByAccountCreationDateAfter(LocalDate startDate);
}