package com.example.project.service;

import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.InbodyAnalyses;
import com.example.project.repository.InbodyAnalysesRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class InbodyAnalysesService {
    private final InbodyAnalysesRepository inbodyAnalysesRepository;

    public InbodyAnalyses getInbodyAnalysis(Integer id) {
        return inbodyAnalysesRepository.findById(id).orElse(null);
    }

    public Set<InbodyAnalyses> getAnalysesByBmiRange(Float minBmi, Float maxBmi) {
        return inbodyAnalysesRepository.findByBmiBetween(minBmi, maxBmi);
    }

    public Set<InbodyAnalyses> getAnalysesByWeightRange(Float minWeight, Float maxWeight) {
        return inbodyAnalysesRepository.findByWeightBetween(minWeight, maxWeight);
    }

    public Set<InbodyAnalyses> getAnalysesByFatPercentRange(Float minFat, Float maxFat) {
        return inbodyAnalysesRepository.findByFatPercentBetween(minFat, maxFat);
    }

    public InbodyAnalyses saveInbodyAnalysis(InbodyAnalyses analysis) {
        return inbodyAnalysesRepository.save(analysis);
    }
}