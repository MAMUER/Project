package com.example.project.service;

import lombok.AllArgsConstructor;

import org.springframework.data.jpa.domain.Specification;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.Members;
import com.example.project.model.Trainers;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.TrainingType;
import com.example.project.repository.TrainingScheduleRepository;
import com.example.project.repository.TrainingTypeRepository;

import jakarta.persistence.criteria.Join;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import java.util.stream.Collectors;

@Service
@AllArgsConstructor
@Transactional
public class TrainingScheduleService {
    private final TrainingScheduleRepository trainingScheduleRepository;
    private final MembersService membersService;
    private final TrainingTypeRepository trainingTypeRepository;
    private final CustomUserDetailsService userDetailsService;

    @SuppressWarnings("null")
    public TrainingSchedule getTrainingSchedule(Integer id) {
        return trainingScheduleRepository.findById(id).orElse(null);
    }

    public List<TrainingSchedule> getAllTrainingSchedules() {
        return trainingScheduleRepository.findAll();
    }

    public Set<TrainingSchedule> getTrainingSchedulesByTrainer(Integer trainerId) {
        return trainingScheduleRepository.findByTrainerIdTrainer(trainerId);
    }

    public Set<TrainingSchedule> getTrainingSchedulesByTrainingType(Integer trainingTypeId) {
        return trainingScheduleRepository.findByTrainingTypeIdTrainingType(trainingTypeId);
    }

    public Set<TrainingSchedule> getTrainingSchedulesByDateRange(LocalDateTime start, LocalDateTime end) {
        return trainingScheduleRepository.findBySessionDateBetween(start, end);
    }

    public Set<TrainingSchedule> getTrainingSet(Set<Integer> trainerId,
            Set<Integer> trainingTypeId,
            LocalDateTime sessionDateStart,
            LocalDateTime sessionDateEnd,
            Integer sessionTimeStart,
            Integer sessionTimeEnd) {

        Specification<TrainingSchedule> spec = Specification.unrestricted();

        // Фильтр по тренерам
        if (trainerId != null && !trainerId.isEmpty()) {
            spec = spec.and((root, query, cb) -> {
                Join<TrainingSchedule, Trainers> trainerJoin = root.join("trainer");
                return trainerJoin.get("idTrainer").in(trainerId);
            });
        }

        // Фильтр по типам тренировок
        if (trainingTypeId != null && !trainingTypeId.isEmpty()) {
            spec = spec.and((root, query, cb) -> {
                Join<TrainingSchedule, TrainingType> typeJoin = root.join("trainingType");
                return typeJoin.get("idTrainingType").in(trainingTypeId);
            });
        }

        // Фильтр по дате начала
        if (sessionDateStart != null) {
            spec = spec.and((root, query, cb) -> cb.greaterThanOrEqualTo(root.get("sessionDate"), sessionDateStart));
        }

        // Фильтр по дате окончания
        if (sessionDateEnd != null) {
            spec = spec.and((root, query, cb) -> cb.lessThanOrEqualTo(root.get("sessionDate"), sessionDateEnd));
        }

        // Фильтр по времени начала (продолжительности)
        if (sessionTimeStart != null && sessionTimeStart > 0) {
            spec = spec.and((root, query, cb) -> cb.greaterThanOrEqualTo(root.get("sessionTime"), sessionTimeStart));
        }

        // Фильтр по времени окончания (продолжительности)
        if (sessionTimeEnd != null && sessionTimeEnd > 0) {
            spec = spec.and((root, query, cb) -> cb.lessThanOrEqualTo(root.get("sessionTime"), sessionTimeEnd));
        }

        return new HashSet<>(trainingScheduleRepository.findAll(spec));
    }

    public Set<Event> trainingScheduleToEventSet(Set<TrainingSchedule> trainingScheduleSet, Integer trainerSchedule) {
        Set<Event> eventsSet = new HashSet<>();

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);

        switch (role) {
            case "member" -> {
                Integer memberId = userDetailsService.getUserId(username);
                Members member = membersService.getMember(memberId);

                // Получаем ID тренировок члена для быстрого сравнения
                Set<Integer> memberTrainingIds = new HashSet<>();
                if (member != null && member.getTrainingSchedules() != null) {
                    member.getTrainingSchedules().forEach(training -> memberTrainingIds.add(training.getIdSession()));
                }

                for (TrainingSchedule training : trainingScheduleSet) {
                    // Пропускаем персональные тренировки, если член не записан и это не расписание
                    // тренера
                    if (training.getTrainingType().getIdTrainingType() == 5
                            && !memberTrainingIds.contains(training.getIdSession())
                            && !Integer.valueOf(1).equals(trainerSchedule)) {
                        continue;
                    }

                    String color = memberTrainingIds.contains(training.getIdSession()) ? "#3e4684" : "#b2b4d4";

                    helpMethod(eventsSet, training, color);
                }
            }

            case "trainer" -> {
                for (TrainingSchedule training : trainingScheduleSet) {
                    String color = "#3e4684";

                    helpMethod(eventsSet, training, color);
                }
            }

            default -> {
            }
        }

        return eventsSet;
    }

    private void helpMethod(Set<Event> eventsSet, TrainingSchedule training, String color) {
        LocalDateTime endTime = training.getSessionDate().plusMinutes(training.getSessionTime());
        String trainingType = training.getTrainingType().getTrainingTypeName();
        String trainingDateStart = training.getSessionDate().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        String trainingDateEnd = endTime.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        int sessionId = training.getIdSession();

        eventsSet.add(new Event(trainingType, trainingDateStart, trainingDateEnd, sessionId, color));
    }

    /**
     * Альтернативная реализация с использованием Stream API
     */
    public Set<Event> convertTrainingToEvents(Set<TrainingSchedule> trainingScheduleSet, Integer trainerSchedule) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);

        if ("member".equals(role)) {
            Integer memberId = userDetailsService.getUserId(username);
            Members member = membersService.getMember(memberId);

            Set<Integer> memberTrainingIds = member.getTrainingSchedules().stream()
                    .map(TrainingSchedule::getIdSession)
                    .collect(Collectors.toSet());

            return trainingScheduleSet.stream()
                    .filter(training -> !isPersonalTrainingFiltered(training, memberTrainingIds, trainerSchedule))
                    .map(training -> createTrainingEvent(training,
                            memberTrainingIds.contains(training.getIdSession())))
                    .collect(Collectors.toSet());

        } else if ("trainer".equals(role)) {
            return trainingScheduleSet.stream()
                    .map(training -> createTrainingEvent(training, true))
                    .collect(Collectors.toSet());
        }

        return Collections.emptySet();
    }

    private boolean isPersonalTrainingFiltered(TrainingSchedule training, Set<Integer> memberTrainingIds,
            Integer trainerSchedule) {
        return training.getTrainingType().getIdTrainingType() == 5
                && !memberTrainingIds.contains(training.getIdSession())
                && !Integer.valueOf(1).equals(trainerSchedule);
    }

    private Event createTrainingEvent(TrainingSchedule training, boolean isMemberTraining) {
        String color = isMemberTraining ? "#3e4684" : "#b2b4d4";
        LocalDateTime endTime = training.getSessionDate().plusMinutes(training.getSessionTime());

        String trainingType = training.getTrainingType().getTrainingTypeName();
        String trainingDateStart = training.getSessionDate().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);
        String trainingDateEnd = endTime.format(DateTimeFormatter.ISO_LOCAL_DATE_TIME);

        return new Event(trainingType, trainingDateStart, trainingDateEnd, training.getIdSession(), color);
    }

    public LocalDateTime getTrainingDateStart(int scheduleId) {
        TrainingSchedule training = trainingScheduleRepository.findById(scheduleId).orElse(null);
        return training != null ? training.getSessionDate() : null;
    }

    public LocalDateTime getTrainingDateEnd(int scheduleId) {
        TrainingSchedule training = trainingScheduleRepository.findById(scheduleId).orElse(null);
        return training != null ? training.getSessionDate().plusMinutes(training.getSessionTime()) : null;
    }

    public List<TrainingType> getTrainingTypes() {
        return trainingTypeRepository.findAll();
    }

    public Trainers getTrainer(int scheduleId) {
        TrainingSchedule training = trainingScheduleRepository.findById(scheduleId).orElse(null);
        return training != null ? training.getTrainer() : null;
    }

    @SuppressWarnings("null")
    public TrainingSchedule saveAndFlush(TrainingSchedule training) {
        return trainingScheduleRepository.saveAndFlush(training);
    }

    @SuppressWarnings("null")
    public TrainingSchedule save(TrainingSchedule training) {
        return trainingScheduleRepository.save(training);
    }

    @SuppressWarnings("null")
    public void deleteTrainingSchedule(Integer id) {
        trainingScheduleRepository.deleteById(id);
    }

    /**
     * Дополнительные полезные методы
     */
    public Set<TrainingSchedule> getCurrentUserTrainingSchedules() {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);

        if ("member".equals(role)) {
            Integer memberId = userDetailsService.getUserId(username);
            Members member = membersService.getMember(memberId);
            return new HashSet<>(member.getTrainingSchedules());
        } else if ("trainer".equals(role)) {
            Integer trainerId = userDetailsService.getUserId(username);
            return getTrainingSchedulesByTrainer(trainerId);
        }

        return Collections.emptySet();
    }

    public boolean isTrainingBelongsToCurrentUser(Integer trainingId) {
        TrainingSchedule training = getTrainingSchedule(trainingId);
        if (training == null)
            return false;

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);
        Integer userId = userDetailsService.getUserId(username);

        if ("member".equals(role)) {
            return training.getMembers().stream()
                    .anyMatch(member -> Objects.equals(member.getIdMember(), userId));
        } else if ("trainer".equals(role)) {
            return training.getTrainer().getIdTrainer() == userId;
        }

        return false;
    }

    public Set<TrainingSchedule> getTrainingsByMemberId(Integer memberId) {
        return trainingScheduleRepository.findByMembersIdMember(memberId);
    }
}