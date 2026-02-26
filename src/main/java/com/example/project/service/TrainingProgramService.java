package com.example.project.service;

import com.example.project.model.ProgramDay;
import com.example.project.model.ProgramExercise;
import com.example.project.model.TrainingProgram;
import com.example.project.repository.TrainingProgramRepository;

import lombok.AllArgsConstructor;

import java.util.*;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@AllArgsConstructor
public class TrainingProgramService {
    private final TrainingProgramRepository trainingProgramRepository;

    @SuppressWarnings("null")
    @Transactional(readOnly = true)
    public TrainingProgram getProgram(Integer programId) {
        return trainingProgramRepository.findById(programId).orElse(null);
    }

    @Transactional(readOnly = true)
    public List<TrainingProgram> getMemberPrograms(Integer memberId) {
        // ИСПОЛЬЗУЕМ НОВЫЙ МЕТОД С JOIN FETCH
        return trainingProgramRepository.findByMemberIdWithDetails(memberId);
    }

    @Transactional(readOnly = true)
    public List<TrainingProgram> getActiveMemberPrograms(Integer memberId) {
        List<TrainingProgram> programs = getMemberPrograms(memberId);
        return programs.stream()
                .filter(TrainingProgram::getIsActive)
                .collect(Collectors.toList());
    }

    @SuppressWarnings("null")
    public TrainingProgram saveProgram(TrainingProgram program) {
        return trainingProgramRepository.save(program);
    }

    @Transactional
    public void deactivateOtherPrograms(Integer memberId, Integer activeProgramId) {
        List<TrainingProgram> programs = getMemberPrograms(memberId);
        for (TrainingProgram program : programs) {
            if (!program.getIdProgram().equals(activeProgramId)) {
                program.setIsActive(false);
                trainingProgramRepository.save(program);
            }
        }
    }

    public int getTotalExercisesCount(TrainingProgram program) {
        if (program.getProgramDays() == null) {
            return 0;
        }

        int totalExercises = 0;
        for (ProgramDay day : program.getProgramDays()) {
            if (day.getExercises() != null) {
                totalExercises += day.getExercises().size();
            }
        }
        return totalExercises;
    }

    // ДОБАВИТЬ: метод для получения программы с планом питания
    @SuppressWarnings("null")
    @Transactional(readOnly = true)
    public TrainingProgram getProgramWithNutritionPlan(Integer programId) {
        return trainingProgramRepository.findById(programId).orElse(null);
    }

    // В TrainingProgramService добавьте:
    public List<ProgramDay> getSortedProgramDays(TrainingProgram program) {
        if (program == null || program.getProgramDays() == null) {
            return new ArrayList<>();
        }

        return program.getProgramDays().stream()
                .filter(Objects::nonNull)
                .sorted(Comparator.comparing(ProgramDay::getDayNumber))
                .collect(Collectors.toList());
    }

    public List<ProgramExercise> getSortedExercises(ProgramDay day) {
        if (day == null || day.getExercises() == null) {
            return new ArrayList<>();
        }

        return day.getExercises().stream()
                .filter(Objects::nonNull)
                .sorted(Comparator.comparing(ProgramExercise::getOrderIndex))
                .collect(Collectors.toList());
    }
}