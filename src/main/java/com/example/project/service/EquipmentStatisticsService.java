package com.example.project.service;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import com.example.project.repository.EquipmentStatisticsRepository;
@AllArgsConstructor
@Service
public class EquipmentStatisticsService {
    private final EquipmentStatisticsRepository equipmentStatisticsRepository;

    public int getNumberOfApproaches(int statisticId) {
        return equipmentStatisticsRepository
            .findById(statisticId)
            .orElse(null)
            .getApproaches();
    }
    
    public int getAmountOfKilocalories(int statisticId) {
        return equipmentStatisticsRepository
            .findById(statisticId)
            .orElse(null)
            .getKilocalories();
    }

    public String getActivityName(int statisticId) {
        return equipmentStatisticsRepository
            .findById(statisticId)
            .orElse(null)
            .getActivityType()
            .getActivityName();
    }
}
