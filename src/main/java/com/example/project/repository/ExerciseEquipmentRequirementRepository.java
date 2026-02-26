package com.example.project.repository;

import com.example.project.model.ExerciseEquipmentRequirement;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ExerciseEquipmentRequirementRepository extends JpaRepository<ExerciseEquipmentRequirement, Integer> {
    List<ExerciseEquipmentRequirement> findByExerciseIdExercise(Integer exerciseId);
    List<ExerciseEquipmentRequirement> findByEquipmentTypeIdEquipmentType(Integer equipmentTypeId);
}