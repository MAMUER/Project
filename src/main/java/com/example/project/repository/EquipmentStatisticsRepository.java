package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.EquipmentStatistics;

@Repository
public interface EquipmentStatisticsRepository extends JpaRepository<EquipmentStatistics, Integer> {

    Set<EquipmentStatistics> findByApproachesGreaterThan(int approaches);

    Set<EquipmentStatistics> findByKilocaloriesBetween(int minCalories, int maxCalories);

    Set<EquipmentStatistics> findByActivityTypeIdActivity(Integer activityTypeId);
}