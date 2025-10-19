package com.example.project.service;

import lombok.AllArgsConstructor;

import org.hibernate.Hibernate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.Achievements;
import com.example.project.model.EquipmentStatistics;
import com.example.project.model.Feedback;
import com.example.project.model.InbodyAnalyses;
import com.example.project.model.Members;
import com.example.project.model.MembersHaveAchievements;
import com.example.project.model.NutritionPlan;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.VisitsHistory;
import com.example.project.model.Accounts.MembersAccounts;
import com.example.project.repository.MembersRepository;
import com.example.project.repository.TrainingScheduleRepository;

import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

@Service
@Transactional
@AllArgsConstructor
public class MembersService {
    private final MembersRepository membersRepository;
    private final TrainingScheduleRepository trainingScheduleRepository;

    @PersistenceContext
    private EntityManager entityManager;

    public Members getMember(Integer id) {
        return membersRepository.findById(id).orElse(null);
    }

    public List<Members> getAllMembers() {
        return membersRepository.findAll();
    }

    public Set<Members> getMembersByClub(String clubName) {
        return membersRepository.findByClubClubName(clubName);
    }

    public Set<Members> getActiveMembers() {
        return membersRepository.findByEndTrialDateIsNullOrEndTrialDateAfter(LocalDate.now());
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

    public Set<TrainingSchedule> getSetOfTrainingSchedule(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? member.getTrainingSchedules() : Collections.emptySet();
    }

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

    public Set<Feedback> getSetOfFeedbacks(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        if (member != null && member.getMembersAccount() != null) {
            return new HashSet<>(member.getMembersAccount().getFeedbacks());
        }
        return Collections.emptySet();
    }

    public Set<NutritionPlan> getNutritionPlans(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getNutritionPlans()) : Collections.emptySet();
    }

    public String getCurrentNutritionPlanDescription(int memberId) {
        Set<NutritionPlan> plans = getNutritionPlans(memberId);
        for (NutritionPlan plan : plans) {
            if (plan.getEndDate() == null || plan.getEndDate().isAfter(LocalDate.now())) {
                return plan.getNutritionDescription();
            }
        }
        return null;
    }

    public Set<VisitsHistory> getSetOfVisits(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getVisitsHistory()) : Collections.emptySet();
    }

    public LocalDate getLastVisitDate(int memberId) {
        List<VisitsHistory> visits = getListOfVisits(memberId);
        if (!visits.isEmpty()) {
            visits.sort(Comparator.comparing(VisitsHistory::getVisitDate).reversed());
            return visits.get(0).getVisitDate();
        }
        return null;
    }

    private List<VisitsHistory> getListOfVisits(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new ArrayList<>(member.getVisitsHistory()) : Collections.emptyList();
    }

    public String getMemberRolename(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? member.getMembershipRole().getRoleName() : null;
    }

    public Set<InbodyAnalyses> getSetOfInbodyAnalyses(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getInbodyAnalyses()) : Collections.emptySet();
    }

    public Set<EquipmentStatistics> getSetOfEquipmentStatistics(int memberId) {
        Members member = membersRepository.findById(memberId).orElse(null);
        return member != null ? new HashSet<>(member.getEquipmentStatistics()) : Collections.emptySet();
    }

    public String getActivityName(int memberId, int statisticsId) {
        Set<EquipmentStatistics> equipmentStatistics = getSetOfEquipmentStatistics(memberId);
        for (EquipmentStatistics equipmentStatistic : equipmentStatistics) {
            if (equipmentStatistic.getIdStatistics() == statisticsId) {
                return equipmentStatistic.getActivityType().getActivityName();
            }
        }
        return null;
    }

    public Members save(Members member) {
        return membersRepository.save(member);
    }

    public void deleteMember(Integer id) {
        membersRepository.deleteById(id);
    }

    @Transactional
    public void addTrainingToMember(Integer memberId, Integer trainingId) {
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));
        TrainingSchedule training = trainingScheduleRepository.findById(trainingId)
                .orElseThrow(() -> new RuntimeException("Training not found"));

        member.addTraining(training);
        membersRepository.save(member);
    }

    @Transactional(readOnly = true)
    public Members getMemberWithTrainings(Integer memberId) {
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));

        // Инициализируем ленивую коллекцию внутри транзакции
        member.getTrainingSchedules().size(); // Это загрузит коллекцию

        return member;
    }

    @Transactional(readOnly = true)
    public Set<TrainingSchedule> getMemberTrainings(Integer memberId) {
        Members member = getMemberWithTrainings(memberId);
        return member.getTrainingSchedules();
    }

    @Transactional(readOnly = true)
    public boolean isMemberSignedUpForTraining(Integer memberId, Integer trainingId) {
        // Используем нативный запрос для проверки связи в промежуточной таблице
        return membersRepository.existsTrainingForMember(memberId, trainingId);
    }

    @Transactional(readOnly = true)
    public boolean isMemberSignedUpForTraining(Integer memberId, TrainingSchedule training) {
        return isMemberSignedUpForTraining(memberId, training.getIdSession());
    }

    @Transactional(readOnly = true)
    public Set<TrainingSchedule> getTrainingSchedules(Integer memberId) {
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));

        // Явно инициализируем коллекцию внутри транзакции
        Hibernate.initialize(member.getTrainingSchedules());

        return member.getTrainingSchedules();
    }

    public void unsubscribeFromTraining(Integer memberId, Integer trainingId) {
        Members member = entityManager.find(Members.class, memberId);
        TrainingSchedule training = entityManager.find(TrainingSchedule.class, trainingId);

        if (member != null && training != null) {
            // Работаем с инициализированной коллекцией в транзакции
            member.getTrainingSchedules().remove(training);
            training.getMembers().remove(member);

            entityManager.merge(member);
            entityManager.merge(training);
        }
    }
}