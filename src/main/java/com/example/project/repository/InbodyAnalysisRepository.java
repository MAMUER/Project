package com.example.project.repository;

import com.example.project.model.InbodyAnalysis;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.Set;

@Repository
public interface InbodyAnalysisRepository extends JpaRepository<InbodyAnalysis, Integer> {
    
    // Базовые методы поиска по диапазонам
    Set<InbodyAnalysis> findByBmiBetween(Float minBmi, Float maxBmi);
    
    Set<InbodyAnalysis> findByFatPercentBetween(Float minFat, Float maxFat);
    
    Set<InbodyAnalysis> findByMusclePercentBetween(Float minMuscle, Float maxMuscle);
    
    Set<InbodyAnalysis> findByWeightBetween(Float minWeight, Float maxWeight);
    
    Set<InbodyAnalysis> findByHeightBetween(Float minHeight, Float maxHeight);
    
    // Поиск анализов с BMI выше/ниже указанного
    Set<InbodyAnalysis> findByBmiGreaterThanEqual(Float minBmi);
    
    Set<InbodyAnalysis> findByBmiLessThanEqual(Float maxBmi);
    
    // Статистические запросы
    @Query("SELECT AVG(ia.bmi) FROM InbodyAnalysis ia WHERE ia.bmi IS NOT NULL")
    Double findAverageBmi();
    
    @Query("SELECT AVG(ia.weight) FROM InbodyAnalysis ia WHERE ia.weight IS NOT NULL")
    Double findAverageWeight();
    
    @Query("SELECT AVG(ia.fatPercent) FROM InbodyAnalysis ia WHERE ia.fatPercent IS NOT NULL")
    Double findAverageFatPercent();
    
    @Query("SELECT AVG(ia.musclePercent) FROM InbodyAnalysis ia WHERE ia.musclePercent IS NOT NULL")
    Double findAverageMusclePercent();
}