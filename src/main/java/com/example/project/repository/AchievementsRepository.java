package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.Achievements;

@Repository
public interface AchievementsRepository extends JpaRepository<Achievements, Integer> {
    
    Set<Achievements> findByAchievementTitleContaining(String title);
    
    Set<Achievements> findByAchievementDescriptionIsNotNull();
}