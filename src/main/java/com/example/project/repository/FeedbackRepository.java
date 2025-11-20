package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import com.example.project.model.Feedback;

@Repository
public interface FeedbackRepository extends JpaRepository<Feedback, Integer> {

    Set<Feedback> findByRatingGreaterThanEqual(short minRating);
    Set<Feedback> findByRatingLessThanEqual(short maxRating);
    
    // ИСПРАВЛЕНО: используем поле username (String) вместо связи
    Set<Feedback> findByUsername(String username);
    
    Set<Feedback> findByFeedbackDateBetween(LocalDate startDate, LocalDate endDate);
    Set<Feedback> findByFeedbackTextContaining(String keyword);
}