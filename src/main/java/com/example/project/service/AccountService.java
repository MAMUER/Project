package com.example.project.service;

import java.time.LocalDate;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.lang.Nullable;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.Clubs;
import com.example.project.model.Members;
import com.example.project.model.UsersPhoto;
import com.example.project.model.Accounts.MembersAccounts;
import com.example.project.model.Accounts.StaffAccounts;
import com.example.project.model.Accounts.TrainersAccounts;
import com.example.project.repository.ClubsRepository;
import com.example.project.repository.MembersAccountsRepository;
import com.example.project.repository.MembersRepository;
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
    private final UsersPhotoRepository usersPhotoRepository;
    private final ClubsRepository clubsRepository;

    private static final Logger logger = LoggerFactory.getLogger(AccountService.class);

    public void updateLastLogin(String username) {
        AccountInfo accountInfo = getAccountInfo(username);
        if (accountInfo != null) {
            LocalDate currentDate = LocalDate.now();

            switch (accountInfo.role().toLowerCase()) {
                case "member" -> {
                    MembersAccounts memberAccount = membersAccountsService.getMemberAccount(username);
                    if (memberAccount != null) {
                        memberAccount.setLastLogin(currentDate);
                        membersAccountsService.saveMemberAccount(memberAccount);
                    }
                }
                case "trainer" -> {
                    TrainersAccounts trainerAccount = trainersAccountsService.getTrainerAccount(username);
                    if (trainerAccount != null) {
                        trainerAccount.setLastLogin(currentDate);
                        trainersAccountsService.saveTrainerAccount(trainerAccount);
                    }
                }
                case "staff" -> {
                    StaffAccounts staffAccount = staffAccountsService.getStaffAccount(username);
                    if (staffAccount != null) {
                        staffAccount.setLastLogin(currentDate);
                        staffAccountsService.saveStaffAccount(staffAccount);
                    }
                }
            }
        }
    }

    public String getPasswordByUsername(@Nullable String username) {
        // Явная проверка в начале метода
        if (username == null || username.trim().isEmpty()) {
            logger.error("Cannot get password - username is null or empty");
            return null;
        }

        // Теперь username точно не null, можно использовать
        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            return member.getPassword();
        }

        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            return trainer.getPassword();
        }

        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            return staff.getPassword();
        }

        logger.error(" User NOT found in any account table: {}", username);
        return null;
    }

    public Integer getIdByUsername(String username) {

        @SuppressWarnings("null")
        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            return member.getMember().getIdMember();
        }

        @SuppressWarnings("null")
        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            return trainer.getTrainer().getIdTrainer();
        }

        @SuppressWarnings("null")
        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            return staff.getStaff().getIdStaff();
        }

        logger.error(" User ID NOT found: {}", username);
        return null;
    }

    public String getRoleByUsername(String username) {

        @SuppressWarnings("null")
        MembersAccounts member = membersAccountsRepo.findById(username).orElse(null);
        if (member != null) {
            return member.getUserRole();
        }

        @SuppressWarnings("null")
        TrainersAccounts trainer = trainersAccountsRepo.findById(username).orElse(null);
        if (trainer != null) {
            return trainer.getUserRole();
        }

        @SuppressWarnings("null")
        StaffAccounts staff = staffAccountsRepo.findById(username).orElse(null);
        if (staff != null) {
            return staff.getUserRole();
        }

        logger.error(" User role NOT found: {}", username);
        return null;
    }

    public String getUsernameById(Integer userId) {

        // Поиск по members
        MembersAccounts member = membersAccountsRepo.findByMemberIdMember(userId).orElse(null);
        if (member != null) {
            return member.getUsername();
        }

        // Поиск по trainers
        TrainersAccounts trainer = trainersAccountsRepo.findByTrainerIdTrainer(userId).orElse(null);
        if (trainer != null) {
            return trainer.getUsername();
        }

        // Поиск по staff
        StaffAccounts staff = staffAccountsRepo.findByStaffIdStaff(userId).orElse(null);
        if (staff != null) {
            return staff.getUsername();
        }

        logger.error(" Username NOT found for ID: {}", userId);
        return null;
    }

    public AccountInfo getAccountInfo(String username) {

        String password = getPasswordByUsername(username);
        if (password == null) {
            logger.error(" Account info NOT found - password is null for: {}", username);
            return null;
        }

        Integer id = getIdByUsername(username);
        String role = getRoleByUsername(username);

        AccountInfo accountInfo = new AccountInfo(username, password, id, role);

        return accountInfo;
    }

    public AccountInfo getAccountInfoById(Integer userId) {

        String username = getUsernameById(userId);
        if (username == null) {
            logger.error(" Account info NOT found for ID: {}", userId);
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
        @SuppressWarnings("null")
        boolean exists = membersAccountsRepo.findById(username).isPresent();
        return exists;
    }

    public boolean isTrainerAccount(String username) {
        @SuppressWarnings("null")
        boolean exists = trainersAccountsRepo.findById(username).isPresent();
        return exists;
    }

    public boolean isStaffAccount(String username) {
        @SuppressWarnings("null")
        boolean exists = staffAccountsRepo.findById(username).isPresent();
        return exists;
    }

    public String getAccountType(String username) {
        if (isMemberAccount(username)) {
            return "MEMBER";
        }
        if (isTrainerAccount(username)) {
            return "TRAINER";
        }
        if (isStaffAccount(username)) {
            return "STAFF";
        }
        logger.error("📝 Account type NOT found for: {}", username);
        return null;
    }

    /**
     * Проверка пароля для отладки
     */
    public boolean checkPassword(String username, String rawPassword) {

        String encodedPassword = getPasswordByUsername(username);
        if (encodedPassword == null) {
            logger.error(" Cannot check password - user not found: {}", username);
            return false;
        }

        boolean matches = passwordEncoder.matches(rawPassword, encodedPassword);

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

        @SuppressWarnings("null")
        boolean inMembers = membersAccountsRepo.findById(username).isPresent();
        result.append("Members table: ").append(inMembers ? " FOUND" : " NOT FOUND").append("\n");

        @SuppressWarnings("null")
        boolean inTrainers = trainersAccountsRepo.findById(username).isPresent();
        result.append("Trainers table: ").append(inTrainers ? " FOUND" : " NOT FOUND").append("\n");

        @SuppressWarnings("null")
        boolean inStaff = staffAccountsRepo.findById(username).isPresent();
        result.append("Staff table: ").append(inStaff ? " FOUND" : " NOT FOUND").append("\n");

        // Информация через AccountService
        result.append("\n--- AccountService Results ---\n");
        AccountInfo accountInfo = getAccountInfo(username);
        if (accountInfo != null) {
            result.append("AccountInfo: ").append(accountInfo).append("\n");
        } else {
            result.append("AccountInfo:  NULL\n");
        }

        return result.toString();
    }

    // В методе registerMember исправить создание Members:
    @Transactional
    public boolean registerMember(String username, String password, String firstName,
            String lastName, LocalDate birthDate,
            String clubName, Integer gender) {
        try {

            // Проверка существования username
            if (getAccountInfo(username) != null) {
                logger.error(" Registration failed - user already exists: {}", username);
                return false;
            }

            // Получаем клуб
            Clubs club = getClubByName(clubName);
            if (club == null) {
                logger.error(" Club not found: {}", clubName);
                return false;
            }

            // Получаем фото по умолчанию
            UsersPhoto defaultPhoto = getDefaultPhoto();

            // Создание нового члена
            Members member = new Members();
            member.setFirstName(firstName);
            member.setSecondName(lastName);
            member.setBirthDate(birthDate);
            member.setGender(gender);
            member.setClub(club);

            Members savedMember = membersRepository.save(member);

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

            return true;

        } catch (Exception e) {
            logger.error(" Registration failed for {}: {}", username, e.getMessage(), e);
            return false;
        }
    }

    /**
     * Получить клуб по имени
     */
    private Clubs getClubByName(String clubName) {
        return clubsRepository.findByClubName(clubName).orElse(null);
    }

    private UsersPhoto getDefaultPhoto() {
        return usersPhotoRepository.findById(1)
                .orElseThrow(() -> new RuntimeException("Default photo not found"));
    }
}
