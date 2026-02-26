package com.example.project.service;

import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.*;
import com.example.project.repository.EquipmentRepository;
import com.example.project.repository.ExerciseRepository;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@AllArgsConstructor
@Transactional
public class ClubCapabilityService {
    private final ExerciseRepository exerciseRepository;
    private final EquipmentRepository equipmentRepository;
    
    public boolean canClubSupportExercise(String clubName, Integer exerciseId) {
        // Используем новый метод с JOIN FETCH
        Exercise exercise = exerciseRepository.findByIdWithEquipmentRequirements(exerciseId).orElse(null);
        if (exercise == null) {
            log.warn("Упражнение {} не найдено", exerciseId);
            return false;
        }

        // Теперь коллекция инициализирована
        Set<ExerciseEquipmentRequirement> requirements = exercise.getEquipmentRequirements();
        
        // Если требований нет - упражнение доступно
        if (requirements == null || requirements.isEmpty()) {
            return true;
        }

        // Проверяем каждое требование
        for (ExerciseEquipmentRequirement requirement : requirements) {
            if (requirement.getIsRequired()) {
                Integer availableQuantity = getAvailableEquipmentCount(
                    clubName, 
                    requirement.getEquipmentType().getIdEquipmentType()
                );
                
                if (availableQuantity < requirement.getQuantityRequired()) {
                    return false;
                }
            }
        }
        return true;
    }

    /**
     * Получает количество доступного оборудования определенного типа в клубе
     */
    public Integer getAvailableEquipmentCount(String clubName, Integer equipmentTypeId) {
        // Используем новый метод репозитория
        List<Equipment> equipment = equipmentRepository.findByClubNameAndEquipmentTypeId(clubName, equipmentTypeId);
        
        int totalQuantity = equipment.stream()
            .mapToInt(Equipment::getQuantity)
            .sum();
            
        log.trace("Клуб {}: оборудование типа {} - {} шт.", clubName, equipmentTypeId, totalQuantity);
        return totalQuantity;
    }

    /**
     * Получает список доступных упражнений для группы мышц в клубе
     */
    public List<Exercise> getAvailableExercisesByMuscleGroup(String clubName, String muscleGroup) {
        List<Exercise> allExercises = exerciseRepository.findByMuscleGroup(muscleGroup);
        
        List<Exercise> availableExercises = allExercises.stream()
            .filter(exercise -> canClubSupportExercise(clubName, exercise.getIdExercise()))
            .collect(Collectors.toList());
        
        return availableExercises;
    }

    public List<Exercise> getAllAvailableExercises(String clubName) {
        // Используем метод с JOIN FETCH
        List<Exercise> allExercises = exerciseRepository.findAllWithEquipmentRequirements();
        
        List<Exercise> availableExercises = allExercises.stream()
            .filter(exercise -> canClubSupportExercise(clubName, exercise.getIdExercise()))
            .collect(Collectors.toList());
        
        return availableExercises;
    }

    /**
     * Анализирует возможности клуба для различных целей тренировок
     */
    public Map<String, Object> analyzeClubCapabilities(String clubName) {
        Map<String, Object> analysis = new HashMap<>();
        
        List<Exercise> availableExercises = getAllAvailableExercises(clubName);
        
        // Анализ по группам мышц
        Map<String, Long> exercisesByMuscleGroup = availableExercises.stream()
            .collect(Collectors.groupingBy(
                Exercise::getMuscleGroup,
                Collectors.counting()
            ));
        
        analysis.put("exercisesByMuscleGroup", exercisesByMuscleGroup);
        
        // Анализ по уровням сложности
        Map<Integer, Long> exercisesByDifficulty = availableExercises.stream()
            .collect(Collectors.groupingBy(
                Exercise::getDifficultyLevel,
                Collectors.counting()
            ));
        
        analysis.put("exercisesByDifficulty", exercisesByDifficulty);
        
        // Общая статистика
        analysis.put("totalAvailableExercises", availableExercises.size());
        analysis.put("totalExercisesInDatabase", exerciseRepository.findAll().size());
        
        // Анализ оборудования
        analysis.put("equipmentSummary", getClubEquipmentSummary(clubName));
        
        return analysis;
    }

    /**
     * Получает оборудование клуба сгруппированное по типам
     */
    public Map<String, Integer> getClubEquipmentSummary(String clubName) {
        List<Equipment> clubEquipment = equipmentRepository.findByClubName(clubName);
        
        Map<String, Integer> equipmentSummary = clubEquipment.stream()
            .collect(Collectors.groupingBy(
                equipment -> equipment.getEquipmentType().getTypeName(),
                Collectors.summingInt(Equipment::getQuantity)
            ));
        return equipmentSummary;
    }

    /**
     * Получает рекомендации по улучшению возможностей клуба
     */
    @SuppressWarnings("null")
    public Map<String, Object> getClubImprovementRecommendations(String clubName) {
        Map<String, Object> recommendations = new HashMap<>();
        List<String> improvementSuggestions = new ArrayList<>();
        
        List<Exercise> unavailableExercises = exerciseRepository.findAll().stream()
            .filter(exercise -> !canClubSupportExercise(clubName, exercise.getIdExercise()))
            .collect(Collectors.toList());
        
        // Анализируем, какое оборудование нужно добавить
        Map<String, Integer> missingEquipment = new HashMap<>();
        
        for (Exercise exercise : unavailableExercises) {
            Set<ExerciseEquipmentRequirement> requirements = exercise.getEquipmentRequirements();
            if (requirements != null) {
                for (ExerciseEquipmentRequirement requirement : requirements) {
                    if (requirement.getIsRequired()) {
                        String equipmentType = requirement.getEquipmentType().getTypeName();
                        Integer available = getAvailableEquipmentCount(clubName, requirement.getEquipmentType().getIdEquipmentType());
                        Integer required = requirement.getQuantityRequired();
                        
                        if (available < required) {
                            int needed = required - available;
                            missingEquipment.merge(equipmentType, needed, Integer::sum);
                        }
                    }
                }
            }
        }
        
        if (!missingEquipment.isEmpty()) {
            improvementSuggestions.add("Рекомендуется добавить оборудование:");
            missingEquipment.forEach((equipment, quantity) -> {
                improvementSuggestions.add(String.format("• %s: +%d шт.", equipment, quantity));
            });
        }
        
        recommendations.put("suggestions", improvementSuggestions);
        recommendations.put("unavailableExercisesCount", unavailableExercises.size());
        recommendations.put("missingEquipment", missingEquipment);
        
        return recommendations;
    }
}