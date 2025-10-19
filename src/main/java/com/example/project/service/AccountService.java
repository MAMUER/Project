package com.example.project.service;

import java.time.LocalDate;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.Clubs;
import com.example.project.model.Members;
import com.example.project.model.MembershipRole;
import com.example.project.model.UsersPhoto;
import com.example.project.model.Accounts.MembersAccounts;
import com.example.project.model.Accounts.StaffAccounts;
import com.example.project.model.Accounts.TrainersAccounts;
import com.example.project.repository.ClubsRepository;
import com.example.project.repository.MembersAccountsRepository;
import com.example.project.repository.MembersRepository;
import com.example.project.repository.MembershipRoleRepository;
import com.example.project.repository.StaffAccountsRepository;
import com.example.project.repository.TrainersAccountsRepository;
import com.example.project.repository.UsersPhotoRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AccountService {
    private final MembersAccountsRepository membersAccountsRepo;
    private final TrainersAccountsRepository trainersAccountsRepo;
    private final StaffAccountsRepository staffAccountsRepo;
    private final MembersRepository membersRepository;
    private final PasswordEncoder passwordEncoder;
    private final MembersAccountsService membersAccountsService;
    private final TrainersAccountsService trainersAccountsService;
    private final StaffAccountsService staffAccountsService;
    private final MembershipRoleRepository membershipRoleRepository;
    private final UsersPhotoRepository usersPhotoRepository;
    private final ClubsRepository clubsRepository;

    private static final Logger logger = LoggerFactory.getLogger(AccountService.class);

    public void updateLastLogin(String username) {
        AccountInfo accountInfo = getAccountInfo(username);
        if (accountInfo != null) {
            LocalDate currentDate = LocalDate.now();

            switch (accountInfo.role().toLowerCase()) {
                case "member":
                    MembersAccounts memberAccount = membersAccountsService.getMemberAccount(username);
                    if (memberAccount != null) {
                        memberAccount.setLastLogin(currentDate);
                        membersAccountsService.saveMemberAccount(memberAccount);
                    }
                    break;
                case "trainer":
                    TrainersAccounts trainerAccount = trainersAccountsService.getTrainerAccount(username);
                    if (trainerAccount != null) {
                        trainerAccount.setLastLogin(currentDate);
                        trainersAccountsService.saveTrainerAccount(trainerAccount);
                    }
                    break;
                case "staff":
                    StaffAccounts staffAccount = staffAccountsService.getStaffAccount(username);
                    if (staffAccount != null) {
                        staffAccount.setLastLogin(currentDate);
                        staffAccountsService.saveStaffAccount(staffAccount);
                    }
                    break;
            }
        }
    }

    public String getPasswordByUsername(String username) {
        logger.info("🔍 Searching password for user: {}", username);

        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            logger.info("✅ Found in members_accounts: {}", username);
            return member.getPassword();
        }

        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            logger.info("✅ Found in trainers_accounts: {}", username);
            return trainer.getPassword();
        }

        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            logger.info("✅ Found in staff_accounts: {}", username);
            return staff.getPassword();
        }

        logger.error("❌ User NOT found in any account table: {}", username);
        return null;
    }

    public Integer getIdByUsername(String username) {
        logger.info("🔍 Searching ID for user: {}", username);

        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            logger.info("✅ Found member ID: {}", member.getMember().getIdMember());
            return member.getMember().getIdMember();
        }

        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            logger.info("✅ Found trainer ID: {}", trainer.getTrainer().getIdTrainer());
            return trainer.getTrainer().getIdTrainer();
        }

        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            logger.info("✅ Found staff ID: {}", staff.getStaff().getIdStaff());
            return staff.getStaff().getIdStaff();
        }

        logger.error("❌ User ID NOT found: {}", username);
        return null;
    }

    public String getRoleByUsername(String username) {
        logger.info("🔍 Searching role for user: {}", username);

        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            logger.info("✅ Found member role: {}", member.getUserRole());
            return member.getUserRole();
        }

        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            logger.info("✅ Found trainer role: {}", trainer.getUserRole());
            return trainer.getUserRole();
        }

        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            logger.info("✅ Found staff role: {}", staff.getUserRole());
            return staff.getUserRole();
        }

        logger.error("❌ User role NOT found: {}", username);
        return null;
    }

    public String getUsernameById(Integer userId) {
        logger.info("🔍 Searching username for ID: {}", userId);

        // Поиск по members
        MembersAccounts member = membersAccountsRepo.findByMemberIdMember(userId).orElse(null);
        if (member != null) {
            logger.info("✅ Found member username: {}", member.getUsername());
            return member.getUsername();
        }

        // Поиск по trainers
        TrainersAccounts trainer = trainersAccountsRepo.findByTrainerIdTrainer(userId).orElse(null);
        if (trainer != null) {
            logger.info("✅ Found trainer username: {}", trainer.getUsername());
            return trainer.getUsername();
        }

        // Поиск по staff
        StaffAccounts staff = staffAccountsRepo.findByStaffIdStaff(userId).orElse(null);
        if (staff != null) {
            logger.info("✅ Found staff username: {}", staff.getUsername());
            return staff.getUsername();
        }

        logger.error("❌ Username NOT found for ID: {}", userId);
        return null;
    }

    public AccountInfo getAccountInfo(String username) {
        logger.info("🎯 Getting account info for: {}", username);

        String password = getPasswordByUsername(username);
        if (password == null) {
            logger.error("❌ Account info NOT found - password is null for: {}", username);
            return null;
        }

        Integer id = getIdByUsername(username);
        String role = getRoleByUsername(username);

        AccountInfo accountInfo = new AccountInfo(username, password, id, role);
        logger.info("✅ Account info found: {}", accountInfo);

        return accountInfo;
    }

    public AccountInfo getAccountInfoById(Integer userId) {
        logger.info("🎯 Getting account info by ID: {}", userId);

        String username = getUsernameById(userId);
        if (username == null) {
            logger.error("❌ Account info NOT found for ID: {}", userId);
            return null;
        }

        return getAccountInfo(username);
    }

    // DTO для передачи информации об аккаунте
    public record AccountInfo(String username, String password, Integer id, String role) {
        @Override
        public String toString() {
            return String.format("AccountInfo[username=%s, password=%s, id=%s, role=%s]",
                    username, password, id, role);
        }
    }

    // Дополнительные методы для проверки существования аккаунтов
    public boolean isMemberAccount(String username) {
        boolean exists = membersAccountsRepo.findById(username).isPresent();
        logger.info("👤 Member account check for {}: {}", username, exists);
        return exists;
    }

    public boolean isTrainerAccount(String username) {
        boolean exists = trainersAccountsRepo.findById(username).isPresent();
        logger.info("🏋️ Trainer account check for {}: {}", username, exists);
        return exists;
    }

    public boolean isStaffAccount(String username) {
        boolean exists = staffAccountsRepo.findById(username).isPresent();
        logger.info("👔 Staff account check for {}: {}", username, exists);
        return exists;
    }

    public String getAccountType(String username) {
        if (isMemberAccount(username)) {
            logger.info("📝 Account type for {}: MEMBER", username);
            return "MEMBER";
        }
        if (isTrainerAccount(username)) {
            logger.info("📝 Account type for {}: TRAINER", username);
            return "TRAINER";
        }
        if (isStaffAccount(username)) {
            logger.info("📝 Account type for {}: STAFF", username);
            return "STAFF";
        }
        logger.error("📝 Account type NOT found for: {}", username);
        return null;
    }

    /**
     * Проверка пароля для отладки
     */
    public boolean checkPassword(String username, String rawPassword) {
        logger.info("🔐 Checking password for: {}", username);

        String encodedPassword = getPasswordByUsername(username);
        if (encodedPassword == null) {
            logger.error("❌ Cannot check password - user not found: {}", username);
            return false;
        }

        boolean matches = passwordEncoder.matches(rawPassword, encodedPassword);
        logger.info("🔐 Password check result for {}: {}", username, matches);

        return matches;
    }

    /**
     * Полная диагностика пользователя
     */
    public String diagnoseUser(String username) {
        StringBuilder result = new StringBuilder();
        result.append("=== USER DIAGNOSIS: ").append(username).append(" ===\n");

        // Проверка в каждой таблице
        result.append("\n--- Direct Repository Checks ---\n");

        boolean inMembers = membersAccountsRepo.findById(username).isPresent();
        result.append("Members table: ").append(inMembers ? "✅ FOUND" : "❌ NOT FOUND").append("\n");

        boolean inTrainers = trainersAccountsRepo.findById(username).isPresent();
        result.append("Trainers table: ").append(inTrainers ? "✅ FOUND" : "❌ NOT FOUND").append("\n");

        boolean inStaff = staffAccountsRepo.findById(username).isPresent();
        result.append("Staff table: ").append(inStaff ? "✅ FOUND" : "❌ NOT FOUND").append("\n");

        // Информация через AccountService
        result.append("\n--- AccountService Results ---\n");
        AccountInfo accountInfo = getAccountInfo(username);
        if (accountInfo != null) {
            result.append("AccountInfo: ").append(accountInfo).append("\n");
        } else {
            result.append("AccountInfo: ❌ NULL\n");
        }

        return result.toString();
    }

    private MembershipRole getDefaultMembershipRole() {
        // Попробуем найти роль по ID = 1 (обычно это базовая роль)
        MembershipRole defaultRole = membershipRoleRepository.findById(1)
                .orElse(null);

        if (defaultRole == null) {
            // Если роль с ID=1 не найдена, возьмем первую доступную
            defaultRole = membershipRoleRepository.findAll().stream()
                    .findFirst()
                    .orElseThrow(() -> new RuntimeException("No membership roles found in database"));
        }

        logger.info("🎯 Using default membership role: {}", defaultRole.getRoleName());
        return defaultRole;
    }

    @Transactional
    public boolean registerMember(String username, String password, String email, String firstName,
            String lastName, String phoneNumber, LocalDate birthDate,
            String clubName, Integer gender, Integer membershipPeriod) {
        try {
            logger.info("📝 Starting registration for: {}", username);

            // Проверка существования username
            if (getAccountInfo(username) != null) {
                logger.error("❌ Registration failed - user already exists: {}", username);
                return false;
            }

            // Получаем роль по умолчанию
            MembershipRole defaultRole = getDefaultMembershipRole();

            // Получаем клуб
            Clubs club = getClubByName(clubName);
            if (club == null) {
                logger.error("❌ Club not found: {}", clubName);
                return false;
            }

            // Получаем фото по умолчанию
            UsersPhoto defaultPhoto = getDefaultPhoto();

            // Рассчитываем дату окончания trial периода
            LocalDate endTrialDate = calculateEndTrialDate(membershipPeriod);

            // Создание нового члена
            Members member = new Members();
            member.setFirstName(firstName);
            member.setSecondName(lastName);
            member.setEmail(email);
            member.setPhoneNumber(phoneNumber);
            member.setBirthDate(birthDate);
            member.setStartTrialDate(LocalDate.now());
            member.setEndTrialDate(endTrialDate);
            member.setGender(gender);
            member.setMembershipRole(defaultRole);
            member.setClub(club);

            Members savedMember = membersRepository.save(member);
            logger.info("✅ Member created with ID: {}", savedMember.getIdMember());

            // Создание аккаунта
            MembersAccounts account = new MembersAccounts();
            account.setUsername(username);
            account.setPassword(passwordEncoder.encode(password));
            account.setMember(savedMember);
            account.setUserRole("MEMBER");
            account.setAccountCreationDate(LocalDate.now());
            account.setLastLogin(LocalDate.now());
            account.setUserPhoto(defaultPhoto);

            membersAccountsRepo.save(account);
            logger.info("✅ Account created successfully for: {}", username);

            return true;

        } catch (Exception e) {
            logger.error("❌ Registration failed for {}: {}", username, e.getMessage(), e);
            return false;
        }
    }

    /**
     * Получить клуб по имени
     */
    private Clubs getClubByName(String clubName) {
        return clubsRepository.findByClubName(clubName).orElse(null);
    }

    /**
     * Рассчитать дату окончания trial периода
     */
    private LocalDate calculateEndTrialDate(Integer membershipPeriod) {
        return LocalDate.now().plusMonths(membershipPeriod);
    }

    private UsersPhoto getDefaultPhoto() {
        return usersPhotoRepository.findById(1)
                .orElseThrow(() -> new RuntimeException("Default photo not found"));
    }
}