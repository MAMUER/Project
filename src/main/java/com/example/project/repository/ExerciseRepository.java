package com.example.project.repository;

import com.example.project.model.Exercise;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.Set;

@Repository
public interface ExerciseRepository extends JpaRepository<Exercise, Integer> {
    Set<Exercise> findByMuscleGroup(String muscleGroup);
    Set<Exercise> findByDifficultyLevel(Integer difficultyLevel);
    Set<Exercise> findByEquipmentRequiredContaining(String equipment);
    Set<Exercise> findByMuscleGroupAndDifficultyLevel(String muscleGroup, Integer difficultyLevel);

    @Query("SELECT DISTINCT e.muscleGroup FROM Exercise e")
    Set<String> findDistinctMuscleGroups();
}