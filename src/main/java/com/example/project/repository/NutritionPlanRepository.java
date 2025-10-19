package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.NutritionPlan;

@Repository
public interface NutritionPlanRepository extends JpaRepository<NutritionPlan, Integer> {

    Set<NutritionPlan> findByMemberIdMember(Integer memberId);

    Set<NutritionPlan> findByStartDateBetween(LocalDate start, LocalDate end);

    Set<NutritionPlan> findByEndDateIsNull();

    Set<NutritionPlan> findByNutritionDescriptionContaining(String keyword);

    default Set<NutritionPlan> findActivePlans(LocalDate date) {
        return findByStartDateLessThanEqualAndEndDateIsNullOrEndDateGreaterThanEqual(date, date);
    }

    Set<NutritionPlan> findByStartDateLessThanEqualAndEndDateIsNullOrEndDateGreaterThanEqual(LocalDate date,
            LocalDate date2);
}