package com.example.project.service;

import lombok.AllArgsConstructor;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.TrainingType;
import com.example.project.repository.TrainingTypeRepository;
@Service
@AllArgsConstructor
public class TrainingTypeService {
    private final TrainingTypeRepository trainingTypeRepository;
    
    @SuppressWarnings("null")
    public TrainingType getTrainingType(Integer trainingTypeId) {
        return trainingTypeRepository.findById(trainingTypeId).orElse(null);
    }

    public List<TrainingType> getAllTrainingTypes() {
        return trainingTypeRepository.findAll();
    }

    public TrainingType getTrainingTypeByName(String name) {
        return trainingTypeRepository.findByTrainingTypeName(name).orElse(null);
    }

    public Set<TrainingType> getTrainingTypesWithDescription() {
        return trainingTypeRepository.findByWorkoutDescriptionIsNotNull();
    }

    @SuppressWarnings("null")
    public TrainingType saveTrainingType(TrainingType trainingType) {
        return trainingTypeRepository.save(trainingType);
    }
}