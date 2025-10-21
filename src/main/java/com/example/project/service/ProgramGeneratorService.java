package com.example.project.service;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import com.example.project.dto.ProgramRequest;
import com.example.project.model.*;
import com.example.project.repository.ExerciseRepository;

import java.time.LocalDate;
import java.util.*;

@Service
@AllArgsConstructor
public class ProgramGeneratorService {
    private final ExerciseRepository exerciseRepository;
    private final TrainingProgramService trainingProgramService;
    private final MembersService membersService; // Это используется в generateProgram

    public TrainingProgram generateProgram(Integer memberId, ProgramRequest request) {
        // Используем membersService для получения данных пользователя
        Members member = membersService.getMember(memberId);

        // Создаем программу
        TrainingProgram program = new TrainingProgram();
        program.setMember(member);
        program.setProgramName("Программа: " + request.getGoal() + " (" + request.getLevel() + ")");
        program.setGoal(request.getGoal());
        program.setLevel(request.getLevel());
        program.setDurationWeeks(request.getDurationWeeks());
        program.setCreatedDate(LocalDate.now());
        program.setIsActive(true);

        // Генерируем дни тренировок
        int daysPerWeek = getDaysPerWeek(request.getLevel());
        Set<ProgramDay> programDays = generateProgramDays(program, daysPerWeek, request.getGoal(), request.getLevel());
        program.setProgramDays(programDays);

        // Используем trainingProgramService для сохранения
        trainingProgramService.deactivateOtherPrograms(memberId, null);
        return trainingProgramService.saveProgram(program);
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

    private Set<ProgramDay> generateProgramDays(TrainingProgram program, int daysPerWeek, String goal, String level) {
        Set<ProgramDay> days = new HashSet<>();
        String[][] dayTemplates = getDayTemplates(goal, level, daysPerWeek);

        for (int i = 0; i < daysPerWeek; i++) {
            ProgramDay day = new ProgramDay();
            day.setTrainingProgram(program);
            day.setDayNumber(i + 1);
            day.setDayName(getDayName(i));
            day.setMuscleGroups(dayTemplates[i][0]);

            Set<ProgramExercise> exercises = generateExercisesForDay(day, dayTemplates[i], level);
            day.setExercises(exercises);
            days.add(day);
        }
        return days;
    }

    private String[][] getDayTemplates(String goal, String level, int daysPerWeek) {
        if ("похудение".equalsIgnoreCase(goal)) {
            if (daysPerWeek == 3) {
                return new String[][] {
                        { "Грудные, Трицепсы", "Жим штанги лежа", "Жим гантелей лежа", "Отжимания на брусьях",
                                "Кардио" },
                        { "Ноги, Плечи", "Приседания со штангой", "Жим ногами", "Жим штанги стоя", "Кардио" },
                        { "Спина, Бицепсы", "Тяга штанги в наклоне", "Подтягивания", "Тяга вертикального блока",
                                "Кардио" }
                };
            } else {
                return new String[][] {
                        { "Грудные", "Жим штанги лежа", "Жим гантелей лежа", "Сведение рук в кроссовере", "Кардио" },
                        { "Ноги", "Приседания со штангой", "Жим ногами", "Выпады с гантелями", "Кардио" },
                        { "Спина", "Становая тяга", "Тяга штанги в наклоне", "Тяга вертикального блока", "Кардио" },
                        { "Плечи, Руки", "Жим штанги стоя", "Махи гантелями", "Подъем штанги на бицепс", "Кардио" }
                };
            }
        } else { // Набор массы или поддержание
            if (daysPerWeek == 3) {
                return new String[][] {
                        { "Грудные, Трицепсы", "Жим штанги лежа", "Жим гантелей лежа", "Отжимания на брусьях",
                                "Французский жим" },
                        { "Ноги, Плечи", "Приседания со штангой", "Жим ногами", "Жим штанги стоя", "Махи гантелями" },
                        { "Спина, Бицепсы", "Становая тяга", "Тяга штанги в наклоне", "Подтягивания",
                                "Подъем штанги на бицепс" }
                };
            } else {
                return new String[][] {
                        { "Грудные", "Жим штанги лежа", "Жим гантелей лежа", "Отжимания на брусьях",
                                "Сведение рук в кроссовере" },
                        { "Ноги", "Приседания со штангой", "Жим ногами", "Выпады с гантелями", "Подъем на носки" },
                        { "Спина", "Становая тяга", "Тяга штанги в наклоне", "Тяга вертикального блока", "Шраги" },
                        { "Плечи", "Жим штанги стоя", "Махи гантелями в стороны", "Тяга штанги к подбородку",
                                "Разведения в наклоне" },
                        { "Руки", "Подъем штанги на бицепс", "Молотковые сгибания", "Французский жим",
                                "Разгибания на блоке" }
                };
            }
        }
    }

    private Set<ProgramExercise> generateExercisesForDay(ProgramDay day, String[] dayTemplate, String level) {
        Set<ProgramExercise> exercises = new HashSet<>();
        int difficulty = getDifficultyLevel(level);

        // Пропускаем первый элемент - это группа мышц
        for (int i = 1; i < dayTemplate.length; i++) {
            String exerciseName = dayTemplate[i];

            if ("Кардио".equals(exerciseName)) {
                // Добавляем кардио упражнения
                exercises.addAll(generateCardioExercises(day, 20, 30)); // 20-30 минут кардио
            } else {
                // Ищем упражнение в базе
                Exercise exercise = findExerciseByName(exerciseName);
                if (exercise != null) {
                    ProgramExercise programExercise = createProgramExercise(day, exercise, difficulty, i);
                    exercises.add(programExercise);
                }
            }
        }
        return exercises;
    }

    private ProgramExercise createProgramExercise(ProgramDay day, Exercise exercise, int difficulty, int order) {
        ProgramExercise programExercise = new ProgramExercise();
        programExercise.setProgramDay(day);
        programExercise.setExercise(exercise);
        programExercise.setOrderIndex(order);

        // Настройка параметров в зависимости от сложности
        switch (difficulty) {
            case 1: // Начальный
                programExercise.setSets(3);
                programExercise.setReps(10);
                programExercise.setRestSeconds(60);
                break;
            case 2: // Средний
                programExercise.setSets(4);
                programExercise.setReps(8);
                programExercise.setRestSeconds(45);
                break;
            case 3: // Продвинутый
                programExercise.setSets(5);
                programExercise.setReps(6);
                programExercise.setRestSeconds(30);
                break;
        }

        return programExercise;
    }

    private Set<ProgramExercise> generateCardioExercises(ProgramDay day, int minTime, int maxTime) {
        Set<ProgramExercise> cardioExercises = new HashSet<>();
        String[] cardioTypes = { "Бег на дорожке", "Велотренажер", "Эллиптический тренажер" };

        Random random = new Random();
        String cardioType = cardioTypes[random.nextInt(cardioTypes.length)];
        Exercise cardioExercise = findExerciseByName(cardioType);

        if (cardioExercise != null) {
            ProgramExercise programExercise = new ProgramExercise();
            programExercise.setProgramDay(day);
            programExercise.setExercise(cardioExercise);
            programExercise.setSets(1);
            programExercise.setReps(minTime + random.nextInt(maxTime - minTime)); // Время в минутах
            programExercise.setRestSeconds(0);
            programExercise.setOrderIndex(99); // Кардио в конце
            cardioExercises.add(programExercise);
        }

        return cardioExercises;
    }

    private Exercise findExerciseByName(String name) {
        // Используем репозиторий для поиска по имени
        return exerciseRepository.findAll().stream()
                .filter(e -> e.getExerciseName().equalsIgnoreCase(name))
                .findFirst()
                .orElse(null);
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
}