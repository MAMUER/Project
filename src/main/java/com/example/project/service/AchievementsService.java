package com.example.project.service;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Achievements;
import com.example.project.repository.AchievementsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class AchievementsService {
    private final AchievementsRepository achievementsRepository;

    @SuppressWarnings("null")
    public Achievements getAchievement(Integer id) {
        return achievementsRepository.findById(id).orElse(null);
    }

    public List<Achievements> getAllAchievements() {
        return achievementsRepository.findAll();
    }

    public Set<Achievements> getAchievementsByTitle(String title) {
        return achievementsRepository.findByAchievementTitleContaining(title);
    }

    public Set<Achievements> getAchievementsWithDescription() {
        return achievementsRepository.findByAchievementDescriptionIsNotNull();
    }

    @SuppressWarnings("null")
    public Achievements saveAchievement(Achievements achievement) {
        return achievementsRepository.save(achievement);
    }

    @SuppressWarnings("null")
    public void deleteAchievement(Integer id) {
        achievementsRepository.deleteById(id);
    }
}