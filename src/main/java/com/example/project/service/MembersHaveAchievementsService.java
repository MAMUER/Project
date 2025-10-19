package com.example.project.service;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.MembersHaveAchievements;
import com.example.project.repository.MembersHaveAchievementsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class MembersHaveAchievementsService {
    private final MembersHaveAchievementsRepository membersHaveAchievementsRepository;

    public Set<MembersHaveAchievements> getAchievementsByMember(Integer memberId) {
        return membersHaveAchievementsRepository.findByMemberIdMember(memberId);
    }

    public Set<MembersHaveAchievements> getAchievementsByAchievement(Integer achievementId) {
        return membersHaveAchievementsRepository.findByAchievementIdAchievement(achievementId);
    }

    public Set<MembersHaveAchievements> getAchievementsByDateRange(LocalDate startDate, LocalDate endDate) {
        return membersHaveAchievementsRepository.findByReceiptDateBetween(startDate, endDate);
    }

    public MembersHaveAchievements assignAchievementToMember(MembersHaveAchievements memberAchievement) {
        return membersHaveAchievementsRepository.save(memberAchievement);
    }

    public void removeAchievementFromMember(Integer memberId, Integer achievementId) {
        MembersHaveAchievements memberAchievement = membersHaveAchievementsRepository
                .findByMemberIdMemberAndAchievementIdAchievement(memberId, achievementId).orElse(null);
        if (memberAchievement != null) {
            membersHaveAchievementsRepository.delete(memberAchievement);
        }
    }
}