package com.example.project.service;

import java.time.LocalDate;
import java.time.Period;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.Set;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.dto.ProgramRequest;
import com.example.project.model.Exercise;
import com.example.project.model.Members;
import com.example.project.model.NutritionPlan;
import com.example.project.model.ProgramDay;
import com.example.project.model.ProgramExercise;
import com.example.project.model.TrainingProgram;

import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;

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
        } catch (IllegalArgumentException e) {
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
        if (age >= 18 && age <= 29) {
            return "18-29";
        } else if (age >= 30 && age <= 39) {
            return "30-39";
        } else if (age >= 40 && age <= 49) {
            return "40-49";
        } else if (age >= 50 && age <= 59) {
            return "50-59";
        } else {
            return "60+";
        }
    }

    private TrainingProgram createTrainingProgram(Members member, ProgramRequest request, String clubName,
            String ageGroup) {
        TrainingProgram program = new TrainingProgram();
        program.setMember(member);
        program.setProgramName(generateAdaptiveProgramName(request, clubName, ageGroup));
        program.setGoal(request.getGoal());
        program.setLevel(request.getLevel());
        program.setDurationWeeks(request.getDurationWeeks());
        program.setCreatedDate(LocalDate.now());
        program.setIsActive(true);
        return program;
    }

    private String generateAdaptiveProgramName(ProgramRequest request, String clubName, String ageGroup) {
        String goalName = getGoalDisplayName(request.getGoal());
        String levelName = getLevelDisplayName(request.getLevel());
        String ageRange = ageGroup;
        return String.format("Программа %s (%s) - %s [Возраст: %s]", goalName, levelName, clubName, ageRange);
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
        day.setDayName(dayName);

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
        baseDays = switch (level.toLowerCase()) {
            case "начальный" ->
                4;
            case "средний" ->
                5;
            case "продвинутый" ->
                6;
            default ->
                4;
        };

        // Корректировка по возрасту
        return switch (ageGroup) {
            case "50-59" ->
                Math.min(baseDays, 4);
            case "60+" ->
                Math.min(baseDays, 3);
            default ->
                baseDays;
        };
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
        // Определяем базовую структуру в зависимости от уровня
        String[][] levelBasedStructure = getLevelBasedStructure(level);

        // Применяем возрастные корректировки
        if ("50-59".equals(ageGroup) || "60+".equals(ageGroup)) {
            return applyAgeAdjustments(dayIndex, levelBasedStructure, ageGroup);
        } else {
            // Для младших возрастных групп используем структуру по уровню без изменений
            return levelBasedStructure[dayIndex % levelBasedStructure.length];
        }
    }

    private String[][] getLevelBasedStructure(String level) {
        return switch (level.toLowerCase()) {
            case "начальный" ->
                new String[][]{
                    {"Грудные", "Кардио"},
                    {"Спина", "Кардио"},
                    {"Ноги", "Кардио"}
                };
            case "средний" ->
                new String[][]{
                    {"Грудные", "Трицепсы", "Кардио"},
                    {"Спина", "Бицепсы", "Кардио"},
                    {"Ноги", "Плечи", "Кардио"}
                };
            case "продвинутый" ->
                new String[][]{
                    {"Грудные", "Трицепсы", "Плечи", "Кардио"},
                    {"Спина", "Бицепсы", "Кардио"},
                    {"Ноги", "Кардио"},
                    {"Плечи", "Руки", "Кардио"}
                };
            default ->
                new String[][]{
                    {"Грудные", "Трицепсы", "Кардио"},
                    {"Спина", "Бицепсы", "Кардио"},
                    {"Ноги", "Плечи", "Кардио"}
                };
        };
    }

    private String[] applyAgeAdjustments(int dayIndex, String[][] levelBasedStructure, String ageGroup) {
        String[] originalDay = levelBasedStructure[dayIndex % levelBasedStructure.length];

        // Для возрастных групп уменьшаем количество силовых групп
        if ("50-59".equals(ageGroup)) {
            // Убираем одну силовую группу, оставляем основные
            return filterMuscleGroups(originalDay, 1);
        } else if ("60+".equals(ageGroup)) {
            // Оставляем только одну силовую группу + кардио
            return filterMuscleGroups(originalDay, 1, true);
        }

        return originalDay;
    }

    private String[] filterMuscleGroups(String[] original, int maxStrengthGroups) {
        return filterMuscleGroups(original, maxStrengthGroups, false);
    }

    private String[] filterMuscleGroups(String[] original, int maxStrengthGroups, boolean prioritizeCardio) {
        List<String> strengthGroups = new ArrayList<>();
        boolean hasCardio = false;

        for (String group : original) {
            if ("Кардио".equals(group)) {
                hasCardio = true;
            } else {
                strengthGroups.add(group);
            }
        }

        // Оставляем только приоритетные группы мышц
        List<String> result = new ArrayList<>();

        // Для 60+ оставляем только ноги и спину как наиболее важные
        if (prioritizeCardio) {
            if (strengthGroups.contains("Ноги")) {
                result.add("Ноги");
            } else if (!strengthGroups.isEmpty()) {
                result.add(strengthGroups.get(0));
            }
        } else {
            // Добавляем не больше maxStrengthGroups силовых групп
            result.addAll(strengthGroups.stream()
                    .limit(maxStrengthGroups)
                    .collect(Collectors.toList()));
        }

        // Всегда добавляем кардио для возрастных групп
        if (hasCardio || prioritizeCardio) {
            result.add("Кардио");
        }

        return result.toArray(String[]::new);
    }

    private String[] getMassGainMuscleGroups(int dayIndex, String level, String ageGroup) {
        // Для возрастных групп 50+ применяем более щадящий подход даже при наборе массы
        if ("50-59".equals(ageGroup) || "60+".equals(ageGroup)) {
            return getSeniorMassGainMuscleGroups(dayIndex, level, ageGroup);
        }

        // Для остальных возрастных групп используем стандартную логику
        return getStandardMassGainMuscleGroups(dayIndex, level);
    }

    private String[] getStandardMassGainMuscleGroups(int dayIndex, String level) {
        if ("начальный".equalsIgnoreCase(level)) {
            switch (dayIndex % 3) {
                case 0 -> {
                    return new String[]{"Грудные", "Трицепсы"};
                }
                case 1 -> {
                    return new String[]{"Спина", "Бицепсы"};
                }
                case 2 -> {
                    return new String[]{"Ноги", "Плечи"};
                }
            }
        } else {
            switch (dayIndex % 4) {
                case 0 -> {
                    return new String[]{"Грудные"};
                }
                case 1 -> {
                    return new String[]{"Спина"};
                }
                case 2 -> {
                    return new String[]{"Ноги"};
                }
                case 3 -> {
                    return new String[]{"Плечи", "Руки"};
                }
            }
        }
        return new String[]{"Кардио"};
    }

    private String[] getSeniorMassGainMuscleGroups(int dayIndex, String level, String ageGroup) {
        // Для пожилых людей фокус на функциональные группы мышц
        if ("60+".equals(ageGroup)) {
            // Для 60+ - минимальная нагрузка, фокус на ноги и спину
            // Учитываем уровень подготовки
            int cycleLength = "начальный".equalsIgnoreCase(level) ? 2 : 3;

            return switch (dayIndex % cycleLength) {
                case 0 ->
                    new String[]{"Ноги", "Кардио"};
                case 1 ->
                    new String[]{"Спина", "Кардио"};
                case 2 ->
                    new String[]{"Плечи", "Кардио"}; // Только для среднего/продвинутого уровня
                default ->
                    new String[]{"Кардио"};
            };
        } else { // 50-59
            // Для 50-59 - более сбалансированный подход с акцентом на безопасность
            // Учитываем уровень подготовки
            if ("начальный".equalsIgnoreCase(level)) {
                return switch (dayIndex % 2) {
                    case 0 ->
                        new String[]{"Грудные", "Кардио"};
                    case 1 ->
                        new String[]{"Спина", "Ноги", "Кардио"};
                    default ->
                        new String[]{"Кардио"};
                };
            } else {
                return switch (dayIndex % 3) {
                    case 0 ->
                        new String[]{"Грудные", "Трицепсы", "Кардио"};
                    case 1 ->
                        new String[]{"Спина", "Бицепсы", "Кардио"};
                    case 2 ->
                        new String[]{"Ноги", "Плечи", "Кардио"};
                    default ->
                        new String[]{"Кардио"};
                };
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

    // Количество упражнений в зависимости от возраста и группы мышц
    private int getExercisesCountForAge(String muscleGroup, String ageGroup) {
        // Определяем важность группы мышц для конкретного возраста
        Map<String, Integer> exerciseCounts = new HashMap<>();

        switch (ageGroup) {
            case "18-29", "30-39" -> {
                exerciseCounts.put("Грудные", 3);
                exerciseCounts.put("Спина", 3);
                exerciseCounts.put("Ноги", 4);  // Ногам больше внимания
                exerciseCounts.put("Плечи", 3);
                exerciseCounts.put("Бицепсы", 2);
                exerciseCounts.put("Трицепсы", 2);
            }
            case "40-49" -> {
                exerciseCounts.put("Грудные", 2);
                exerciseCounts.put("Спина", 3);  // Спине больше внимания (профилактика)
                exerciseCounts.put("Ноги", 3);
                exerciseCounts.put("Плечи", 2);
                exerciseCounts.put("Бицепсы", 1);
                exerciseCounts.put("Трицепсы", 1);
            }
            case "50-59" -> {
                exerciseCounts.put("Грудные", 2);
                exerciseCounts.put("Спина", 3);  // Спина важна для осанки
                exerciseCounts.put("Ноги", 3);   // Ноги важны для мобильности
                exerciseCounts.put("Плечи", 1);
                exerciseCounts.put("Бицепсы", 1);
                exerciseCounts.put("Трицепсы", 1);
            }
            case "60+" -> {
                exerciseCounts.put("Грудные", 1);
                exerciseCounts.put("Спина", 2);  // Спина в приоритете
                exerciseCounts.put("Ноги", 2);   // Ноги в приоритете
                exerciseCounts.put("Плечи", 1);
                exerciseCounts.put("Бицепсы", 1);
                exerciseCounts.put("Трицепсы", 1);
            }
            default -> {
                return 2;
            }
        }

        return exerciseCounts.getOrDefault(muscleGroup, 2);
    }

    private int getDifficultyLevel(String level, String ageGroup) {
        int baseDifficulty;
        baseDifficulty = switch (level.toLowerCase()) {
            case "начальный" ->
                1;
            case "средний" ->
                2;
            case "продвинутый" ->
                3;
            default ->
                1;
        };

        // Корректировка сложности по возрасту
        return switch (ageGroup) {
            case "40-49" ->
                Math.min(baseDifficulty, 2);
            case "50-59" ->
                Math.min(baseDifficulty, 2);
            case "60+" ->
                1;
            default ->
                baseDifficulty;
        };
    }

    private ProgramExercise createProgramExercise(ProgramDay day, Exercise exercise, int difficulty, int order,
            String ageGroup) {
        ProgramExercise programExercise = new ProgramExercise();
        programExercise.setProgramDay(day);
        programExercise.setExercise(exercise);
        programExercise.setOrderIndex(order);

        // Настройки в зависимости от возраста и сложности
        switch (ageGroup) {
            case "18-29" ->
                setExerciseParametersYoung(programExercise, difficulty);
            case "30-39" ->
                setExerciseParametersYoung(programExercise, difficulty);
            case "40-49" ->
                setExerciseParametersMiddleAge(programExercise, difficulty);
            case "50-59" ->
                setExerciseParametersSenior(programExercise, difficulty);
            case "60+" ->
                setExerciseParametersElderly(programExercise);
            default ->
                setExerciseParametersYoung(programExercise, difficulty);
        }

        return programExercise;
    }

    private void setExerciseParametersYoung(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1 -> {
                exercise.setSets(3);
                exercise.setReps(12);
                exercise.setRestSeconds(60);
            }
            case 2 -> {
                exercise.setSets(4);
                exercise.setReps(10);
                exercise.setRestSeconds(45);
            }
            case 3 -> {
                exercise.setSets(5);
                exercise.setReps(8);
                exercise.setRestSeconds(30);
            }
        }
    }

    private void setExerciseParametersMiddleAge(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1 -> {
                exercise.setSets(3);
                exercise.setReps(10);
                exercise.setRestSeconds(75);
            }
            case 2 -> {
                exercise.setSets(3);
                exercise.setReps(8);
                exercise.setRestSeconds(60);
            }
            case 3 -> {
                exercise.setSets(4);
                exercise.setReps(6);
                exercise.setRestSeconds(45);
            }
        }
    }

    private void setExerciseParametersSenior(ProgramExercise exercise, int difficulty) {
        switch (difficulty) {
            case 1 -> {
                exercise.setSets(2);
                exercise.setReps(12);
                exercise.setRestSeconds(90);
            }
            case 2 -> {
                exercise.setSets(3);
                exercise.setReps(10);
                exercise.setRestSeconds(75);
            }
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
        return switch (ageGroup) {
            case "18-29" ->
                30;
            case "30-39" ->
                25;
            case "40-49" ->
                20;
            case "50-59" ->
                15;
            case "60+" ->
                10;
            default ->
                20;
        };
    }

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
        // Определяем сложность диеты на основе уровня подготовки и возраста
        String difficulty = getNutritionDifficulty(level, ageGroup);

        // Определяем пол для более точного подбора
        Integer gender = member.getGender(); // 1 - мужской, 0 - женский

        // Ищем подходящие диеты по цели и сложности
        List<NutritionPlan> suitablePlans = nutritionPlanService.findByGoalAndDifficulty(goal, difficulty);

        if (!suitablePlans.isEmpty()) {
            // Если есть несколько планов, можно выбрать более подходящий по полу
            if (suitablePlans.size() > 1) {
                // Попробуем найти план, учитывающий пол (например, в описании)
                for (NutritionPlan plan : suitablePlans) {
                    String description = plan.getNutritionDescription();
                    if (description != null) {
                        if (gender == 1 && description.toLowerCase().contains("мужск")) {
                            return plan;
                        } else if (gender == 0 && description.toLowerCase().contains("женск")) {
                            return plan;
                        }
                    }
                }
            }
            return suitablePlans.get(0);
        }

        // Если не нашли по сложности, ищем только по цели
        List<NutritionPlan> goalPlans = nutritionPlanService.findByGoal(goal);
        if (!goalPlans.isEmpty()) {
            return goalPlans.get(0);
        }

        return null;
    }

    private String getNutritionDifficulty(String level, String ageGroup) {
        // Базовая сложность от уровня подготовки
        String baseDifficulty = switch (level.toLowerCase()) {
            case "начальный" ->
                "легкий";
            case "средний" ->
                "средний";
            case "продвинутый" ->
                "сложный";
            default ->
                "легкий";
        };

        // Корректировка для возрастных групп
        if ("60+".equals(ageGroup)) {
            return "легкий"; // Для пожилых всегда легкая диета
        } else if ("50-59".equals(ageGroup) && "сложный".equals(baseDifficulty)) {
            return "средний"; // Понижаем сложность для 50-59
        }

        return baseDifficulty;
    }

    private String determineClubName(Members member) {
        if (member.getClub() != null) {
            return member.getClub().getClubName();
        } else {
            throw new IllegalArgumentException("Пользователь не привязан к клубу. Обратитесь к администратору.");
        }
    }

    private Map<String, List<Exercise>> getAvailableExercisesByMuscleGroup(String clubName) {
        Map<String, List<Exercise>> availableExercises = new HashMap<>();
        String[] muscleGroups = {"Грудные", "Спина", "Ноги", "Плечи", "Бицепсы", "Трицепсы", "Кардио"};
        for (String muscleGroup : muscleGroups) {
            List<Exercise> exercises = clubCapabilityService
                    .getAvailableExercisesByMuscleGroup(clubName, muscleGroup);
            availableExercises.put(muscleGroup, exercises);
        }
        return availableExercises;
    }

    private List<Exercise> selectRandomExercises(List<Exercise> exercises, int count) {
        if (exercises.size() <= count) {
            return new ArrayList<>(exercises);
        }
        List<Exercise> shuffled = new ArrayList<>(exercises);
        Collections.shuffle(shuffled);
        return shuffled.subList(0, count);
    }

    private String getGoalDisplayName(String goal) {
        return switch (goal.toLowerCase()) {
            case "похудение" ->
                "Похудение";
            case "набор мышц" ->
                "Набор мышц";
            case "поддержание" ->
                "Поддержание формы";
            default ->
                goal;
        };
    }

    private String getLevelDisplayName(String level) {
        return switch (level.toLowerCase()) {
            case "начальный" ->
                "Начальный Уровень";
            case "средний" ->
                "Средний Уровень";
            case "продвинутый" ->
                "Продвинутый Уровень";
            default ->
                level;
        };
    }
}
