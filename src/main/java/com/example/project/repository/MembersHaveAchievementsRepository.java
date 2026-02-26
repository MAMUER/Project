package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.MembersHaveAchievements;

@Repository
public interface MembersHaveAchievementsRepository
        extends JpaRepository<MembersHaveAchievements, com.example.project.model.MembersHaveAchievementsId> {

    Set<MembersHaveAchievements> findByMemberIdMember(Integer memberId);

    Set<MembersHaveAchievements> findByAchievementIdAchievement(Integer achievementId);

    Set<MembersHaveAchievements> findByReceiptDateBetween(LocalDate start, LocalDate end);

    Optional<MembersHaveAchievements> findByMemberIdMemberAndAchievementIdAchievement(Integer memberId,
            Integer achievementId);
}