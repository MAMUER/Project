package com.example.project.repository;

import java.time.LocalDateTime;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import com.example.project.model.TrainingSchedule;

@Repository
public interface TrainingScheduleRepository
                extends JpaRepository<TrainingSchedule, Integer>, 
                                                   JpaSpecificationExecutor<TrainingSchedule> {

        Set<TrainingSchedule> findByTrainerIdTrainer(Integer trainerId);

        Set<TrainingSchedule> findByTrainingTypeIdTrainingType(Integer trainingTypeId);

        Set<TrainingSchedule> findBySessionDateBetween(LocalDateTime start, LocalDateTime end);

        Set<TrainingSchedule> findBySessionTimeGreaterThan(Integer minTime);

        Set<TrainingSchedule> findByMembersIdMember(Integer memberId);
}