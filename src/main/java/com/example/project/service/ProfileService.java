package com.example.project.service;

import com.example.project.model.Members;
import com.example.project.model.Staff;
import com.example.project.model.Trainers;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.StaffSchedule;
import com.example.project.model.Achievements;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class ProfileService {

    private final MembersService membersService;
    private final TrainersService trainersService;
    private final StaffService staffService;
    private final TrainingScheduleService trainingScheduleService;
    private final NewsService newsService;

    @Transactional(readOnly = true)
    public ProfileData getMemberProfile(Integer id) {
        Members member = membersService.getMember(id);
        if (member == null) {
            return null;
        }

        ProfileData data = new ProfileData();
        data.setMemberId(id);
        data.setMemberClub(member.getClub());
        data.setMember(member);
        data.setFeedbacks(member.getMembersAccount() != null ? 
            member.getMembersAccount().getFeedbacks() : Collections.emptyList());
        data.setAchievements(membersService.getSetOfMemberAchievements(id));

        LocalDateTime now = LocalDateTime.now();
        Set<TrainingSchedule> memberTrainings = trainingScheduleService.getTrainingsByMemberId(id);
        Set<TrainingSchedule> upcomingWorkouts = memberTrainings.stream()
                .filter(workout -> workout.getSessionDate().isAfter(now))
                .limit(3)
                .collect(Collectors.toSet());

        data.setWorkouts(upcomingWorkouts);
        data.setWorkoutsCount(memberTrainings.stream()
                .filter(workout -> workout.getSessionDate().isAfter(now))
                .count());
        data.setPhotoURL(membersService.getPhotoUrl(id));
        data.setAllNews(newsService.getAllNews());

        return data;
    }

    @Transactional(readOnly = true)
    public ProfileData getTrainerProfile(Integer id) {
        Trainers trainer = trainersService.getTrainer(id);
        if (trainer == null) {
            return null;
        }

        ProfileData data = new ProfileData();
        data.setTrainerId(id);
        data.setTrainer(trainer);

        LocalDateTime nowTrainer = LocalDateTime.now();
        List<TrainingSchedule> trainerWorkouts = trainersService.getSetOfTrainingSchedule(id);
        Set<TrainingSchedule> upcomingTrainerWorkouts = trainerWorkouts.stream()
                .filter(workout -> workout.getSessionDate().isAfter(nowTrainer))
                .limit(3)
                .collect(Collectors.toSet());

        data.setWorkouts(upcomingTrainerWorkouts);
        data.setWorkoutsCount(trainerWorkouts.stream()
                .filter(workout -> workout.getSessionDate().isAfter(nowTrainer))
                .count());
        data.setPhotoURL(trainersService.getPhotoUrl(id));

        return data;
    }

    @Transactional(readOnly = true)
    public ProfileData getStaffProfile(Integer id) {
        Staff staff = staffService.getStaff(id);
        if (staff == null) {
            return null;
        }

        ProfileData data = new ProfileData();
        data.setStaffId(id);
        data.setStaff(staff);
        data.setPhotoURL(staffService.getPhotoUrl(id));

        LocalDate nowStaff = LocalDate.now();
        List<StaffSchedule> staffSchedules = staffService.getSetOfStaffSchedule(id);
        Set<StaffSchedule> upcomingStaffSchedules = staffSchedules.stream()
                .filter(work -> work.getDate().isAfter(nowStaff))
                .limit(3)
                .collect(Collectors.toSet());

        data.setStaffSchedule(upcomingStaffSchedules);

        return data;
    }

    // DTO класс для передачи данных профиля
    public static class ProfileData {
        private Integer memberId;
        private Object memberClub;
        private Members member;
        private Collection<?> feedbacks;
        private Set<Achievements> achievements;
        private Set<TrainingSchedule> workouts;
        private Long workoutsCount;
        private String photoURL;
        private List<?> allNews;
        
        private Integer trainerId;
        private Trainers trainer;
        
        private Integer staffId;
        private Staff staff;
        private Set<StaffSchedule> staffSchedule;

        // Геттеры и сеттеры
        public Integer getMemberId() { return memberId; }
        public void setMemberId(Integer memberId) { this.memberId = memberId; }
        
        public Object getMemberClub() { return memberClub; }
        public void setMemberClub(Object memberClub) { this.memberClub = memberClub; }
        
        public Members getMember() { return member; }
        public void setMember(Members member) { this.member = member; }
        
        public Collection<?> getFeedbacks() { return feedbacks; }
        public void setFeedbacks(Collection<?> feedbacks) { this.feedbacks = feedbacks; }
        
        public Set<Achievements> getAchievements() { return achievements; }
        public void setAchievements(Set<Achievements> achievements) { this.achievements = achievements; }
        
        public Set<TrainingSchedule> getWorkouts() { return workouts; }
        public void setWorkouts(Set<TrainingSchedule> workouts) { this.workouts = workouts; }
        
        public Long getWorkoutsCount() { return workoutsCount; }
        public void setWorkoutsCount(Long workoutsCount) { this.workoutsCount = workoutsCount; }
        
        public String getPhotoURL() { return photoURL; }
        public void setPhotoURL(String photoURL) { this.photoURL = photoURL; }
        
        public List<?> getAllNews() { return allNews; }
        public void setAllNews(List<?> allNews) { this.allNews = allNews; }
        
        public Integer getTrainerId() { return trainerId; }
        public void setTrainerId(Integer trainerId) { this.trainerId = trainerId; }
        
        public Trainers getTrainer() { return trainer; }
        public void setTrainer(Trainers trainer) { this.trainer = trainer; }
        
        public Integer getStaffId() { return staffId; }
        public void setStaffId(Integer staffId) { this.staffId = staffId; }
        
        public Staff getStaff() { return staff; }
        public void setStaff(Staff staff) { this.staff = staff; }
        
        public Set<StaffSchedule> getStaffSchedule() { return staffSchedule; }
        public void setStaffSchedule(Set<StaffSchedule> staffSchedule) { this.staffSchedule = staffSchedule; }
    }
}