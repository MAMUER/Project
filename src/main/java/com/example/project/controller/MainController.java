package com.example.project.controller;

import java.security.Principal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;

import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.ui.Model;
import org.springframework.validation.BindingResult;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

import com.example.project.client.StatsServiceClient;
import com.example.project.dto.NewsDTO;
import com.example.project.dto.ProgramRequest;
import com.example.project.model.Achievements;
import com.example.project.model.Clubs;
import com.example.project.model.Members;
import com.example.project.model.ProgramDay;
import com.example.project.model.ProgramExercise;
import com.example.project.model.Staff;
import com.example.project.model.StaffSchedule;
import com.example.project.model.Trainers;
import com.example.project.model.TrainingProgram;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.TrainingType;
import com.example.project.model.accounts.MembersAccounts;
import com.example.project.model.accounts.StaffAccounts;
import com.example.project.model.accounts.TrainersAccounts;
import com.example.project.repository.MembersAccountsRepository;
import com.example.project.repository.StaffAccountsRepository;
import com.example.project.repository.TrainersAccountsRepository;
import com.example.project.service.AccountService;
import com.example.project.service.AdaptiveProgramGenerator;
import com.example.project.service.ClubCapabilityService;
import com.example.project.service.ClubsService;
import com.example.project.service.CustomUserDetailsService;
import com.example.project.service.Event;
import com.example.project.service.MembersService;
import com.example.project.service.NewsService;
import com.example.project.service.PasswordValidationService;
import com.example.project.service.ProfileService;
import com.example.project.service.StaffScheduleService;
import com.example.project.service.StaffService;
import com.example.project.service.TrainersService;
import com.example.project.service.TrainingProgramService;
import com.example.project.service.TrainingRequest;
import com.example.project.service.TrainingScheduleService;
import com.example.project.service.TrainingTypeService;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Controller
@CrossOrigin("*")
@AllArgsConstructor
@Tag(name = "Основной контроллер", description = "Основные API для работы с пользователями, тренировками, программами и расписанием")
public class MainController {

    private final MembersService membersService;
    private final ClubsService clubsService;
    private final TrainingScheduleService trainingScheduleService;
    private final StaffScheduleService staffScheduleService;
    private final TrainersService trainersService;
    private final StaffService staffService;
    private final TrainingTypeService trainingTypeService;
    private final NewsService newsService;
    private final CustomUserDetailsService userDetailsService;
    private final AccountService accountService;
    private final MembersAccountsRepository membersAccountsRepo;
    private final TrainersAccountsRepository trainersAccountsRepo;
    private final StaffAccountsRepository staffAccountsRepo;
    private final ClubCapabilityService clubCapabilityService;
    private final TrainingProgramService trainingProgramService;
    private final PasswordValidationService passwordValidationService;
    private final AdaptiveProgramGenerator adaptiveProgramGenerator;
    private final StatsServiceClient statsServiceClient;

    @GetMapping("/programs/member/{id}")
    @Operation(summary = "Получить программы тренировок участника", description = "Возвращает все программы тренировок для указанного участника")
    @ApiResponse(responseCode = "200", description = "Программы успешно получены")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String memberPrograms(
            @Parameter(description = "ID участника", example = "123") @PathVariable Integer id,
            Model model) {
        if (funcBool(id)) return "redirect:/access-denied";

        Members member = membersService.getMember(id);
        List<TrainingProgram> programs = trainingProgramService.getMemberPrograms(id);
        TrainingProgram activeProgram = programs.stream()
                .filter(TrainingProgram::getIsActive)
                .findFirst()
                .orElse(null);

        Map<Integer, Integer> exerciseCounts = new HashMap<>();
        for (TrainingProgram program : programs) {
            exerciseCounts.put(program.getIdProgram(), trainingProgramService.getTotalExercisesCount(program));
        }

        Map<Integer, List<ProgramExercise>> sortedExercisesByDay = new HashMap<>();
        if (activeProgram != null) {
            List<ProgramDay> sortedDays = trainingProgramService.getSortedProgramDays(activeProgram);
            model.addAttribute("sortedProgramDays", sortedDays);
            for (ProgramDay day : sortedDays) {
                List<ProgramExercise> sortedExercises = trainingProgramService.getSortedExercises(day);
                sortedExercisesByDay.put(day.getDayNumber(), sortedExercises);
            }
        } else {
            model.addAttribute("sortedProgramDays", new ArrayList<>());
        }

        model.addAttribute("member", member);
        model.addAttribute("memberId", id);
        model.addAttribute("programs", programs);
        model.addAttribute("activeProgram", activeProgram);
        model.addAttribute("programRequest", new ProgramRequest());
        model.addAttribute("exerciseCounts", exerciseCounts);
        model.addAttribute("sortedExercisesByDay", sortedExercisesByDay);

        return "programs";
    }

    private boolean funcBool(@PathVariable @Parameter(description = "ID участника", example = "123") Integer id) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        if (!currentUserId.equals(id) || !"member".equals(currentUserRole)) {
            return true;
        }
        return false;
    }

    @GetMapping("/programs/generate/{id}")
    @Transactional(readOnly = true)
    @Operation(summary = "Форма для генерации программы тренировок", description = "Показывает форму для создания новой программы тренировок")
    @ApiResponse(responseCode = "200", description = "Форма успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String generateProgramForm(
            @Parameter(description = "ID участника", example = "123") @PathVariable Integer id,
            Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);

        if (!currentUserId.equals(id)) {
            return "redirect:/access-denied";
        }

        try {
            Members member = membersService.getMember(id);
            String clubName = "Неизвестный клуб";
            String clubSchedule = "Расписание не доступно";

            if (member != null && member.getClub() != null) {
                clubName = member.getClub().getClubName();
                clubSchedule = getClubScheduleFormatted(member.getClub());
            }

            int age = membersService.calculateAge(id);
            String ageGroup = membersService.getAgeGroup(id);
            boolean hasInbodyAnalysis = membersService.hasInbodyAnalysis(id);

            model.addAttribute("membersService", membersService);
            model.addAttribute("member", member);
            model.addAttribute("memberId", id);
            model.addAttribute("clubName", clubName);
            model.addAttribute("age", age);
            model.addAttribute("ageGroup", ageGroup);
            model.addAttribute("hasInbodyAnalysis", hasInbodyAnalysis);
            model.addAttribute("programRequest", new ProgramRequest());
            model.addAttribute("clubSchedule", clubSchedule);

            return "generate-program";

        } catch (Exception e) {
            log.error("Критическая ошибка при загрузке формы генерации программы для пользователя {}: {}", id,
                    e.getMessage());

            Members member = membersService.getMember(id);

            model.addAttribute("membersService", membersService);
            model.addAttribute("member", member);
            model.addAttribute("memberId", id);
            model.addAttribute("clubName", "Клуб не загружен");
            model.addAttribute("age", 0);
            model.addAttribute("ageGroup", "Неизвестно");
            model.addAttribute("hasInbodyAnalysis", false);
            model.addAttribute("programRequest", new ProgramRequest());

            return "generate-program";
        }
    }

    private String getClubScheduleFormatted(Clubs club) {
        if (club == null) {
            return "Расписание не доступно";
        }

        try {
            if (club.getSchedule() != null && !club.getSchedule().isEmpty()) {
                return parseClubSchedule(club.getSchedule());
            }
        } catch (Exception e) {
            log.warn("Не удалось распарсить расписание клуба: {}", e.getMessage());
        }

        return "Пн-Вс: 7:00-23:00";
    }

    private String parseClubSchedule(String scheduleJson) {
        if (scheduleJson.contains("Понедельник") || scheduleJson.contains("понедельник")) {
            return "Пн-Вс: 7:00-23:00";
        }
        return "Ежедневно: 7:00-23:00";
    }

    @PostMapping("/programs/generate/{id}")
    @Operation(summary = "Создать программу тренировок", description = "Генерирует новую программу тренировок для участника")
    @ApiResponse(responseCode = "200", description = "Программа успешно создана")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String generateProgram(
            @Parameter(description = "ID участника", example = "123") @PathVariable Integer id,
            @Parameter(description = "Данные для создания программы") @ModelAttribute ProgramRequest programRequest,
            Model model) {
        try {
            if (programRequest.getTrainingDays() == null || programRequest.getTrainingDays().isEmpty()) {
                throw new IllegalArgumentException("Выберите хотя бы один день для тренировок");
            }

            if (programRequest.getPreferredTime() == null) {
                throw new IllegalArgumentException("Выберите предпочтительное время тренировок");
            }

            Members member = membersService.getMember(id);
            if (member == null) {
                throw new IllegalArgumentException("Пользователь не найден");
            }

            if (!isScheduleCompatible(member.getClub(), programRequest)) {
                throw new IllegalArgumentException("Выбранное время тренировок не совместимо с расписанием клуба");
            }

            adaptiveProgramGenerator.generateAdaptiveProgram(id, programRequest);

            model.addAttribute("success", "Программа тренировок успешно создана с учетом возможностей клуба!");
            return "redirect:/programs/member/" + id;

        } catch (IllegalArgumentException e) {
            // Специфическая обработка для ошибок валидации
            log.warn("Ошибка валидации при создании программы для пользователя {}: {}", id, e.getMessage());
            handleProgramGenerationError(id, programRequest, model, e.getMessage());
            return "generate-program";

        } catch (Exception e) {
            // Общая обработка для непредвиденных ошибок
            log.error("Критическая ошибка при создании программы для пользователя {}: {}", id, e.getMessage(), e);
            handleProgramGenerationError(id, programRequest, model, "Внутренняя ошибка сервера: " + e.getMessage());
            return "generate-program";
        }
    }

    private void handleProgramGenerationError(Integer id, ProgramRequest programRequest, Model model, String errorMessage) {
        model.addAttribute("error", errorMessage);
        model.addAttribute("memberId", id);
        model.addAttribute("programRequest", programRequest);

        Members member = membersService.getMember(id);
        if (member != null) {
            model.addAttribute("member", member);
            model.addAttribute("age", membersService.calculateAge(id));
            model.addAttribute("ageGroup", membersService.getAgeGroup(id));
            model.addAttribute("hasInbodyAnalysis", membersService.hasInbodyAnalysis(id));
            model.addAttribute("clubName",
                    member.getClub() != null ? member.getClub().getClubName() : "Неизвестный клуб");
            model.addAttribute("clubSchedule", getClubScheduleFormatted(member.getClub()));
        } else {
            model.addAttribute("member", null);
            model.addAttribute("age", 0);
            model.addAttribute("ageGroup", "Неизвестно");
            model.addAttribute("hasInbodyAnalysis", false);
            model.addAttribute("clubName", "Клуб не загружен");
            model.addAttribute("clubSchedule", "Расписание не доступно");
        }

        model.addAttribute("membersService", membersService);
    }

    private boolean isScheduleCompatible(Clubs club, ProgramRequest request) {
        if (club == null) {
            log.warn("Клуб не указан для пользователя, пропускаем проверку расписания");
            return true;
        }

        String preferredTime = request.getPreferredTime();

        return switch (preferredTime) {
            case "УТРО" ->
                true;
            case "ДЕНЬ" ->
                true;
            case "ВЕЧЕР" ->
                club.getSchedule() != null
                && (club.getSchedule().contains("22:00")
                || club.getSchedule().contains("23:00"));
            default ->
                true;
        }; // 7:00-11:00 обычно в пределах работы
        // 11:00-17:00 обычно в пределах работы
        // Проверяем, что клуб работает до 22:00
    }

    @PostMapping("/programs/activate/{memberId}/{programId}")
    @Operation(summary = "Активировать программу тренировок", description = "Активирует выбранную программу тренировок для участника")
    @ApiResponse(responseCode = "200", description = "Программа успешно активирована")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String activateProgram(
            @Parameter(description = "ID участника", example = "123") @PathVariable Integer memberId,
            @Parameter(description = "ID программы", example = "456") @PathVariable Integer programId) {
        trainingProgramService.deactivateOtherPrograms(memberId, programId);

        TrainingProgram program = trainingProgramService.getProgram(programId);
        if (program != null) {
            program.setIsActive(true);
            trainingProgramService.saveProgram(program);
        }

        return "redirect:/programs/member/" + memberId;
    }

    @GetMapping("/")
    @Operation(summary = "Перенаправление на страницу входа", description = "Перенаправляет пользователя на страницу авторизации")
    public String redirectToLogin() {
        return "redirect:/login";
    }

    @GetMapping("/login")
    @Operation(summary = "Страница входа", description = "Отображает форму для входа в систему")
    public String login(
            @Parameter(description = "Параметр ошибки", example = "true") @RequestParam(required = false) String error,
            @Parameter(description = "Параметр выхода", example = "true") @RequestParam(required = false) String logout) {
        return "login";
    }

    @GetMapping("/logout")
    @Operation(summary = "Выход из системы", description = "Выполняет выход пользователя из системы")
    public String logout() {
        return "redirect:/login?logout";
    }

    @GetMapping("/registration")
    @Operation(summary = "Форма регистрации", description = "Отображает форму для регистрации нового пользователя")
    public String registrationForm(Model model) {
        model.addAttribute("clubs", clubsService.getAllClubs());
        return "registration";
    }

    @PostMapping("/registration")
    @Operation(summary = "Регистрация нового пользователя", description = "Регистрирует нового участника в системе")
    @ApiResponse(responseCode = "200", description = "Регистрация успешна")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации данных")
    public String registerUser(
            @Parameter(description = "Имя пользователя", example = "john_doe", required = true) @RequestParam String username,
            @Parameter(description = "Пароль", required = true) @RequestParam String password,
            @Parameter(description = "Подтверждение пароля", required = true) @RequestParam String confirmPassword,
            @Parameter(description = "Имя", example = "John", required = true) @RequestParam String firstName,
            @Parameter(description = "Фамилия", example = "Doe", required = true) @RequestParam String lastName,
            @Parameter(description = "Дата рождения", example = "1990-01-15", required = true) @RequestParam String birthDate,
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'", required = true) @RequestParam String clubName,
            @Parameter(description = "Пол (1 - мужской, 2 - женский)", example = "1", required = true) @RequestParam Integer gender,
            Model model) {
        if (!password.equals(confirmPassword)) {
            model.addAttribute("error", "Пароли не совпадают");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        PasswordValidationService.PasswordValidationResult passwordValidation = passwordValidationService
                .validatePassword(password, username, firstName, lastName);
        if (!passwordValidation.isValid()) {
            model.addAttribute("error", "Ненадежный пароль: " + passwordValidation.getErrorMessage());
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        if (accountService.getAccountInfo(username) != null) {
            model.addAttribute("error", "Пользователь с таким именем уже существует");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        try {
            LocalDate parsedBirthDate = LocalDate.parse(birthDate);

            if (parsedBirthDate.isAfter(LocalDate.now().minusYears(18))) {
                model.addAttribute("error", "Регистрация доступна только с 18 лет");
                model.addAttribute("clubs", clubsService.getAllClubs());
                return "registration";
            }

            boolean registrationSuccess = accountService.registerMember(
                    username, password, firstName, lastName, parsedBirthDate, clubName, gender);

            if (registrationSuccess) {
                model.addAttribute("success", "Регистрация прошла успешно! Теперь вы можете войти в систему.");
                return "registration";
            } else {
                model.addAttribute("error", "Ошибка при регистрации. Попробуйте позже.");
                model.addAttribute("clubs", clubsService.getAllClubs());
                return "registration";
            }

        } catch (Exception e) {
            model.addAttribute("error", "Ошибка при обработке данных: " + e.getMessage());
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }
    }

    @GetMapping("/profile/{role}/{id}")
    @Transactional
    @Operation(summary = "Профиль пользователя", description = "Отображает профиль пользователя в зависимости от его роли")
    @ApiResponse(responseCode = "200", description = "Профиль успешно загружен")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String profile(
            @Parameter(description = "ID пользователя", example = "123") @PathVariable Integer id,
            @Parameter(description = "Роль пользователя", example = "member", schema = @Schema(allowableValues = {
        "member", "trainer", "staff"})) @PathVariable String role,
            Model model) {
        model.addAttribute("role", role);
        model.addAttribute("membersService", membersService);

        if (methHelp(id, role)) return "redirect:/access-denied";

        switch (role) {
            case "member" -> {
                Members member = membersService.getMember(id);

                // Проверяем member на null перед использованием
                if (member == null) {
                    return "redirect:/not-found";
                }

                model.addAttribute("memberId", id);
                model.addAttribute("memberClub", member.getClub());
                model.addAttribute("member", member);

                // Убираем избыточную проверку member != null, так как мы уже проверили выше
                // Вместо этого проверяем member.getMembersAccount() на null
                if (member.getMembersAccount() != null) {
                    model.addAttribute("feedbacks", member.getMembersAccount().getFeedbacks());
                } else {
                    model.addAttribute("feedbacks", Collections.emptyList());
                }

                model.addAttribute("achievements", membersService.getSetOfMemberAchievements(id));

                LocalDateTime now = LocalDateTime.now();
                Set<TrainingSchedule> memberTrainings = trainingScheduleService.getTrainingsByMemberId(id);
                Set<TrainingSchedule> upcomingWorkouts = memberTrainings.stream()
                        .filter(workout -> workout.getSessionDate().isAfter(now))
                        .limit(3)
                        .collect(Collectors.toSet());

                model.addAttribute("workouts", upcomingWorkouts);
                model.addAttribute("workoutsCount", memberTrainings.stream()
                        .filter(workout -> workout.getSessionDate().isAfter(now))
                        .count());
                model.addAttribute("photoURL", membersService.getPhotoUrl(id));
                model.addAttribute("allNews", newsService.getAllNews());
            }

            case "trainer" -> {
                Trainers trainer = trainersService.getTrainer(id);

                // Проверяем trainer на null перед использованием
                if (trainer == null) {
                    return "redirect:/not-found";
                }

                model.addAttribute("trainerId", id);
                model.addAttribute("trainer", trainer);

                LocalDateTime nowTrainer = LocalDateTime.now();
                List<TrainingSchedule> trainerWorkouts = trainersService.getSetOfTrainingSchedule(id);
                Set<TrainingSchedule> upcomingTrainerWorkouts = trainerWorkouts.stream()
                        .filter(workout -> workout.getSessionDate().isAfter(nowTrainer))
                        .limit(3)
                        .collect(Collectors.toSet());

                model.addAttribute("workouts", upcomingTrainerWorkouts);
                model.addAttribute("workoutsCount", trainerWorkouts.stream()
                        .filter(workout -> workout.getSessionDate().isAfter(nowTrainer))
                        .count());
                model.addAttribute("photoURL", trainersService.getPhotoUrl(id));
            }

            case "staff" -> {
                Staff staff = staffService.getStaff(id);

                // Проверяем staff на null перед использованием
                if (staff == null) {
                    return "redirect:/not-found";
                }

                model.addAttribute("staffId", id);
                model.addAttribute("staff", staff);
                model.addAttribute("photoURL", staffService.getPhotoUrl(id));

                LocalDate nowStaff = LocalDate.now();
                List<StaffSchedule> staffSchedules = staffService.getSetOfStaffSchedule(id);
                Set<StaffSchedule> upcomingStaffSchedules = staffSchedules.stream()
                        .filter(work -> work.getDate().isAfter(nowStaff))
                        .limit(3)
                        .collect(Collectors.toSet());

                model.addAttribute("staffSchedule", upcomingStaffSchedules);
            }
            default -> {
                return "redirect:/not-found";
            }
        }
        return "profile";
    }

    private boolean methHelp(@PathVariable @Parameter(description = "ID пользователя", example = "123") Integer id, @PathVariable @Parameter(description = "Роль пользователя", example = "member", schema = @Schema(allowableValues = {
            "member", "trainer", "staff"})) String role) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        if (!currentUserId.equals(id) || !currentUserRole.equals(role)) {
            return true;
        }
        return false;
    }

    private final ProfileService profileService;

// Удалите старый метод profile() и замените на:
    @GetMapping("/profile/member/{id}")
    @Transactional
    public String profileMember(@PathVariable Integer id, Model model) {
        if (funcBool(id)) return "redirect:/access-denied";

        ProfileService.ProfileData data = profileService.getMemberProfile(id);
        if (data == null) {
            return "redirect:/not-found";
        }

        model.addAttribute("role", "member");
        model.addAttribute("membersService", membersService);
        model.addAttribute("memberId", data.getMemberId());
        model.addAttribute("memberClub", data.getMemberClub());
        model.addAttribute("member", data.getMember());
        model.addAttribute("feedbacks", data.getFeedbacks());
        model.addAttribute("achievements", data.getAchievements());
        model.addAttribute("workouts", data.getWorkouts());
        model.addAttribute("workoutsCount", data.getWorkoutsCount());
        model.addAttribute("photoURL", data.getPhotoURL());
        model.addAttribute("allNews", data.getAllNews());

        return "profile";
    }

    @GetMapping("/profile/trainer/{id}")
    @Transactional
    public String profileTrainer(@PathVariable Integer id, Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        if (!currentUserId.equals(id) || !"trainer".equals(currentUserRole)) {
            return "redirect:/access-denied";
        }

        ProfileService.ProfileData data = profileService.getTrainerProfile(id);
        if (data == null) {
            return "redirect:/not-found";
        }

        model.addAttribute("role", "trainer");
        model.addAttribute("trainerId", data.getTrainerId());
        model.addAttribute("trainer", data.getTrainer());
        model.addAttribute("workouts", data.getWorkouts());
        model.addAttribute("workoutsCount", data.getWorkoutsCount());
        model.addAttribute("photoURL", data.getPhotoURL());

        return "profile";
    }

    @GetMapping("/profile/staff/{id}")
    @Transactional
    public String profileStaff(@PathVariable Integer id, Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        if (!currentUserId.equals(id) || !"staff".equals(currentUserRole)) {
            return "redirect:/access-denied";
        }

        ProfileService.ProfileData data = profileService.getStaffProfile(id);
        if (data == null) {
            return "redirect:/not-found";
        }

        model.addAttribute("role", "staff");
        model.addAttribute("staffId", data.getStaffId());
        model.addAttribute("staff", data.getStaff());
        model.addAttribute("photoURL", data.getPhotoURL());
        model.addAttribute("staffSchedule", data.getStaffSchedule());

        return "profile";
    }

    @GetMapping("/profile/member/{id}/news")
    @ResponseBody
    @Transactional(readOnly = true)
    @Operation(summary = "Получить новости для участника", description = "Возвращает список новостей, отфильтрованных по клубу участника")
    @ApiResponse(responseCode = "200", description = "Новости успешно получены", content = @Content(mediaType = "application/json", schema = @Schema(implementation = NewsDTO.class)))
    public List<NewsDTO> getMemberNews(
            @Parameter(description = "ID участника", example = "123") @PathVariable Integer id,
            @Parameter(description = "Название клуба для фильтрации", example = "Фитнес Центр 'Энергия'") @RequestParam(required = false) String club) {
        // Убираем избыточную проверку club != null, так как String.isEmpty() уже проверяет на null
        if (club != null && !club.isEmpty()) {
            return newsService.getNewsByClubDTO(club);
        }
        return newsService.getAllNewsWithClubsDTO();
    }

    @GetMapping("/calendar/{role}/{id}")
    @Operation(summary = "Календарь пользователя", description = "Отображает календарь тренировок/работы в зависимости от роли")
    @ApiResponse(responseCode = "200", description = "Календарь успешно загружен")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String calendar(
            @Parameter(description = "ID пользователя", example = "123") @PathVariable Integer id,
            @Parameter(description = "Роль пользователя", example = "member", schema = @Schema(allowableValues = {
        "member", "trainer", "staff"})) @PathVariable String role,
            Model model) {
        model.addAttribute("role", role);
        model.addAttribute("TrainersSet", trainersService.getAllTrainers());
        model.addAttribute("TrainingTypeSet", trainingScheduleService.getTrainingTypes());

        if (methHelp(id, role)) return "redirect:/access-denied";

        switch (role) {
            case "member" -> {
                Members member = membersService.getMember(id);
                model.addAttribute("member", member);
                model.addAttribute("memberId", id);
            }
            case "trainer" -> {
                Trainers trainer = trainersService.getTrainer(id);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", id);
                model.addAttribute("TrainingRequest", new TrainingRequest());
            }
            case "staff" -> {
                Staff staff = staffService.getStaff(id);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", id);
            }
            default -> {
            }
        }

        return "calendar";
    }

    @GetMapping("/calendar/training/{id}")
    @Transactional(readOnly = true)
    @Operation(summary = "Детали тренировки", description = "Отображает детальную информацию о тренировке")
    @ApiResponse(responseCode = "200", description = "Информация о тренировке загружена")
    @ApiResponse(responseCode = "404", description = "Тренировка не найдена")
    public String training(
            @Parameter(description = "ID тренировки", example = "789") @PathVariable Integer id,
            Model model) {
        TrainingSchedule workout = trainingScheduleService.getTrainingSchedule(id);

        if (workout == null) {
            return "redirect:/not-found";
        }

        model.addAttribute("training", workout);
        model.addAttribute("session_date", workout.getSessionDate());

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);
        Integer userId = userDetailsService.getUserId(username);

        switch (role) {
            case "member" -> {
                Members member = membersService.getMember(userId);
                model.addAttribute("trainer", trainingScheduleService.getTrainer(id));
                model.addAttribute("member", member);
                model.addAttribute("memberId", userId);
            }
            case "trainer" -> {
                Trainers trainer = trainersService.getTrainer(userId);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", userId);
            }
            case "staff" -> {
                Staff staff = staffService.getStaff(userId);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", userId);
            }
            default -> {
            }
        }
        model.addAttribute("role", role);
        return "training";
    }

    @PostMapping("/calendar/training/subscribe")
    @Operation(summary = "Записаться на тренировку", description = "Добавляет участника на тренировку")
    @ApiResponse(responseCode = "200", description = "Успешно записан на тренировку")
    @ApiResponse(responseCode = "400", description = "Ошибка при записи")
    public String trainingSignup(
            @Parameter(description = "Данные для записи на тренировку") @ModelAttribute TrainingRequest form,
            Model model) {
        Integer memberId = form.getMemberId();
        Integer trainingId = form.getTrainingId();

        Members member = membersService.getMember(memberId);
        TrainingSchedule training = trainingScheduleService.getTrainingSchedule(trainingId);

        if (member != null && training != null) {
            Set<TrainingSchedule> memberTrainings = membersService.getSetOfTrainingSchedule(memberId);
            memberTrainings.add(training);
            membersService.save(member);
        }

        return "redirect:/calendar/training/" + trainingId;
    }

    @PostMapping("/calendar/training/unsubscribe")
    @Transactional
    @Operation(summary = "Отписаться от тренировки", description = "Удаляет участника из списка записанных на тренировку")
    @ApiResponse(responseCode = "200", description = "Успешно отписан от тренировки")
    public String trainingUnsubscribe(
            @Parameter(description = "Данные для отписки от тренировки") @ModelAttribute TrainingRequest form) {
        Integer memberId = form.getMemberId();
        Integer trainingId = form.getTrainingId();

        try {
            Members member = membersService.getMember(memberId);
            TrainingSchedule training = trainingScheduleService.getTrainingSchedule(trainingId);

            if (member != null && training != null) {
                member.getTrainingSchedules().remove(training);
                membersService.save(member);
            }
            return "redirect:/calendar/member/" + memberId;

        } catch (Exception e) {
            return "redirect:/calendar/member/" + memberId;
        }
    }

    @PostMapping("/calendar/training/add")
    @Operation(summary = "Добавить тренировку", description = "Создает новую тренировку (для тренеров)")
    @ApiResponse(responseCode = "200", description = "Тренировка успешно создана")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    public String trainingAdd(
            @Parameter(description = "Данные для создания тренировки") @ModelAttribute TrainingRequest form,
            BindingResult result,
            Model model) {
        if (result.hasErrors()) {
            return "redirect:/calendar/trainer/" + form.getTrainerId();
        }
        try {
            TrainingSchedule training = new TrainingSchedule();
            training.setTrainer(trainersService.getTrainer(form.getTrainerId()));
            training.setTrainingType(trainingTypeService.getTrainingType(form.getTrainingTypeId()));
            training.setSessionDate(form.getSessionDate());
            training.setSessionTime(form.getSessionTime());
            TrainingSchedule savedTraining = trainingScheduleService.save(training);
            return "redirect:/calendar/training/" + savedTraining.getIdSession();
        } catch (Exception e) {
            model.addAttribute("error", "Ошибка при создании тренировки");
            return "redirect:/calendar/trainer/" + form.getTrainerId();
        }
    }

    @GetMapping("/statistic/{role}/{id}")
    @Operation(summary = "Статистика пользователя", description = "Отображает статистику пользователя (достижения, тренировки и т.д.)")
    @ApiResponse(responseCode = "200", description = "Статистика успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String statistic(
            @Parameter(description = "ID пользователя", example = "123") @PathVariable Integer id,
            @Parameter(description = "Роль пользователя", example = "member", schema = @Schema(allowableValues = {
        "member", "trainer", "staff"})) @PathVariable String role,
            Model model) {
        if (methHelp(id, role)) return "redirect:/access-denied";

        switch (role) {
            case "member" -> {
                Members member = membersService.getMember(id);
                Set<Achievements> achievements = membersService.getSetOfMemberAchievements(id);
                model.addAttribute("achievements", achievements);
                model.addAttribute("member", member);
                model.addAttribute("memberId", id);

                // Добавляем статистику тренировок
                Set<TrainingSchedule> memberTrainings = trainingScheduleService.getTrainingsByMemberId(id);
                model.addAttribute("statistics", memberTrainings);
            }
            case "trainer" -> {
                Trainers trainer = trainersService.getTrainer(id);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", id);

                // Для тренеров показываем статистику их тренировок
                List<TrainingSchedule> trainerWorkouts = trainersService.getSetOfTrainingSchedule(id);
                model.addAttribute("statistics", trainerWorkouts);
            }
            case "staff" -> {
                Staff staff = staffService.getStaff(id);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", id);

                // Для персонала добавляем данные из микросервиса статистики
                try {
                    // Получаем данные из микросервиса статистики
                    Integer todayVisits = statsServiceClient.getTodayVisits(1); // ID клуба по умолчанию
                    model.addAttribute("todayVisits", todayVisits);

                    boolean statsServiceHealthy = statsServiceClient.isServiceHealthy();
                    model.addAttribute("statsServiceHealthy", statsServiceHealthy);

                    if (statsServiceHealthy) {
                        List<Map<String, Object>> topMembers = statsServiceClient.getTopActiveMembers();
                        model.addAttribute("topMembers", topMembers);

                        Map<String, Object> weeklyStats = statsServiceClient.getVisitsStats(1, "week");
                        model.addAttribute("weeklyStats", weeklyStats);
                    }

                    log.info("Stats service data loaded for staff ID: {}", id);

                } catch (Exception e) {
                    log.error("Error loading stats service data: {}", e.getMessage());
                    model.addAttribute("statsServiceHealthy", false);
                    model.addAttribute("statsError", "Не удалось загрузить данные статистики");
                }
            }
            default -> {
            }
        }
        model.addAttribute("role", role);

        return "statistic";
    }

    @GetMapping("/trainers")
    @Operation(summary = "Список тренеров", description = "Отображает список всех тренеров в системе")
    public String trainers(Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);
        Integer userId = userDetailsService.getUserId(username);

        switch (role) {
            case "member" -> {
                Members member = membersService.getMember(userId);
                model.addAttribute("member", member);
                model.addAttribute("memberId", userId);
            }
            case "trainer" -> {
                Trainers trainer = trainersService.getTrainer(userId);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", userId);
            }
            case "staff" -> {
                Staff staff = staffService.getStaff(userId);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", userId);
            }
            default -> {
            }
        }
        model.addAttribute("role", role);
        model.addAttribute("trainers", trainersService.getAllTrainers());
        model.addAttribute("trainingRequest", new TrainingRequest());

        return "trainers";
    }

    @PostMapping("/trainers/subscribe")
    @Transactional
    @Operation(summary = "Записаться к тренеру", description = "Записывает участника на персональную тренировку к тренеру")
    @ApiResponse(responseCode = "200", description = "Успешно записан к тренеру")
    @ApiResponse(responseCode = "400", description = "Ошибка при записи")
    public String subscribe(
            @Parameter(description = "Данные для записи к тренеру") @ModelAttribute TrainingRequest form,
            Model model,
            Principal principal) {
        try {
            // Проверки безопасности...
            String username = principal.getName();
            Integer currentMemberId = userDetailsService.getUserId(username);

            if (!currentMemberId.equals(form.getMemberId())) {
                model.addAttribute("error", "Доступ запрещен");
                return "redirect:/trainers";
            }

            // Создаем тренировку
            Trainers trainer = trainersService.getTrainer(form.getTrainerId());
            TrainingType trainingType = trainingTypeService.getTrainingType(5);

            TrainingSchedule personalTraining = new TrainingSchedule();
            personalTraining.setTrainer(trainer);
            personalTraining.setTrainingType(trainingType);
            personalTraining.setSessionDate(form.getSessionDate());
            personalTraining.setSessionTime(60);

            TrainingSchedule savedTraining = trainingScheduleService.save(personalTraining);

            // Связываем члена с тренировкой
            membersService.addTrainingToMember(form.getMemberId(), savedTraining.getIdSession());

            return "redirect:/calendar/training/" + savedTraining.getIdSession();

        } catch (Exception e) {
            model.addAttribute("error", "Ошибка при записи на персональную тренировку: " + e.getMessage());
            return "redirect:/trainers";
        }
    }

    @GetMapping("/trainings")
    @ResponseBody
    @Operation(summary = "Получить список тренировок", description = "Возвращает список тренировок с фильтрацией по различным параметрам")
    @ApiResponse(responseCode = "200", description = "Список тренировок успешно получен", content = @Content(mediaType = "application/json"))
    public String getTrainings(
            @Parameter(description = "ID тренеров для фильтрации") @RequestParam(value = "id_trainer", required = false) Set<Integer> trainerId,
            @Parameter(description = "ID типов тренировок для фильтрации") @RequestParam(value = "id_training_type", required = false) Set<Integer> trainingTypeId,
            @Parameter(description = "Начальная дата для фильтрации", example = "2024-01-15") @RequestParam(value = "session_date_start", required = false) String sessionDateStart,
            @Parameter(description = "Конечная дата для фильтрации", example = "2024-01-31") @RequestParam(value = "session_date_end", required = false) String sessionDateEnd,
            @Parameter(description = "Минимальное время тренировки (в минутах)", example = "30") @RequestParam(value = "session_time_start", required = false) Integer sessionTimeStart,
            @Parameter(description = "Максимальное время тренировки (в минутах)", example = "120") @RequestParam(value = "session_time_end", required = false) Integer sessionTimeEnd,
            @Parameter(description = "Фильтр по расписанию тренера (0/1)", example = "0") @RequestParam(value = "trainer_schedule", required = false, defaultValue = "0") Integer trainerSchedule) {

        if (trainerId != null && trainerId.isEmpty()) {
            trainerId = null;
        }
        if (trainingTypeId != null && trainingTypeId.isEmpty()) {
            trainingTypeId = null;
        }
        if (sessionTimeStart != null && sessionTimeStart <= 0) {
            sessionTimeStart = null;
        }
        if (sessionTimeEnd != null && sessionTimeEnd <= 0) {
            sessionTimeEnd = null;
        }

        LocalDateTime startDate = null;
        LocalDateTime endDate = null;

        try {
            Set<TrainingSchedule> trainingScheduleSet = trainingScheduleService.getTrainingSet(
                    trainerId, trainingTypeId, startDate, endDate, sessionTimeStart, sessionTimeEnd);

            Set<Event> eventsSet = trainingScheduleService.trainingScheduleToEventSet(
                    trainingScheduleSet, trainerSchedule != null ? trainerSchedule : 0);

            ObjectMapper mapper = new ObjectMapper();
            return mapper.writerWithDefaultPrettyPrinter().writeValueAsString(eventsSet);
        } catch (JsonProcessingException e) {
            log.error("Ошибка при сериализации событий в JSON", e); // Замена printStackTrace
            return "[]";
        }
    }

    @GetMapping("/calendar/work/events")
    @ResponseBody
    @Operation(summary = "Получить рабочие события", description = "Возвращает список рабочих событий для staff")
    @ApiResponse(responseCode = "200", description = "События успешно получены", content = @Content(mediaType = "application/json"))
    public String getWorks() {
        List<StaffSchedule> staffScheduleSet = staffScheduleService.getAllStaffSchedules();
        List<Event> eventsSet = staffScheduleService.staffScheduleToEvents(staffScheduleSet);

        ObjectMapper mapper = new ObjectMapper();
        try {
            return mapper.writerWithDefaultPrettyPrinter().writeValueAsString(eventsSet);
        } catch (JsonProcessingException e) {
            log.error("Ошибка при сериализации рабочих событий в JSON", e); // Замена printStackTrace
            return "[]";
        }
    }

    // === API для анализа возможностей клуба ===
    @GetMapping("/club/capabilities/{clubName}")
    @ResponseBody
    @Operation(summary = "Анализ возможностей клуба", description = "Возвращает анализ возможностей и оснащенности клуба")
    @ApiResponse(responseCode = "200", description = "Анализ успешно выполнен", content = @Content(mediaType = "application/json"))
    public Map<String, Object> getClubCapabilities(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName) {
        return clubCapabilityService.analyzeClubCapabilities(clubName);
    }

    @GetMapping("/club/equipment/{clubName}")
    @ResponseBody
    @Operation(summary = "Оборудование клуба", description = "Возвращает список оборудования с количеством для указанного клуба")
    @ApiResponse(responseCode = "200", description = "Оборудование успешно получено", content = @Content(mediaType = "application/json"))
    public Map<String, Integer> getClubEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName) {
        return clubCapabilityService.getClubEquipmentSummary(clubName);
    }

    @GetMapping("/direct-diagnose")
    @ResponseBody
    @Operation(summary = "Прямая диагностика пользователя", description = "Диагностический метод для проверки данных пользователя в БД")
    @ApiResponse(responseCode = "200", description = "Диагностика выполнена")
    public String directDiagnose(
            @Parameter(description = "Имя пользователя для диагностики", example = "andrew.kachalkin") @RequestParam String username) {
        // Добавляем проверку на null для username
        if (username == null || username.trim().isEmpty()) {
            return "Ошибка: имя пользователя не может быть пустым";
        }

        StringBuilder result = new StringBuilder();
        result.append("=== DIRECT DIAGNOSIS: ").append(username).append(" ===\n\n");

        result.append("1. TrainersAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<TrainersAccounts> trainer = trainersAccountsRepo.findById(username);
            if (trainer.isPresent()) {
                result.append("    FOUND: ").append(trainer.get().getUsername()).append("\n");
                result.append("   Password: ").append(trainer.get().getPassword()).append("\n");
                result.append("   Role: ").append(trainer.get().getUserRole()).append("\n");
                result.append("   Trainer ID: ").append(trainer.get().getTrainer().getIdTrainer()).append("\n");
            } else {
                result.append("    NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("    ERROR: ").append(e.getMessage()).append("\n");
        }

        result.append("\n2. MembersAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<MembersAccounts> member = membersAccountsRepo.findById(username);
            if (member.isPresent()) {
                result.append("    FOUND\n");
            } else {
                result.append("    NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("    ERROR: ").append(e.getMessage()).append("\n");
        }

        result.append("\n3. StaffAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<StaffAccounts> staff = staffAccountsRepo.findById(username);
            if (staff.isPresent()) {
                result.append("    FOUND\n");
            } else {
                result.append("    NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("    ERROR: ").append(e.getMessage()).append("\n");
        }

        result.append("\n4. All trainers in database:\n");
        try {
            List<TrainersAccounts> allTrainers = trainersAccountsRepo.findAll();
            result.append("   Total trainers: ").append(allTrainers.size()).append("\n");
            for (TrainersAccounts t : allTrainers) {
                result.append("   - ").append(t.getUsername())
                        .append(" (ID: ").append(t.getTrainer().getIdTrainer()).append(")\n");
            }
        } catch (Exception e) {
            result.append("    ERROR: ").append(e.getMessage()).append("\n");
        }

        return result.toString();
    }

    @GetMapping("/diagnose")
    @ResponseBody
    @Operation(summary = "Диагностика пользователя", description = "Проверка существования пользователя в системе")
    @ApiResponse(responseCode = "200", description = "Диагностика выполнена")
    public String diagnose(
            @Parameter(description = "Имя пользователя", example = "andrew.kachalkin") @RequestParam String username) {
        return accountService.diagnoseUser(username);
    }

    @GetMapping("/check-password")
    @ResponseBody
    @Operation(summary = "Проверка пароля", description = "Проверяет корректность пароля для указанного пользователя")
    @ApiResponse(responseCode = "200", description = "Проверка выполнена")
    public String checkPassword(
            @Parameter(description = "Имя пользователя", example = "andrew.kachalkin") @RequestParam String username,
            @Parameter(description = "Пароль для проверки") @RequestParam String password) {
        boolean result = accountService.checkPassword(username, password);
        return String.format("Password check for %s: %s", username, result ? " VALID" : " INVALID");
    }

    @GetMapping("/access-denied")
    @Operation(summary = "Доступ запрещен", description = "Страница отображается при отсутствии доступа")
    public String accessDenied() {
        return "access-denied";
    }

    @GetMapping("/not-found")
    @Operation(summary = "Не найдено", description = "Страница отображается при отсутствии запрашиваемого ресурса")
    public String notFound() {
        return "not-found";
    }
}
