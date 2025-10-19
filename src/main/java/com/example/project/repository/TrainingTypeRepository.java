package com.example.project.repository;

import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.TrainingType;

@Repository
public interface TrainingTypeRepository extends JpaRepository<TrainingType, Integer> {

    Optional<TrainingType> findByTrainingTypeName(String trainingTypeName);

    Set<TrainingType> findByTrainingTypeNameContaining(String name);

    Set<TrainingType> findByWorkoutDescriptionIsNotNull();
}