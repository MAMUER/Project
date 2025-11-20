package com.example.project.service;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.project.model.*;
import com.example.project.model.Accounts.MembersAccounts;
import com.example.project.repository.MembersRepository;
import com.example.project.repository.TrainingScheduleRepository;
import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import java.time.LocalDate;
import java.util.*;

@Service
@Transactional
@AllArgsConstructor
public class MembersService {
    private final MembersRepository membersRepository;
    private final TrainingScheduleRepository trainingScheduleRepository;

    @PersistenceContext
    private EntityManager entityManager;

    // === Базовые CRUD операции ===
    public Members getMember(Integer id) {
        return membersRepository.findById(id).orElse(null);
    }

    public List<Members> getAllMembers() {
        return membersRepository.findAll();
    }

    public Set<Members> getMembersByClub(String clubName) {
        return membersRepository.findByClubClubName(clubName);
    }

    public Set<Members> searchMembersByName(String firstName, String secondName) {
        if (firstName != null && secondName != null) {
            return membersRepository.findByFirstNameContainingAndSecondNameContaining(firstName, secondName);
        } else if (firstName != null) {
            return membersRepository.findByFirstNameContaining(firstName);
        } else if (secondName != null) {
            return membersRepository.findBySecondNameContaining(secondName);
        }
        return Collections.emptySet();
    }

    // === Методы для работы с достижениями ===
    public Set<Achievements> getSetOfMemberAchievements(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        if (member != null) {
            Set<Achievements> achievements = new HashSet<>();
            for (MembersHaveAchievements memberAchievement : member.getMembersHaveAchievements()) {
                achievements.add(memberAchievement.getAchievement());
            }
            return achievements;
        }
        return Collections.emptySet();
    }

    // === Методы для работы с тренировками ===
    public Set<TrainingSchedule> getSetOfTrainingSchedule(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? member.getTrainingSchedules() : Collections.emptySet();
    }

    @Transactional(readOnly = true)
    public Members getMemberWithTrainings(Integer memberId) {
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));

        // Инициализируем ленивую коллекцию внутри транзакции
        member.getTrainingSchedules().size();

        return member;
    }

    @Transactional(readOnly = true)
    public Set<TrainingSchedule> getMemberTrainings(Integer memberId) {
        Members member = getMemberWithTrainings(memberId);
        return member.getTrainingSchedules();
    }

    @Transactional(readOnly = true)
    public boolean isMemberSignedUpForTraining(Integer memberId, Integer trainingId) {
        return membersRepository.existsTrainingForMember(memberId, trainingId);
    }

    @Transactional
    public void addTrainingToMember(Integer memberId, Integer trainingId) {
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));
        TrainingSchedule training = trainingScheduleRepository.findById(trainingId)
                .orElseThrow(() -> new RuntimeException("Training not found"));

        // Добавляем тренировку через коллекцию
        member.getTrainingSchedules().add(training);
        training.getMembers().add(member);

        membersRepository.save(member);
    }

    public void unsubscribeFromTraining(Integer memberId, Integer trainingId) {
        Members member = entityManager.find(Members.class, memberId);
        TrainingSchedule training = entityManager.find(TrainingSchedule.class, trainingId);

        if (member != null && training != null) {
            member.getTrainingSchedules().remove(training);
            training.getMembers().remove(member);

            entityManager.merge(member);
            entityManager.merge(training);
        }
    }

    // === Методы для работы с аккаунтом ===
    public MembersAccounts getMemberAccount(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? member.getMembersAccount() : null;
    }

    public String getPhotoUrl(int memberId) {
        MembersAccounts memberAccounts = getMemberAccount(memberId);
        try {
            return memberAccounts.getUserPhoto().getImageUrl();
        } catch (Exception e) {
            return "https://i.postimg.cc/Wbznd0qn/1674365371-3-34.jpg";
        }
    }

    // === Методы для работы с отзывами ===
    public Set<Feedback> getSetOfFeedbacks(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        if (member != null && member.getMembersAccount() != null) {
            // Получаем отзывы через аккаунт
            return new HashSet<>(member.getMembersAccount().getFeedbacks());
        }
        return Collections.emptySet();
    }

    // === Методы для работы с планами питания ===
    public Set<NutritionPlan> getNutritionPlans(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getNutritionPlans()) : Collections.emptySet();
    }

    public String getCurrentNutritionPlanDescription(int memberId) {
        Set<NutritionPlan> plans = getNutritionPlans(memberId);
        for (NutritionPlan plan : plans) {
            return plan.getNutritionDescription();
        }
        return null;
    }

    // === Методы для работы с историей посещений ===
    public Set<VisitsHistory> getSetOfVisits(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getVisitsHistory()) : Collections.emptySet();
    }

    public LocalDate getLastVisitDate(int memberId) {
        Set<VisitsHistory> visits = getSetOfVisits(memberId);
        if (!visits.isEmpty()) {
            // Используем Stream API для поиска максимальной даты
            return visits.stream()
                    .map(VisitsHistory::getVisitDate)
                    .max(LocalDate::compareTo)
                    .orElse(null);
        }
        return null;
    }

    // === Методы для работы с Inbody анализами ===
    public Set<InbodyAnalysis> getSetOfInbodyAnalyses(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getInbodyAnalysis()) : Collections.emptySet();
    }

    public String getGenderName(Integer gender) {
        if (gender == null)
            return "Не указан";
        switch (gender) {
            case 0:
                return "Женский";
            case 1:
                return "Мужской";
            case 2:
                return "Другой";
            default:
                return "Неизвестно";
        }
    }

    public String getMemberFullName(Integer memberId) {
        Members member = getMember(memberId);
        return member != null ? member.getFirstName() + " " + member.getSecondName() : "Неизвестный член";
    }

    // === Сохранение и удаление ===
    public Members save(Members member) {
        return membersRepository.save(member);
    }

    public void deleteMember(Integer id) {
        membersRepository.deleteById(id);
    }
}