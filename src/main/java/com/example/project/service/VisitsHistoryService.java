package com.example.project.service;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.VisitsHistory;
import com.example.project.repository.VisitsHistoryRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class VisitsHistoryService {
    private final VisitsHistoryRepository visitsHistoryRepository;

    @SuppressWarnings("null")
    public VisitsHistory getVisitHistory(Integer id) {
        return visitsHistoryRepository.findById(id).orElse(null);
    }

    public Set<VisitsHistory> getVisitsByDateRange(LocalDate startDate, LocalDate endDate) {
        return visitsHistoryRepository.findByVisitDateBetween(startDate, endDate);
    }

    public Set<VisitsHistory> getVisitsByDate(LocalDate visitDate) {
        return visitsHistoryRepository.findByVisitDate(visitDate);
    }

    public Set<VisitsHistory> getRecentVisits(LocalDate startDate) {
        return visitsHistoryRepository.findByVisitDateAfter(startDate);
    }

    @SuppressWarnings("null")
    public VisitsHistory saveVisit(VisitsHistory visit) {
        return visitsHistoryRepository.save(visit);
    }

    @SuppressWarnings("null")
    public void deleteVisit(Integer id) {
        visitsHistoryRepository.deleteById(id);
    }
}