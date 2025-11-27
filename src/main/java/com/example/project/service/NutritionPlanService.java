package com.example.project.service;

import java.util.List;
import java.util.Optional;
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

    public List<NutritionPlan> findByGoal(String goal) {
        return nutritionPlanRepository.findByGoalType(goal);
    }

    public List<NutritionPlan> findByGoalAndDifficulty(String goal, String difficulty) {
        return nutritionPlanRepository.findByGoalTypeAndDifficultyLevel(goal, difficulty);
    }

    public List<NutritionPlan> findAllNutritionPlans() {
        return nutritionPlanRepository.findAll();
    }

    public Optional<NutritionPlan> findById(Integer id) {
        return nutritionPlanRepository.findById(id);
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

    // ДОБАВИТЬ: метод для получения плана питания по ID
    public NutritionPlan getNutritionPlanById(Integer id) {
        return nutritionPlanRepository.findById(id).orElse(null);
    }
}