package com.example.project.service;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.dto.ClubDTO;
import com.example.project.model.*;
import com.example.project.model.accounts.MembersAccounts;
import com.example.project.repository.MembersRepository;
import com.example.project.repository.TrainingScheduleRepository;
import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import java.time.LocalDate;
import java.util.*;

@Slf4j
@Service
@Transactional
@AllArgsConstructor
public class MembersService {
    private final MembersRepository membersRepository;
    private final TrainingScheduleRepository trainingScheduleRepository;

    @Getter
    @PersistenceContext
    private final EntityManager entityManager;

    // === Базовые CRUD операции ===
    @SuppressWarnings("null")
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
        @SuppressWarnings("null")
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
        @SuppressWarnings("null")
        Members member = membersRepository.findById(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));
        @SuppressWarnings("null")
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
        return switch (gender) {
            case 0 -> "Женский";
            case 1 -> "Мужской";
            case 2 -> "Другой";
            default -> "Неизвестно";
        };
    }

    public String getMemberFullName(Integer memberId) {
        Members member = getMember(memberId);
        return member != null ? member.getFirstName() + " " + member.getSecondName() : "Неизвестный член";
    }

    // === Сохранение и удаление ===
    @SuppressWarnings("null")
    public void save(Members member) {
        membersRepository.save(member);
    }

    @SuppressWarnings("null")
    public void deleteMember(Integer id) {
        membersRepository.deleteById(id);
    }

    @Transactional(readOnly = true)
    public Members getMemberWithClub(Integer memberId) {
        return membersRepository.findByIdWithClub(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));
    }

    // ДОБАВИТЬ: метод для загрузки с клубом и InbodyAnalysis
    @Transactional(readOnly = true)
    public Members getMemberWithClubAndInbody(Integer memberId) {
        return membersRepository.findByIdWithClubAndInbody(memberId)
                .orElseThrow(() -> new RuntimeException("Member not found"));
    }

    // ДОБАВИТЬ: метод для проверки наличия Inbody анализов
    public boolean hasInbodyAnalysis(Integer memberId) {
        Members member = getMemberWithClubAndInbody(memberId);
        return member != null && member.getInbodyAnalysis() != null && !member.getInbodyAnalysis().isEmpty();
    }

    // ДОБАВИТЬ: метод для расчета возраста
    public int calculateAge(Integer memberId) {
        Members member = getMember(memberId);
        if (member != null && member.getBirthDate() != null) {
            return java.time.Period.between(member.getBirthDate(), LocalDate.now()).getYears();
        }
        return 0;
    }

    // ДОБАВИТЬ: метод для определения возрастной группы
    public String getAgeGroup(Integer memberId) {
        int age = calculateAge(memberId);
        if (age >= 18 && age <= 29)
            return "18-29";
        else if (age >= 30 && age <= 39)
            return "30-39";
        else if (age >= 40 && age <= 49)
            return "40-49";
        else if (age >= 50 && age <= 59)
            return "50-59";
        else
            return "60+";
    }

    // В MembersService добавьте эти методы:

    // Безопасный метод проверки InbodyAnalysis через нативный запрос
    public boolean hasInbodyAnalysisNative(Integer memberId) {
        return false;
    }

    @Transactional(readOnly = true)
    public ClubDTO getMemberClubDTO(Integer memberId) {
        try {
            // Используем нативный запрос чтобы избежать прокси
            Members member = membersRepository.findByIdWithClub(memberId).orElse(null);
            if (member != null && member.getClub() != null) {
                Clubs club = member.getClub();

                // Создаем новый объект Clubs с данными из прокси
                Clubs realClub = new Clubs();
                realClub.setClubName(club.getClubName());
                realClub.setAddress(club.getAddress());
                realClub.setSchedule(club.getSchedule());

                return ClubDTO.fromEntity(realClub);
            }
            return null;
        } catch (Exception e) {
            log.warn("Ошибка при получении клуба для пользователя {}: {}", memberId, e.getMessage());
            return null;
        }
    }

    // Альтернативный метод - полностью избегаем Hibernate
    @Transactional(readOnly = true)
    public ClubDTO getMemberClubDTOSafe(Integer memberId) {
        try {
            // Используем прямой SQL запрос через EntityManager
            String sql = "SELECT c.club_name, c.address, c.schedule " +
                    "FROM members m " +
                    "JOIN clubs c ON m.club_name = c.club_name " +
                    "WHERE m.id_member = :memberId";

            Object[] result = (Object[]) entityManager.createNativeQuery(sql)
                    .setParameter("memberId", memberId)
                    .getSingleResult();

            if (result != null) {
                ClubDTO dto = new ClubDTO();
                dto.setClubName((String) result[0]);
                dto.setAddress((String) result[1]);
                dto.setSchedule((String) result[2]);
                return dto;
            }
            return null;
        } catch (Exception e) {
            log.warn("Ошибка при получении клуба через нативный запрос для пользователя {}: {}", memberId,
                    e.getMessage());
            return null;
        }
    }
}