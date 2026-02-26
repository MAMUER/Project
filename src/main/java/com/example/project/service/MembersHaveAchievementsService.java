package com.example.project.service;

import java.time.LocalDate;
import java.util.Optional;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.MembersHaveAchievements;
import com.example.project.model.MembersHaveAchievementsId;
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

    // Улучшенный метод для назначения достижения
    public MembersHaveAchievements assignAchievementToMember(Integer memberId, Integer achievementId, LocalDate receiptDate) {
        // Проверяем, нет ли уже такой связи
        Optional<MembersHaveAchievements> existing = membersHaveAchievementsRepository
                .findByMemberIdMemberAndAchievementIdAchievement(memberId, achievementId);
        
        if (existing.isPresent()) {
            // Обновляем дату получения, если достижение уже есть
            MembersHaveAchievements memberAchievement = existing.get();
            memberAchievement.setReceiptDate(receiptDate);
            return membersHaveAchievementsRepository.save(memberAchievement);
        } else {
            // Создаем новую связь
            MembersHaveAchievementsId id = new MembersHaveAchievementsId(memberId, achievementId);
            MembersHaveAchievements memberAchievement = new MembersHaveAchievements();
            memberAchievement.setId(id);
            memberAchievement.setReceiptDate(receiptDate);
            // Note: member и achievement должны быть установлены через отдельные сервисы
            return membersHaveAchievementsRepository.save(memberAchievement);
        }
    }

    public void removeAchievementFromMember(Integer memberId, Integer achievementId) {
        membersHaveAchievementsRepository
                .findByMemberIdMemberAndAchievementIdAchievement(memberId, achievementId).ifPresent(membersHaveAchievementsRepository::delete);
    }

    // Новые полезные методы
    public boolean hasAchievement(Integer memberId, Integer achievementId) {
        return membersHaveAchievementsRepository
                .findByMemberIdMemberAndAchievementIdAchievement(memberId, achievementId)
                .isPresent();
    }

    public long countAchievementsByMember(Integer memberId) {
        return membersHaveAchievementsRepository.findByMemberIdMember(memberId).size();
    }

    public LocalDate getAchievementReceiptDate(Integer memberId, Integer achievementId) {
        return membersHaveAchievementsRepository
                .findByMemberIdMemberAndAchievementIdAchievement(memberId, achievementId)
                .map(MembersHaveAchievements::getReceiptDate)
                .orElse(null);
    }
}