package com.example.project.repository;

import java.util.Set;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import com.example.project.model.NutritionPlan;

@Repository
public interface NutritionPlanRepository extends JpaRepository<NutritionPlan, Integer> {

    Set<NutritionPlan> findByMemberIdMember(Integer memberId);

    Set<NutritionPlan> findByNutritionDescriptionContaining(String keyword);
}