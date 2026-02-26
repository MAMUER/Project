package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.VisitsHistory;

@Repository
public interface VisitsHistoryRepository extends JpaRepository<VisitsHistory, Integer> {
    
    Set<VisitsHistory> findByVisitDateBetween(LocalDate start, LocalDate end);
    
    Set<VisitsHistory> findByVisitDate(LocalDate visitDate);
    
    Set<VisitsHistory> findByVisitDateAfter(LocalDate startDate);
}