package com.example.project.service;

import com.example.project.model.InbodyAnalysis;
import com.example.project.repository.InbodyAnalysisRepository;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@AllArgsConstructor
public class InbodyAnalysisService {
    private final InbodyAnalysisRepository inbodyAnalysisRepository;

    // === Базовые CRUD операции ===
    @SuppressWarnings("null")
    public InbodyAnalysis getInbodyAnalysis(Integer id) {
        return inbodyAnalysisRepository.findById(id).orElse(null);
    }

    @SuppressWarnings("null")
    public InbodyAnalysis saveInbodyAnalysis(InbodyAnalysis analysis) {
        return inbodyAnalysisRepository.save(analysis);
    }

    @SuppressWarnings("null")
    public void deleteInbodyAnalysis(Integer id) {
        inbodyAnalysisRepository.deleteById(id);
    }

    public List<InbodyAnalysis> getAllAnalyses() {
        return inbodyAnalysisRepository.findAll();
    }

    // === Методы поиска по диапазонам ===
    public Set<InbodyAnalysis> getAnalysesByBmiRange(Float minBmi, Float maxBmi) {
        return inbodyAnalysisRepository.findByBmiBetween(minBmi, maxBmi);
    }

    public Set<InbodyAnalysis> getAnalysesByWeightRange(Float minWeight, Float maxWeight) {
        return inbodyAnalysisRepository.findByWeightBetween(minWeight, maxWeight);
    }

    public Set<InbodyAnalysis> getAnalysesByFatPercentRange(Float minFat, Float maxFat) {
        return inbodyAnalysisRepository.findByFatPercentBetween(minFat, maxFat);
    }

    public Set<InbodyAnalysis> getAnalysesByMusclePercentRange(Float minMuscle, Float maxMuscle) {
        return inbodyAnalysisRepository.findByMusclePercentBetween(minMuscle, maxMuscle);
    }

    // === Бизнес-логика ===
    public InbodyAnalysis createAnalysis(Float height, Float weight, Float fatPercent, Float musclePercent) {
        InbodyAnalysis analysis = new InbodyAnalysis();
        analysis.setHeight(height);
        analysis.setWeight(weight);
        analysis.setFatPercent(fatPercent);
        analysis.setMusclePercent(musclePercent);
        analysis.setBmi(calculateBMI(height, weight));
        
        return saveInbodyAnalysis(analysis);
    }

    public Float calculateBMI(Float height, Float weight) {
        if (height == null || height <= 0 || weight == null || weight <= 0) {
            return null;
        }
        float heightInMeters = height / 100;
        return weight / (heightInMeters * heightInMeters);
    }

    public String getBMICategory(Float bmi) {
        if (bmi == null) return "Не рассчитан";
        if (bmi < 18.5) return "Недостаточный вес";
        if (bmi < 25) return "Нормальный вес";
        if (bmi < 30) return "Избыточный вес";
        return "Ожирение";
    }

    public String getFatCategory(Float fatPercent) {
        if (fatPercent == null) return "Не указан";
        if (fatPercent < 6) return "Очень низкий";
        if (fatPercent < 14) return "Спортивный";
        if (fatPercent < 18) return "Фитнес";
        if (fatPercent < 25) return "Средний";
        return "Высокий";
    }

    // === Статистические методы ===
    public Map<String, Double> getAverageStatistics() {
        Map<String, Double> statistics = new HashMap<>();
        
        statistics.put("averageBmi", inbodyAnalysisRepository.findAverageBmi());
        statistics.put("averageWeight", inbodyAnalysisRepository.findAverageWeight());
        statistics.put("averageFatPercent", inbodyAnalysisRepository.findAverageFatPercent());
        statistics.put("averageMusclePercent", inbodyAnalysisRepository.findAverageMusclePercent());
        
        return statistics;
    }

    public Map<String, Long> getBmiCategoryStatistics() {
        List<InbodyAnalysis> allAnalyses = inbodyAnalysisRepository.findAll();
        
        return allAnalyses.stream()
                .filter(analysis -> analysis.getBmi() != null)
                .collect(Collectors.groupingBy(
                    analysis -> getBMICategory(analysis.getBmi()),
                    Collectors.counting()
                ));
    }

    public Map<String, Long> getFatCategoryStatistics() {
        List<InbodyAnalysis> allAnalyses = inbodyAnalysisRepository.findAll();
        
        return allAnalyses.stream()
                .filter(analysis -> analysis.getFatPercent() != null)
                .collect(Collectors.groupingBy(
                    analysis -> getFatCategory(analysis.getFatPercent()),
                    Collectors.counting()
                ));
    }

    // === Валидация ===
    public boolean isValidAnalysisData(Float height, Float weight, Float fatPercent, Float musclePercent) {
        return (height == null || (height > 0 && height < 300)) &&
               (weight == null || (weight > 0 && weight < 500)) &&
               (fatPercent == null || (fatPercent >= 0 && fatPercent <= 100)) &&
               (musclePercent == null || (musclePercent >= 0 && musclePercent <= 100));
    }

    public List<String> validateAnalysis(InbodyAnalysis analysis) {
        List<String> errors = new ArrayList<>();
        
        if (analysis.getHeight() != null && (analysis.getHeight() <= 0 || analysis.getHeight() > 300)) {
            errors.add("Рост должен быть между 0 и 300 см");
        }
        
        if (analysis.getWeight() != null && (analysis.getWeight() <= 0 || analysis.getWeight() > 500)) {
            errors.add("Вес должен быть между 0 и 500 кг");
        }
        
        if (analysis.getFatPercent() != null && (analysis.getFatPercent() < 0 || analysis.getFatPercent() > 100)) {
            errors.add("Процент жира должен быть между 0 и 100");
        }
        
        if (analysis.getMusclePercent() != null && (analysis.getMusclePercent() < 0 || analysis.getMusclePercent() > 100)) {
            errors.add("Процент мышц должен быть между 0 и 100");
        }
        
        return errors;
    }
}