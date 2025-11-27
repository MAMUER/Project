package com.example.project.service;

import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.dto.ProgramRequest;
import com.example.project.model.*;

import java.time.LocalDate;
import java.time.Period;
import java.util.*;

@Slf4j
@Service
@AllArgsConstructor
public class AdaptiveProgramGenerator {
    private final TrainingProgramService trainingProgramService;
    private final MembersService membersService;
    private final ClubCapabilityService clubCapabilityService;
    private final NutritionPlanService nutritionPlanService;

    @Transactional
    public TrainingProgram generateAdaptiveProgram(Integer memberId, ProgramRequest request) {
        try {
            Members member = membersService.getMember(memberId);
            if (member == null) {
                throw new IllegalArgumentException("Пользователь не найден");
            }

            // Определяем возраст пользователя
            int age = calculateAge(member.getBirthDate());
            String ageGroup = determineAgeGroup(age);

            // Определяем клуб из профиля пользователя
            String clubName = determineClubName(member);

            log.info("Адаптивная генерация программы для пользователя {} (возраст: {}, группа: {}) в клубе {}",
                    memberId, age, ageGroup, clubName);

            // УЧИТЫВАЕМ РАСПИСАНИЕ при создании программы
            TrainingProgram program = createTrainingProgram(member, request, clubName, ageGroup);

            // Генерируем дни с учетом выбранного расписания
            Set<ProgramDay> programDays = generateAdaptiveProgramDays(
                    program, request, clubName, ageGroup, request.getTrainingDays());

            program.setProgramDays(programDays);

            // Создаем план питания
            assignNutritionPlan(program, request, ageGroup, member);

            trainingProgramService.deactivateOtherPrograms(memberId, null);
            TrainingProgram savedProgram = trainingProgramService.saveProgram(program);

            log.info("Адаптивная программа создана: ID={}, возрастная группа={}, клуб={}",
                    savedProgram.getIdProgram(), ageGroup, clubName);

            return savedProgram;
        } catch (Exception e) {
            log.error("Ошибка при создании адаптивной программы для пользователя {}: {}",
                    memberId, e.getMessage(), e);
            throw new RuntimeException("Не удалось создать программу тренировок: " + e.getMessage(), e);
        }
    }

    // Метод для расчета возраста
    private int calculateAge(LocalDate birthDate) {
        return Period.between(birthDate, LocalDate.now()).getYears();
    }

    // Метод для определения возрастной группы
    private String determineAgeGroup(int age) {
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

    private TrainingProgram createTrainingProgram(Members member, ProgramRequest request, String clubName,
            String ageGroup) {
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
            String clubName, String ageGroup, List<String> selectedDays) {

        // Используем выбранные пользователем дни вместо фиксированного количества
        int daysPerWeek = Math.min(selectedDays.size(),
                getMaxDaysPerWeek(request.getLevel(), ageGroup));

        Map<String, List<Exercise>> availableExercises = getAvailableExercisesByMuscleGroup(clubName);

        Set<ProgramDay> days = new LinkedHashSet<>();

        for (int i = 0; i < daysPerWeek; i++) {
            String dayName = selectedDays.get(i % selectedDays.size());
            ProgramDay day = createAdaptiveProgramDay(program, i, request, availableExercises, ageGroup, dayName);
            days.add(day);
        }

        return days;
    }

    private ProgramDay createAdaptiveProgramDay(TrainingProgram program, int dayIndex,
            ProgramRequest request,
            Map<String, List<Exercise>> availableExercises, String ageGroup, String dayName) {

        ProgramDay day = new ProgramDay();
        day.setTrainingProgram(program);
        day.setDayNumber(dayIndex + 1);
        day.setDayName(dayName); // Используем выбранный пользователем день

        // Определяем группы мышц для дня на основе цели, уровня и возраста
        String[] targetMuscleGroups = determineMuscleGroupsForDay(dayIndex, request.getGoal(), request.getLevel(),
                ageGroup);
        day.setMuscleGroups(String.join(", ", targetMuscleGroups));

        Set<ProgramExercise> exercises = generateAdaptiveExercisesForDay(
                day, targetMuscleGroups, availableExercises, request.getLevel(), ageGroup);
        day.setExercises(exercises);

        return day;
    }

    // Максимальное количество дней в зависимости от уровня и возраста
    private int getMaxDaysPerWeek(String level, String ageGroup) {
        int baseDays;
        switch (level.toLowerCase()) {
            case "начальный":
                baseDays = 4;
                break;
            case "средний":
                baseDays = 5;
                break;
            case "продвинутый":
                baseDays = 6;
                break;
            default:
                baseDays = 4;
        }

        // Корректировка по возрасту
        switch (ageGroup) {
            case "50-59":
                return Math.min(baseDays, 4);
            case "60+":
                return Math.min(baseDays, 3);
            default:
                return baseDays;
        }
    }

    private String[] determineMuscleGroupsForDay(int dayIndex, String goal, String level, String ageGroup) {
        // Адаптивная логика распределения групп мышц с учетом возраста
        if ("похудение".equalsIgnoreCase(goal)) {
            return getWeightLossMuscleGroups(dayIndex, level, ageGroup);
        } else {
            return getMassGainMuscleGroups(dayIndex, level, ageGroup);
        }
    }

    private String[] getWeightLossMuscleGroups(int dayIndex, String level, String ageGroup) {
        // Для старших возрастных групп - больше кардио и меньше силовых
        if ("50-59".equals(ageGroup) || "60+".equals(ageGroup)) {
            switch (dayIndex % 3) {
                case 0:
                    return new String[] { "Ноги", "Кардио" };
                case 1:
                    return new String[] { "Спина", "Кардио" };
                case 2:
                    return new String[] { "Кардио" };
                default:
                    return new String[] { "Кардио" };
            }
        } else {
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
    }

    private Set<ProgramExercise> generateAdaptiveExercisesForDay(ProgramDay day,
            String[] targetMuscleGroups,
            Map<String, List<Exercise>> availableExercises,
            String level, String ageGroup) {
        Set<ProgramExercise> exercises = new LinkedHashSet<>();
        int difficulty = getDifficultyLevel(level, ageGroup);
        int exerciseCounter = 1;

        for (String muscleGroup : targetMuscleGroups) {
            if ("Кардио".equals(muscleGroup)) {
                exercises.addAll(generateCardioExercises(day, availableExercises.get("Кардио"), ageGroup));
            } else {
                List<Exercise> muscleExercises = availableExercises.get(muscleGroup);
                if (muscleExercises != null && !muscleExercises.isEmpty()) {
                    // Выбираем упражнения с учетом возраста
                    int exercisesToSelect = getExercisesCountForAge(muscleGroup, ageGroup);
                    exercisesToSelect = Math.min(exercisesToSelect, muscleExercises.size());
                    List<Exercise> selectedExercises = selectRandomExercises(muscleExercises, exercisesToSelect);

                    for (Exercise exercise : selectedExercises) {
                        ProgramExercise programExercise = createProgramExercise(day, exercise, difficulty,
                                exerciseCounter++, ageGroup);
                        exercises.add(programExercise);
                    }
                } else {
                    log.warn("Для группы мышц {} нет доступных упражнений", muscleGroup);
                }
            }
        }

        return exercises;
    }

    // Количество упражнений в зависимости от возраста
    private int getExercisesCountForAge(String muscleGroup, String ageGroup) {
        switch (ageGroup) {
            case "18-29":
                return 3;
            case "30-39":
                return 3;
            case "40-49":
                return 2;
            case "50-59":
                return 2;
            case "60+":
                return 1;
            default:
                return 2;
        }
    }

    private int getDifficultyLevel(String level, String ageGroup) {
        int baseDifficulty;
        switch (level.toLowerCase()) {
            case "начальный":
                baseDifficulty = 1;
                break;
            case "средний":
                baseDifficulty = 2;
                break;
            case "продвинутый":
                baseDifficulty = 3;
                break;
            default:
                baseDifficulty = 1;
        }

        // Корректировка сложности по возрасту
        switch (ageGroup) {
            case "40-49":
                return Math.min(baseDifficulty, 2);
            case "50-59":
                return Math.min(baseDifficulty, 2);
            case "60+":
                return 1;
            default:
                return baseDifficulty;
        }
    }

    private ProgramExercise createProgramExercise(ProgramDay day, Exercise exercise, int difficulty, int order,
            String ageGroup) {
        ProgramExercise programExercise = new ProgramExercise();
        programExercise.setProgramDay(day);
        programExercise.setExercise(exercise);
        programExercise.setOrderIndex(order);

        // Настройки в зависимости от возраста и сложности
        switch (ageGroup) {
            case "18-29":
                setExerciseParametersYoung(programExercise, difficulty);
                break;
            case "30-39":
                setExerciseParametersYoung(programExercise, difficulty);
                break;
            case "40-49":
                setExerciseParametersMiddleAge(programExercise, difficulty);
                break;
            case "50-59":
                setExerciseParametersSenior(programExercise, difficulty);
                break;
            case "60+":
                setExerciseParametersElderly(programExercise);
                break;
            default:
                setExerciseParametersYoung(programExercise, difficulty);
        }

        return programExercise;
    }

    private void setExerciseParametersYoung(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1:
                exercise.setSets(3);
                exercise.setReps(12);
                exercise.setRestSeconds(60);
                break;
            case 2:
                exercise.setSets(4);
                exercise.setReps(10);
                exercise.setRestSeconds(45);
                break;
            case 3:
                exercise.setSets(5);
                exercise.setReps(8);
                exercise.setRestSeconds(30);
                break;
        }
    }

    private void setExerciseParametersMiddleAge(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1:
                exercise.setSets(3);
                exercise.setReps(10);
                exercise.setRestSeconds(75);
                break;
            case 2:
                exercise.setSets(3);
                exercise.setReps(8);
                exercise.setRestSeconds(60);
                break;
            case 3:
                exercise.setSets(4);
                exercise.setReps(6);
                exercise.setRestSeconds(45);
                break;
        }
    }

    private void setExerciseParametersSenior(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1:
                exercise.setSets(2);
                exercise.setReps(12);
                exercise.setRestSeconds(90);
                break;
            case 2:
                exercise.setSets(3);
                exercise.setReps(10);
                exercise.setRestSeconds(75);
                break;
        }
    }

    private void setExerciseParametersElderly(ProgramExercise exercise) {
        exercise.setSets(2);
        exercise.setReps(10);
        exercise.setRestSeconds(90);
    }

    private Set<ProgramExercise> generateCardioExercises(ProgramDay day, List<Exercise> cardioExercises,
            String ageGroup) {
        Set<ProgramExercise> cardioProgramExercises = new HashSet<>();

        if (cardioExercises != null && !cardioExercises.isEmpty()) {
            Exercise cardioExercise = cardioExercises.get(new Random().nextInt(cardioExercises.size()));

            ProgramExercise programExercise = new ProgramExercise();
            programExercise.setProgramDay(day);
            programExercise.setExercise(cardioExercise);
            programExercise.setSets(1);

            // Время кардио в зависимости от возраста
            int cardioTime = getCardioTimeForAge(ageGroup);
            programExercise.setReps(cardioTime);
            programExercise.setRestSeconds(0);
            programExercise.setOrderIndex(99);
            cardioProgramExercises.add(programExercise);
        }

        return cardioProgramExercises;
    }

    private int getCardioTimeForAge(String ageGroup) {
        switch (ageGroup) {
            case "18-29":
                return 30;
            case "30-39":
                return 25;
            case "40-49":
                return 20;
            case "50-59":
                return 15;
            case "60+":
                return 10;
            default:
                return 20;
        }
    }

    // ДОБАВИТЬ этот новый метод:
    private void assignNutritionPlan(TrainingProgram program, ProgramRequest request, String ageGroup, Members member) {
        try {
            // Получаем подходящую диету на основе цели, уровня и возраста
            NutritionPlan nutritionPlan = findAppropriateNutritionPlan(
                    request.getGoal(),
                    request.getLevel(),
                    ageGroup,
                    member);

            if (nutritionPlan != null) {
                program.setNutritionPlan(nutritionPlan);
            } else {
                log.warn("Не удалось найти подходящий план питания для цели: {}", request.getGoal());
            }

        } catch (Exception e) {
            log.warn("Не удалось назначить план питания для программы {}: {}", program.getProgramName(),
                    e.getMessage());
        }
    }

    private NutritionPlan findAppropriateNutritionPlan(String goal, String level, String ageGroup, Members member) {
        // Определяем сложность диеты на основе уровня подготовки
        String difficulty = getNutritionDifficulty(level, ageGroup);

        // Ищем подходящие диеты по цели и сложности
        List<NutritionPlan> suitablePlans = nutritionPlanService.findByGoalAndDifficulty(goal, difficulty);

        if (!suitablePlans.isEmpty()) {
            return suitablePlans.get(0); // Возвращаем первую подходящую
        }

        // Если не нашли по сложности, ищем только по цели
        List<NutritionPlan> goalPlans = nutritionPlanService.findByGoal(goal);
        if (!goalPlans.isEmpty()) {
            return goalPlans.get(0);
        }

        return null;
    }

    private String getNutritionDifficulty(String level, String ageGroup) {
        switch (level.toLowerCase()) {
            case "начальный":
                return "легкий";
            case "средний":
                return "средний";
            case "продвинутый":
                return "сложный";
            default:
                return "легкий";
        }
    }

    // Остальные методы остаются без изменений
    private String determineClubName(Members member) {
        if (member.getClub() != null) {
            return member.getClub().getClubName();
        } else {
            throw new IllegalArgumentException("Пользователь не привязан к клубу. Обратитесь к администратору.");
        }
    }

    private Map<String, List<Exercise>> getAvailableExercisesByMuscleGroup(String clubName) {
        Map<String, List<Exercise>> availableExercises = new HashMap<>();
        String[] muscleGroups = { "Грудные", "Спина", "Ноги", "Плечи", "Бицепсы", "Трицепсы", "Кардио" };
        for (String muscleGroup : muscleGroups) {
            List<Exercise> exercises = clubCapabilityService
                    .getAvailableExercisesByMuscleGroup(clubName, muscleGroup);
            availableExercises.put(muscleGroup, exercises);
        }
        return availableExercises;
    }

    private String[] getMassGainMuscleGroups(int dayIndex, String level, String ageGroup) {
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

    private List<Exercise> selectRandomExercises(List<Exercise> exercises, int count) {
        if (exercises.size() <= count)
            return new ArrayList<>(exercises);
        List<Exercise> shuffled = new ArrayList<>(exercises);
        Collections.shuffle(shuffled);
        return shuffled.subList(0, count);
    }

    private String getGoalDisplayName(String goal) {
        switch (goal.toLowerCase()) {
            case "похудение":
                return "Похудение";
            case "набор мышц":
                return "Набор мышц";
            case "поддержание":
                return "Поддержание формы";
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