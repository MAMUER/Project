package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.InbodyAnalyses;

@Repository
public interface InbodyAnalysesRepository extends JpaRepository<InbodyAnalyses, Integer> {
    
    Set<InbodyAnalyses> findByBmiBetween(Float minBmi, Float maxBmi);
    
    Set<InbodyAnalyses> findByFatPercentBetween(Float minFat, Float maxFat);
    
    Set<InbodyAnalyses> findByMusclePersentBetween(Float minMuscle, Float maxMuscle);
    
    Set<InbodyAnalyses> findByWeightBetween(Float minWeight, Float maxWeight);
}