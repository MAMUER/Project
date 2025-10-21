package com.example.project.controller;

import java.io.IOException;
import java.security.Principal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.stream.Collectors;

import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.ui.Model;
import org.springframework.validation.BindingResult;

import com.example.project.dto.NewsDTO;
import com.example.project.dto.ProgramRequest;
import com.example.project.model.Achievements;
import com.example.project.model.EquipmentStatistics;
import com.example.project.model.Members;
import com.example.project.model.Staff;
import com.example.project.model.StaffSchedule;
import com.example.project.model.Trainers;
import com.example.project.model.TrainingProgram;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.TrainingType;
import com.example.project.model.Accounts.MembersAccounts;
import com.example.project.model.Accounts.StaffAccounts;
import com.example.project.model.Accounts.TrainersAccounts;
import com.example.project.repository.MembersAccountsRepository;
import com.example.project.repository.StaffAccountsRepository;
import com.example.project.repository.TrainersAccountsRepository;
import com.example.project.service.AccountService;
import com.example.project.service.ClubsService;
import com.example.project.service.CustomUserDetailsService;
import com.example.project.service.Event;
import com.example.project.service.MembersService;
import com.example.project.service.NewsService;
import com.example.project.service.ProgramGeneratorService;
import com.example.project.service.StaffScheduleService;
import com.example.project.service.StaffService;
import com.example.project.service.TrainersService;
import com.example.project.service.TrainingProgramService;
import com.example.project.service.TrainingRequest;
import com.example.project.service.TrainingScheduleService;
import com.example.project.service.TrainingTypeService;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.AllArgsConstructor;

import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

@Controller
@CrossOrigin("*")
@AllArgsConstructor
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
    // Добавьте эти зависимости в MainController
    private final TrainingProgramService trainingProgramService;
    private final ProgramGeneratorService programGeneratorService;

    // Добавьте эти методы в MainController
    @GetMapping("/programs/member/{id}")
    public String memberPrograms(@PathVariable Integer id, Model model) {
        // Проверка доступа
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        if (!currentUserId.equals(id) || !"member".equals(currentUserRole)) {
            return "redirect:/access-denied";
        }

        Members member = membersService.getMember(id);
        List<TrainingProgram> programs = trainingProgramService.getMemberPrograms(id);
        TrainingProgram activeProgram = programs.stream()
                .filter(TrainingProgram::getIsActive)
                .findFirst()
                .orElse(null);

        // Добавляем счетчики упражнений для каждой программы
        Map<Integer, Integer> exerciseCounts = new HashMap<>();
        for (TrainingProgram program : programs) {
            exerciseCounts.put(program.getIdProgram(), trainingProgramService.getTotalExercisesCount(program));
        }

        model.addAttribute("member", member);
        model.addAttribute("memberId", id);
        model.addAttribute("programs", programs);
        model.addAttribute("activeProgram", activeProgram);
        model.addAttribute("programRequest", new ProgramRequest());
        model.addAttribute("exerciseCounts", exerciseCounts); // Добавляем счетчики

        return "programs";
    }

    @GetMapping("/programs/generate/{id}")
    public String generateProgramForm(@PathVariable Integer id, Model model) {
        // Проверка доступа
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);

        if (!currentUserId.equals(id)) {
            return "redirect:/access-denied";
        }

        model.addAttribute("memberId", id);
        model.addAttribute("programRequest", new ProgramRequest());
        return "generate-program";
    }

    @PostMapping("/programs/generate/{id}")
    public String generateProgram(@PathVariable Integer id,
            @ModelAttribute ProgramRequest programRequest,
            Model model) {
        try {
            TrainingProgram program = programGeneratorService.generateProgram(id, programRequest);
            // Программа автоматически сохраняется в generateProgram методе
            // Мы можем добавить логирование или другую обработку при необходимости
            System.out.println("Создана программа: " + program.getProgramName() + " для пользователя " + id);

            model.addAttribute("success", "Программа тренировок успешно создана!");
            return "redirect:/programs/member/" + id;
        } catch (Exception e) {
            model.addAttribute("error", "Ошибка при создании программы: " + e.getMessage());
            model.addAttribute("memberId", id);
            model.addAttribute("programRequest", programRequest);
            return "generate-program";
        }
    }

    @PostMapping("/programs/activate/{memberId}/{programId}")
    public String activateProgram(@PathVariable Integer memberId,
            @PathVariable Integer programId) {
        trainingProgramService.deactivateOtherPrograms(memberId, programId);

        TrainingProgram program = trainingProgramService.getProgram(programId);
        if (program != null) {
            program.setIsActive(true);
            trainingProgramService.saveProgram(program);
        }

        return "redirect:/programs/member/" + memberId;
    }

    @GetMapping("/")
    public String redirectToLogin() {
        return "redirect:/login";
    }

    @GetMapping("/login")
    public String login(@RequestParam(required = false) String error,
            @RequestParam(required = false) String logout) {
        return "login";
    }

    @GetMapping("/logout")
    public String logout() {
        return "redirect:/login?logout";
    }

    @GetMapping("/registration")
    public String registrationForm(Model model) {
        // Загружаем список клубов для выбора
        model.addAttribute("clubs", clubsService.getAllClubs());
        return "registration";
    }

    @PostMapping("/registration")
    public String registerUser(@RequestParam String username,
            @RequestParam String password,
            @RequestParam String confirmPassword,
            @RequestParam String email,
            @RequestParam String firstName,
            @RequestParam String lastName,
            @RequestParam String phoneNumber,
            @RequestParam String birthDate,
            @RequestParam String clubName,
            @RequestParam Integer gender,
            @RequestParam Integer membershipPeriod,
            Model model) {

        // Валидация паролей
        if (!password.equals(confirmPassword)) {
            model.addAttribute("error", "Пароли не совпадают");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        // Проверка на существование пользователя
        if (accountService.getAccountInfo(username) != null) {
            model.addAttribute("error", "Пользователь с таким именем уже существует");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        // Валидация email
        if (!isValidEmail(email)) {
            model.addAttribute("error", "Некорректный email адрес");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        // Валидация телефона
        if (!isValidPhoneNumber(phoneNumber)) {
            model.addAttribute("error", "Некорректный номер телефона");
            model.addAttribute("clubs", clubsService.getAllClubs());
            return "registration";
        }

        try {
            // Парсинг даты рождения
            LocalDate parsedBirthDate = LocalDate.parse(birthDate);

            // Проверка возраста (минимум 14 лет)
            if (parsedBirthDate.isAfter(LocalDate.now().minusYears(14))) {
                model.addAttribute("error", "Регистрация доступна с 14 лет");
                model.addAttribute("clubs", clubsService.getAllClubs());
                return "registration";
            }

            // Регистрация пользователя
            boolean registrationSuccess = accountService.registerMember(
                    username, password, email, firstName, lastName,
                    phoneNumber, parsedBirthDate, clubName, gender, membershipPeriod);

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

    // Вспомогательные методы валидации
    private boolean isValidEmail(String email) {
        String emailRegex = "^[A-Za-z0-9+_.-]+@(.+)$";
        return email != null && email.matches(emailRegex);
    }

    private boolean isValidPhoneNumber(String phoneNumber) {
        String phoneRegex = "^\\+?[0-9]{10,15}$";
        return phoneNumber != null && phoneNumber.matches(phoneRegex);
    }

    @GetMapping("/profile/{role}/{id}")
    @Transactional
    public String profile(@PathVariable Integer id, @PathVariable String role, Model model) {
        model.addAttribute("role", role);
        model.addAttribute("membersService", membersService);

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);

        // Проверка доступа
        if (!currentUserId.equals(id) || !currentUserRole.equals(role)) {
            return "redirect:/access-denied";
        }

        switch (role) {
            case "member":
                Members member = membersService.getMember(id);
                model.addAttribute("memberId", id);
                model.addAttribute("memberClub", member.getClub());
                // Получаем фидбэки через аккаунт
                if (member != null && member.getMembersAccount() != null) {
                    model.addAttribute("feedbacks", member.getMembersAccount().getFeedbacks());
                } else {
                    model.addAttribute("feedbacks", Collections.emptyList());
                }
                model.addAttribute("roleName", member.getMembershipRole().getRoleName());
                model.addAttribute("member", member);
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
                boolean hasActiveSubscription = member.getEndTrialDate() != null &&
                        member.getEndTrialDate().isAfter(LocalDate.now());
                model.addAttribute("hasActiveSubscription", hasActiveSubscription);
                break;

            case "trainer":
                Trainers trainer = trainersService.getTrainer(id);
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
                break;

            case "staff":
                Staff staff = staffService.getStaff(id);
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
                break;
            default:
                break;
        }
        return "profile";
    }

    @GetMapping("/profile/member/{id}")
    public String profileMember(@PathVariable Integer id, Model model) {
        return profile(id, "member", model);
    }

    @GetMapping("/profile/trainer/{id}")
    public String profileTrainer(@PathVariable Integer id, Model model) {
        return profile(id, "trainer", model);
    }

    @GetMapping("/profile/staff/{id}")
    public String profileStaff(@PathVariable Integer id, Model model) {
        return profile(id, "staff", model);
    }

    @GetMapping("/profile/member/{id}/news")
    @ResponseBody
    @Transactional(readOnly = true)
    public List<NewsDTO> getMemberNews(@PathVariable Integer id,
            @RequestParam(required = false) String club) {
        if (club != null && !club.isEmpty()) {
            return newsService.getNewsByClubDTO(club);
        } else {
            return newsService.getAllNewsWithClubsDTO();
        }
    }

    @GetMapping("/calendar/{role}/{id}")
    public String calendar(@PathVariable Integer id, @PathVariable String role, Model model) {
        model.addAttribute("role", role);
        model.addAttribute("TrainersSet", trainersService.getAllTrainers());
        model.addAttribute("TrainingTypeSet", trainingScheduleService.getTrainingTypes());

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();

        // Проверка доступа
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);
        if (!currentUserId.equals(id) || !currentUserRole.equals(role)) {
            return "redirect:/access-denied";
        }

        switch (role) {
            case "member":
                Members member = membersService.getMember(id);
                model.addAttribute("member", member);
                model.addAttribute("memberId", id);
                break;
            case "trainer":
                Trainers trainer = trainersService.getTrainer(id);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", id);
                model.addAttribute("TrainingRequest", new TrainingRequest());
                break;
            case "staff":
                Staff staff = staffService.getStaff(id);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", id);
                break;
            default:
                break;
        }

        return "calendar";
    }

    @GetMapping("/calendar/training/{id}")
    @Transactional(readOnly = true) // Добавьте эту аннотацию
    public String training(@PathVariable Integer id, Model model) {
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
            case "member":
                Members member = membersService.getMember(userId);
                boolean isSignedUp = membersService.getTrainingSchedules(userId).contains(workout);
                model.addAttribute("trainer", trainingScheduleService.getTrainer(id));
                model.addAttribute("isSignedUp", isSignedUp);
                model.addAttribute("member", member);
                model.addAttribute("memberId", userId);
                break;
            case "trainer":
                Trainers trainer = trainersService.getTrainer(userId);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", userId);
                break;
            case "staff":
                Staff staff = staffService.getStaff(userId);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", userId);
                break;
            default:
                break;
        }
        model.addAttribute("role", role);
        return "training";
    }

    @PostMapping("/calendar/training/subscribe")
    public String trainingSignup(@ModelAttribute TrainingRequest form, Model model) {
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
    @Transactional // Добавляем транзакцию
    public String trainingUnsubscribe(@ModelAttribute TrainingRequest form) {
        Integer memberId = form.getMemberId();
        Integer trainingId = form.getTrainingId();

        try {
            Members member = membersService.getMember(memberId);
            TrainingSchedule training = trainingScheduleService.getTrainingSchedule(trainingId);

            if (member != null && training != null) {
                // Работаем напрямую с коллекцией entity
                member.getTrainingSchedules().remove(training);
                membersService.save(member);
            }

            return "redirect:/calendar/member/" + memberId;

        } catch (Exception e) {
            return "redirect:/calendar/member/" + memberId;
        }
    }

    @PostMapping("/calendar/training/add")
    public String trainingAdd(@ModelAttribute TrainingRequest form, BindingResult result, Model model) {
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
    public String statistic(@PathVariable Integer id, @PathVariable String role, Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();

        // Проверка доступа
        Integer currentUserId = userDetailsService.getUserId(username);
        String currentUserRole = userDetailsService.getUserRole(username);
        if (!currentUserId.equals(id) || !currentUserRole.equals(role)) {
            return "redirect:/access-denied";
        }

        switch (role) {
            case "member":
                Members member = membersService.getMember(id);
                Set<EquipmentStatistics> statistics = membersService.getSetOfEquipmentStatistics(id);
                Set<Achievements> achievements = membersService.getSetOfMemberAchievements(id);
                model.addAttribute("statistics", statistics);
                model.addAttribute("achievements", achievements);
                model.addAttribute("member", member);
                model.addAttribute("memberId", id);
                break;
            case "trainer":
                Trainers trainer = trainersService.getTrainer(id);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", id);
                break;
            case "staff":
                Staff staff = staffService.getStaff(id);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", id);
                break;
            default:
                break;
        }
        model.addAttribute("role", role);

        return "statistic";
    }

    @GetMapping("/trainers")
    public String trainers(Model model) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        String role = userDetailsService.getUserRole(username);
        Integer userId = userDetailsService.getUserId(username);

        switch (role) {
            case "member":
                Members member = membersService.getMember(userId);
                model.addAttribute("member", member);
                model.addAttribute("memberId", userId);
                break;
            case "trainer":
                Trainers trainer = trainersService.getTrainer(userId);
                model.addAttribute("trainer", trainer);
                model.addAttribute("trainerId", userId);
                break;
            case "staff":
                Staff staff = staffService.getStaff(userId);
                model.addAttribute("staff", staff);
                model.addAttribute("staffId", userId);
                break;
            default:
                break;
        }
        model.addAttribute("role", role);
        model.addAttribute("trainers", trainersService.getAllTrainers());
        model.addAttribute("trainingRequest", new TrainingRequest());

        return "trainers";
    }

    @PostMapping("/trainers/subscribe")
    @Transactional
    public String subscribe(@ModelAttribute TrainingRequest form, Model model, Principal principal) {
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
    public String getTrainings(@RequestParam(value = "id_trainer", required = false) Set<Integer> trainerId,
            @RequestParam(value = "id_training_type", required = false) Set<Integer> trainingTypeId,
            @RequestParam(value = "session_date_start", required = false) String sessionDateStart,
            @RequestParam(value = "session_date_end", required = false) String sessionDateEnd,
            @RequestParam(value = "session_time_start", required = false) Integer sessionTimeStart,
            @RequestParam(value = "session_time_end", required = false) Integer sessionTimeEnd,
            @RequestParam(value = "trainer_schedule", required = false, defaultValue = "0") Integer trainerSchedule) {

        // Нормализация параметров
        if (trainerId != null && trainerId.isEmpty())
            trainerId = null;
        if (trainingTypeId != null && trainingTypeId.isEmpty())
            trainingTypeId = null;
        if (sessionTimeStart != null && sessionTimeStart <= 0)
            sessionTimeStart = null;
        if (sessionTimeEnd != null && sessionTimeEnd <= 0)
            sessionTimeEnd = null;

        LocalDateTime startDate = null;
        LocalDateTime endDate = null;

        try {
            if (sessionDateStart != null && !sessionDateStart.trim().isEmpty()) {
                startDate = LocalDate.parse(sessionDateStart).atStartOfDay();
            }
            if (sessionDateEnd != null && !sessionDateEnd.trim().isEmpty()) {
                endDate = LocalDate.parse(sessionDateEnd).atTime(23, 59, 59);
            }
        } catch (Exception e) {
            // Игнорируем ошибки парсинга дат
        }

        try {
            Set<TrainingSchedule> trainingScheduleSet = trainingScheduleService.getTrainingSet(
                    trainerId, trainingTypeId, startDate, endDate, sessionTimeStart, sessionTimeEnd);

            Set<Event> eventsSet = trainingScheduleService.trainingScheduleToEventSet(
                    trainingScheduleSet, trainerSchedule != null ? trainerSchedule : 0);

            ObjectMapper mapper = new ObjectMapper();
            return mapper.writerWithDefaultPrettyPrinter().writeValueAsString(eventsSet);
        } catch (Exception e) {
            e.printStackTrace();
            return "[]";
        }
    }

    @GetMapping("/calendar/work/events")
    @ResponseBody
    public String getWorks() {
        List<StaffSchedule> staffScheduleSet = staffScheduleService.getAllStaffSchedules();
        List<Event> eventsSet = staffScheduleService.staffScheduleToEvents(staffScheduleSet);

        ObjectMapper mapper = new ObjectMapper();
        try {
            return mapper.writerWithDefaultPrettyPrinter().writeValueAsString(eventsSet);
        } catch (IOException ioex) {
            return "[]";
        }
    }

    @GetMapping("/access-denied")
    public String accessDenied() {
        return "access-denied";
    }

    @GetMapping("/not-found")
    public String notFound() {
        return "not-found";
    }

    // === ДИАГНОСТИЧЕСКИЕ МЕТОДЫ ===

    @GetMapping("/direct-diagnose")
    @ResponseBody
    public String directDiagnose(@RequestParam String username) {
        StringBuilder result = new StringBuilder();
        result.append("=== DIRECT DIAGNOSIS: ").append(username).append(" ===\n\n");

        result.append("1. TrainersAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<TrainersAccounts> trainer = trainersAccountsRepo.findById(username);
            if (trainer.isPresent()) {
                result.append("   ✅ FOUND: ").append(trainer.get().getUsername()).append("\n");
                result.append("   Password: ").append(trainer.get().getPassword()).append("\n");
                result.append("   Role: ").append(trainer.get().getUserRole()).append("\n");
                result.append("   Trainer ID: ").append(trainer.get().getTrainer().getIdTrainer()).append("\n");
            } else {
                result.append("   ❌ NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("   💥 ERROR: ").append(e.getMessage()).append("\n");
        }

        result.append("\n2. MembersAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<MembersAccounts> member = membersAccountsRepo.findById(username);
            if (member.isPresent()) {
                result.append("   ✅ FOUND\n");
            } else {
                result.append("   ❌ NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("   💥 ERROR: ").append(e.getMessage()).append("\n");
        }

        result.append("\n3. StaffAccountsRepository.findById('").append(username).append("'):\n");
        try {
            Optional<StaffAccounts> staff = staffAccountsRepo.findById(username);
            if (staff.isPresent()) {
                result.append("   ✅ FOUND\n");
            } else {
                result.append("   ❌ NOT FOUND\n");
            }
        } catch (Exception e) {
            result.append("   💥 ERROR: ").append(e.getMessage()).append("\n");
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
            result.append("   💥 ERROR: ").append(e.getMessage()).append("\n");
        }

        return result.toString();
    }

    @GetMapping("/diagnose")
    @ResponseBody
    public String diagnose(@RequestParam String username) {
        return accountService.diagnoseUser(username);
    }

    @GetMapping("/check-password")
    @ResponseBody
    public String checkPassword(@RequestParam String username, @RequestParam String password) {
        boolean result = accountService.checkPassword(username, password);
        return String.format("Password check for %s: %s", username, result ? "✅ VALID" : "❌ INVALID");
    }
}
