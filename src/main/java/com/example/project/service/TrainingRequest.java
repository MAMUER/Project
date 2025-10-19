package com.example.project.service;

import java.time.LocalDateTime;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
public class TrainingRequest {
    private Integer memberId;
    private Integer trainerId;
    private Integer trainingId;
    private Integer trainingTypeId;
    private LocalDateTime sessionDate;
    private Integer sessionTime;
    private String trainingDate; // Для обратной совместимости

    // Конструкторы для разных сценариев использования
    public TrainingRequest(Integer memberId, Integer trainingId) {
        this.memberId = memberId;
        this.trainingId = trainingId;
    }

    public TrainingRequest(Integer trainerId, Integer trainingTypeId,
                           LocalDateTime sessionDate, Integer sessionTime) {
        this.trainerId = trainerId;
        this.trainingTypeId = trainingTypeId;
        this.sessionDate = sessionDate;
        this.sessionTime = sessionTime;
    }

    public TrainingRequest(int memberId, int trainerId, String trainingDate) {
        this.memberId = memberId;
        this.trainerId = trainerId;
        this.trainingDate = trainingDate;
    }

    // Вспомогательные методы
    public boolean hasMemberId() {
        return memberId != null;
    }

    public boolean hasTrainerId() {
        return trainerId != null;
    }

    public boolean hasTrainingId() {
        return trainingId != null;
    }

    public boolean hasSessionDate() {
        return sessionDate != null;
    }
}