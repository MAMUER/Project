package com.example.project.service;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.accounts.TrainersAccounts;
import com.example.project.repository.TrainersAccountsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class TrainersAccountsService {
    private final TrainersAccountsRepository trainersAccountsRepository;

    @SuppressWarnings("null")
    public TrainersAccounts getTrainerAccount(String username) {
        return trainersAccountsRepository.findById(username).orElse(null);
    }

    public TrainersAccounts getTrainerAccountByTrainerId(Integer trainerId) {
        return trainersAccountsRepository.findByTrainerIdTrainer(trainerId).orElse(null);
    }

    public Set<TrainersAccounts> getAccountsByRole(String userRole) {
        return trainersAccountsRepository.findByUserRole(userRole);
    }

    public Set<TrainersAccounts> getInactiveAccounts(LocalDate cutoffDate) {
        return trainersAccountsRepository.findByLastLoginBefore(cutoffDate);
    }

    @SuppressWarnings("null")
    public TrainersAccounts saveTrainerAccount(TrainersAccounts account) {
        return trainersAccountsRepository.save(account);
    }
}