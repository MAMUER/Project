package com.example.project.repository;

import java.util.Set;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import com.example.project.model.NutritionPlan;

@Repository
public interface NutritionPlanRepository extends JpaRepository<NutritionPlan, Integer> {
    List<NutritionPlan> findByGoalType(String goalType);

    List<NutritionPlan> findByGoalTypeAndDifficultyLevel(String goalType, String difficultyLevel);

    List<NutritionPlan> findByDifficultyLevel(String difficultyLevel);

    Set<NutritionPlan> findByNutritionDescriptionContaining(String keyword);

    // Поиск диет по диапазону калорий
    @Query("SELECT np FROM NutritionPlan np WHERE np.caloriesPerDay BETWEEN :minCalories AND :maxCalories")
    List<NutritionPlan> findByCaloriesRange(@Param("minCalories") Integer minCalories,
            @Param("maxCalories") Integer maxCalories);

    // Поиск диет по проценту белка
    @Query("SELECT np FROM NutritionPlan np WHERE np.proteinPercent >= :minProtein")
    List<NutritionPlan> findByMinProtein(@Param("minProtein") Integer minProtein);
}