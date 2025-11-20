package com.example.project.service;

import java.util.Set;
import org.springframework.stereotype.Service;
import com.example.project.model.NutritionPlan;
import com.example.project.repository.NutritionPlanRepository;
import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class NutritionPlanService {
    private final NutritionPlanRepository nutritionPlanRepository;

    public NutritionPlan getNutritionPlan(Integer id) {
        return nutritionPlanRepository.findById(id).orElse(null);
    }

    public Set<NutritionPlan> getNutritionPlansByMember(Integer memberId) {
        return nutritionPlanRepository.findByMemberIdMember(memberId);
    }

    public Set<NutritionPlan> searchNutritionPlans(String keyword) {
        return nutritionPlanRepository.findByNutritionDescriptionContaining(keyword);
    }

    public NutritionPlan saveNutritionPlan(NutritionPlan plan) {
        return nutritionPlanRepository.save(plan);
    }

    public void deleteNutritionPlan(Integer id) {
        nutritionPlanRepository.deleteById(id);
    }
}