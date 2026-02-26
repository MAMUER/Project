package com.example.project.service;

import java.time.LocalDate;
import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Feedback;
import com.example.project.repository.FeedbackRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class FeedbackService {
    private final FeedbackRepository feedbackRepository;

    @SuppressWarnings("null")
    public Feedback getFeedback(Integer id) {
        return feedbackRepository.findById(id).orElse(null);
    }

    public List<Feedback> getAllFeedback() {
        return feedbackRepository.findAll();
    }

    public Set<Feedback> getFeedbackByRating(short minRating) {
        return feedbackRepository.findByRatingGreaterThanEqual(minRating);
    }

    public Set<Feedback> getFeedbackByDateRange(LocalDate startDate, LocalDate endDate) {
        return feedbackRepository.findByFeedbackDateBetween(startDate, endDate);
    }

    public Set<Feedback> searchFeedbackByText(String keyword) {
        return feedbackRepository.findByFeedbackTextContaining(keyword);
    }

    @SuppressWarnings("null")
    public Feedback saveFeedback(Feedback feedback) {
        return feedbackRepository.save(feedback);
    }

    public Set<Feedback> getFeedbackByUsername(String username) {
    return feedbackRepository.findByUsername(username); // ИСПРАВЛЕНО: findByUsername вместо findByMemberAccountUsername
}
}