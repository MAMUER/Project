package com.example.project.service;

import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.dto.ProgramRequest;
import com.example.project.model.*;

import java.time.LocalDate;
import java.util.*;

@Slf4j
@Service
@AllArgsConstructor
public class AdaptiveProgramGenerator {
    private final TrainingProgramService trainingProgramService;
    private final MembersService membersService;
    private final ClubCapabilityService clubCapabilityService;

    @Transactional
    public TrainingProgram generateAdaptiveProgram(Integer memberId, ProgramRequest request) {
        try {
            Members member = membersService.getMember(memberId);
            if (member == null) {
                throw new IllegalArgumentException("Пользователь не найден");
            }

            // Определяем клуб из профиля пользователя
            String clubName = determineClubName(member);

            log.info("Адаптивная генерация программы для пользователя {} в клубе {}",
                    memberId, clubName);

            TrainingProgram program = createTrainingProgram(member, request, clubName);
            Set<ProgramDay> programDays = generateAdaptiveProgramDays(program, request, clubName);
            program.setProgramDays(programDays);

            trainingProgramService.deactivateOtherPrograms(memberId, null);
            TrainingProgram savedProgram = trainingProgramService.saveProgram(program);

            log.info("Адаптивная программа создана: ID={}, клуб={}",
                    savedProgram.getIdProgram(), clubName);

            return savedProgram;
        } catch (Exception e) {
            log.error("Ошибка при создании адаптивной программы для пользователя {}: {}",
                    memberId, e.getMessage(), e);
            throw new RuntimeException("Не удалось создать программу тренировок: " + e.getMessage(), e);
        }
    }

    // УДАЛИТЕ старый метод determineClubName и замените его на этот:
    private String determineClubName(Members member) {
        // Используем только клуб пользователя из профиля
        if (member.getClub() != null) {
            return member.getClub().getClubName();
        } else {
            throw new IllegalArgumentException("Пользователь не привязан к клубу. Обратитесь к администратору.");
        }
    }

    private TrainingProgram createTrainingProgram(Members member, ProgramRequest request, String clubName) {
        TrainingProgram program = new TrainingProgram();
        program.setMember(member);
        program.setProgramName(generateAdaptiveProgramName(request, clubName));
        program.setGoal(request.getGoal());
        program.setLevel(request.getLevel());
        program.setDurationWeeks(request.getDurationWeeks());
        program.setCreatedDate(LocalDate.now());
        program.setIsActive(true);
        return program;
    }

    private String generateAdaptiveProgramName(ProgramRequest request, String clubName) {
        String goalName = getGoalDisplayName(request.getGoal());
        String levelName = getLevelDisplayName(request.getLevel());
        return String.format("Программа %s (%s) - %s", goalName, levelName, clubName);
    }

    private Set<ProgramDay> generateAdaptiveProgramDays(TrainingProgram program, ProgramRequest request,
            String clubName) {
        int daysPerWeek = getDaysPerWeek(request.getLevel());
        Map<String, List<Exercise>> availableExercises = getAvailableExercisesByMuscleGroup(clubName);

        Set<ProgramDay> days = new LinkedHashSet<>();

        for (int i = 0; i < daysPerWeek; i++) {
            ProgramDay day = createAdaptiveProgramDay(program, i, request, availableExercises);
            days.add(day);
        }

        return days;
    }

    private Map<String, List<Exercise>> getAvailableExercisesByMuscleGroup(String clubName) {
        Map<String, List<Exercise>> availableExercises = new HashMap<>();

        // Получаем доступные упражнения для каждой группы мышц
        String[] muscleGroups = { "Грудные", "Спина", "Ноги", "Плечи", "Бицепсы", "Трицепсы", "Кардио" };

        for (String muscleGroup : muscleGroups) {
            List<Exercise> exercises = clubCapabilityService
                    .getAvailableExercisesByMuscleGroup(clubName, muscleGroup);
            availableExercises.put(muscleGroup, exercises);
        }

        return availableExercises;
    }

    private ProgramDay createAdaptiveProgramDay(TrainingProgram program, int dayIndex,
            ProgramRequest request,
            Map<String, List<Exercise>> availableExercises) {
        ProgramDay day = new ProgramDay();
        day.setTrainingProgram(program);
        day.setDayNumber(dayIndex + 1);
        day.setDayName(getDayName(dayIndex));

        // Определяем группы мышц для дня на основе цели и уровня
        String[] targetMuscleGroups = determineMuscleGroupsForDay(dayIndex, request.getGoal(), request.getLevel());
        day.setMuscleGroups(String.join(", ", targetMuscleGroups));

        Set<ProgramExercise> exercises = generateAdaptiveExercisesForDay(
                day, targetMuscleGroups, availableExercises, request.getLevel());
        day.setExercises(exercises);

        return day;
    }

    private String[] determineMuscleGroupsForDay(int dayIndex, String goal, String level) {
        // Адаптивная логика распределения групп мышц
        if ("похудение".equalsIgnoreCase(goal)) {
            return getWeightLossMuscleGroups(dayIndex, level);
        } else {
            return getMassGainMuscleGroups(dayIndex, level);
        }
    }

    private String[] getWeightLossMuscleGroups(int dayIndex, String level) {
        // Логика для похудения - больше кардио и фуллбади
        switch (dayIndex % 3) {
            case 0:
                return new String[] { "Грудные", "Трицепсы", "Кардио" };
            case 1:
                return new String[] { "Спина", "Бицепсы", "Кардио" };
            case 2:
                return new String[] { "Ноги", "Плечи", "Кардио" };
            default:
                return new String[] { "Кардио" };
        }
    }

    private String[] getMassGainMuscleGroups(int dayIndex, String level) {
        // Логика для набора массы - сплиты по группам мышц
        if ("начальный".equalsIgnoreCase(level)) {
            switch (dayIndex % 3) {
                case 0:
                    return new String[] { "Грудные", "Трицепсы" };
                case 1:
                    return new String[] { "Спина", "Бицепсы" };
                case 2:
                    return new String[] { "Ноги", "Плечи" };
            }
        } else {
            // Для среднего и продвинутого уровня - более специализированные сплиты
            switch (dayIndex % 4) {
                case 0:
                    return new String[] { "Грудные" };
                case 1:
                    return new String[] { "Спина" };
                case 2:
                    return new String[] { "Ноги" };
                case 3:
                    return new String[] { "Плечи", "Руки" };
            }
        }
        return new String[] { "Кардио" };
    }

    private Set<ProgramExercise> generateAdaptiveExercisesForDay(ProgramDay day,
            String[] targetMuscleGroups,
            Map<String, List<Exercise>> availableExercises,
            String level) {
        Set<ProgramExercise> exercises = new LinkedHashSet<>();
        int difficulty = getDifficultyLevel(level);
        int exerciseCounter = 1;

        for (String muscleGroup : targetMuscleGroups) {
            if ("Кардио".equals(muscleGroup)) {
                exercises.addAll(generateCardioExercises(day, availableExercises.get("Кардио")));
            } else {
                List<Exercise> muscleExercises = availableExercises.get(muscleGroup);
                if (muscleExercises != null && !muscleExercises.isEmpty()) {
                    // Выбираем 2-3 упражнения для группы мышц
                    int exercisesToSelect = Math.min(2 + difficulty, muscleExercises.size());
                    List<Exercise> selectedExercises = selectRandomExercises(muscleExercises, exercisesToSelect);

                    for (Exercise exercise : selectedExercises) {
                        ProgramExercise programExercise = createProgramExercise(day, exercise, difficulty,
                                exerciseCounter++);
                        exercises.add(programExercise);
                    }
                } else {
                    log.warn("Для группы мышц {} нет доступных упражнений", muscleGroup);
                }
            }
        }

        return exercises;
    }

    private List<Exercise> selectRandomExercises(List<Exercise> exercises, int count) {
        if (exercises.size() <= count) {
            return new ArrayList<>(exercises);
        }

        List<Exercise> shuffled = new ArrayList<>(exercises);
        Collections.shuffle(shuffled);
        return shuffled.subList(0, count);
    }

    private Set<ProgramExercise> generateCardioExercises(ProgramDay day, List<Exercise> cardioExercises) {
        Set<ProgramExercise> cardioProgramExercises = new HashSet<>();

        if (cardioExercises != null && !cardioExercises.isEmpty()) {
            Exercise cardioExercise = cardioExercises.get(new Random().nextInt(cardioExercises.size()));

            ProgramExercise programExercise = new ProgramExercise();
            programExercise.setProgramDay(day);
            programExercise.setExercise(cardioExercise);
            programExercise.setSets(1);
            programExercise.setReps(20 + new Random().nextInt(11)); // 20-30 минут
            programExercise.setRestSeconds(0);
            programExercise.setOrderIndex(99);
            cardioProgramExercises.add(programExercise);
        }

        return cardioProgramExercises;
    }

    // Вспомогательные методы (остаются без изменений)
    private ProgramExercise createProgramExercise(ProgramDay day, Exercise exercise, int difficulty, int order) {
        ProgramExercise programExercise = new ProgramExercise();
        programExercise.setProgramDay(day);
        programExercise.setExercise(exercise);
        programExercise.setOrderIndex(order);

        switch (difficulty) {
            case 1:
                programExercise.setSets(3);
                programExercise.setReps(10);
                programExercise.setRestSeconds(60);
                break;
            case 2:
                programExercise.setSets(4);
                programExercise.setReps(8);
                programExercise.setRestSeconds(45);
                break;
            case 3:
                programExercise.setSets(5);
                programExercise.setReps(6);
                programExercise.setRestSeconds(30);
                break;
        }

        return programExercise;
    }

    private int getDaysPerWeek(String level) {
        switch (level.toLowerCase()) {
            case "начальный":
                return 3;
            case "средний":
                return 4;
            case "продвинутый":
                return 5;
            default:
                return 3;
        }
    }

    private int getDifficultyLevel(String level) {
        switch (level.toLowerCase()) {
            case "начальный":
                return 1;
            case "средний":
                return 2;
            case "продвинутый":
                return 3;
            default:
                return 1;
        }
    }

    private String getDayName(int dayIndex) {
        String[] days = { "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье" };
        return days[dayIndex % days.length];
    }

    private String getGoalDisplayName(String goal) {
        switch (goal.toLowerCase()) {
            case "похудение":
                return "Похудения";
            case "набор_массы":
                return "Набора Массы";
            case "поддержание":
                return "Поддержания Формы";
            default:
                return goal;
        }
    }

    private String getLevelDisplayName(String level) {
        switch (level.toLowerCase()) {
            case "начальный":
                return "Начальный Уровень";
            case "средний":
                return "Средний Уровень";
            case "продвинутый":
                return "Продвинутый Уровень";
            default:
                return level;
        }
    }
}