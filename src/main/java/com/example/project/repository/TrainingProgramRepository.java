package com.example.project.repository;

import com.example.project.model.TrainingProgram;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface TrainingProgramRepository extends JpaRepository<TrainingProgram, Integer> {
    List<TrainingProgram> findByMemberIdMember(Integer memberId);

    List<TrainingProgram> findByMemberIdMemberAndIsActiveTrue(Integer memberId);

    TrainingProgram findByMemberIdMemberAndIsActiveTrueAndIdProgram(Integer memberId, Integer programId);

    @Query("SELECT DISTINCT tp FROM TrainingProgram tp " +
           "LEFT JOIN FETCH tp.programDays pd " +
           "LEFT JOIN FETCH pd.exercises pe " +
           "LEFT JOIN FETCH pe.exercise " +
           "LEFT JOIN FETCH tp.nutritionPlan " + // ДОБАВИТЬ: загружаем план питания
           "WHERE tp.member.idMember = :memberId")
    List<TrainingProgram> findByMemberIdWithDetails(@Param("memberId") Integer memberId);
}