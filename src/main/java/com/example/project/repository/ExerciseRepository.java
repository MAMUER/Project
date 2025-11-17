package com.example.project.repository;

import com.example.project.model.Exercise;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.Set;

@Repository
public interface ExerciseRepository extends JpaRepository<Exercise, Integer> {
    @Query("SELECT e FROM Exercise e LEFT JOIN FETCH e.equipmentRequirements er LEFT JOIN FETCH er.equipmentType WHERE e.idExercise = :id")
    Optional<Exercise> findByIdWithEquipmentRequirements(@Param("id") Integer id);
    
    @Query("SELECT e FROM Exercise e LEFT JOIN FETCH e.equipmentRequirements er LEFT JOIN FETCH er.equipmentType")
    List<Exercise> findAllWithEquipmentRequirements();
    
    @Query("SELECT e FROM Exercise e LEFT JOIN FETCH e.equipmentRequirements er LEFT JOIN FETCH er.equipmentType WHERE e.muscleGroup = :muscleGroup")
    List<Exercise> findByMuscleGroupWithEquipmentRequirements(@Param("muscleGroup") String muscleGroup);
    
    List<Exercise> findByMuscleGroup(String muscleGroup);
    Set<Exercise> findByDifficultyLevel(Integer difficultyLevel);
    Set<Exercise> findByEquipmentRequiredContaining(String equipment);
    Set<Exercise> findByMuscleGroupAndDifficultyLevel(String muscleGroup, Integer difficultyLevel);

    @Query("SELECT DISTINCT e.muscleGroup FROM Exercise e")
    Set<String> findDistinctMuscleGroups();
}