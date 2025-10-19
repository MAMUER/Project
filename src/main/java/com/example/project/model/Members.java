package com.example.project.model;

import lombok.*;

import java.time.LocalDate;
import java.util.Collections;
import java.util.HashSet;
import java.util.Set;

import com.example.project.model.Accounts.MembersAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "members")
public class Members {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_member", nullable = false)
    private Integer idMember;

    @ManyToOne
    @JoinColumn(name = "id_role", nullable = false)
    private MembershipRole membershipRole;

    @ManyToOne
    @JoinColumn(name = "club_name")
    private Clubs club;

    @Column(name = "first_name")
    private String firstName;

    @Column(name = "second_name")
    private String secondName;

    @Column(name = "phone_number")
    private String phoneNumber;

    @Column(name = "email")
    private String email;

    @Column(name = "birth_date")
    private LocalDate birthDate;

    @Column(name = "start_trial_date")
    private LocalDate startTrialDate;

    @Column(name = "end_trial_date")
    private LocalDate endTrialDate;

    @Column(name = "gender")
    private Integer gender;

    @OneToOne(mappedBy = "member", cascade = CascadeType.ALL, orphanRemoval = true)
    private MembersAccounts membersAccount;

    @OneToMany(mappedBy = "member", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<MembersHaveAchievements> membersHaveAchievements = new HashSet<>();

    @ManyToMany
    @JoinTable(name = "members_have_equipment_statistics", joinColumns = @JoinColumn(name = "id_member"), inverseJoinColumns = @JoinColumn(name = "id_statistics"))
    private Set<EquipmentStatistics> equipmentStatistics = new HashSet<>();

    @ManyToMany
    @JoinTable(name = "members_have_inbody_analyses", joinColumns = @JoinColumn(name = "id_member"), inverseJoinColumns = @JoinColumn(name = "id_inbody_analys"))
    private Set<InbodyAnalyses> inbodyAnalyses = new HashSet<>();

    @ManyToMany(fetch = FetchType.LAZY)
    @JoinTable(name = "members_have_training_schedule", joinColumns = @JoinColumn(name = "id_member"), inverseJoinColumns = @JoinColumn(name = "id_session"))
    private Set<TrainingSchedule> trainingSchedules = new HashSet<>();

    @ManyToMany
    @JoinTable(name = "members_have_visits_history", joinColumns = @JoinColumn(name = "id_member"), inverseJoinColumns = @JoinColumn(name = "id_visit"))
    private Set<VisitsHistory> visitsHistory = new HashSet<>();

    @OneToMany(mappedBy = "member", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<NutritionPlan> nutritionPlans = new HashSet<>();

    public Set<Feedback> getFeedbacks() {
        return this.membersAccount != null ? this.membersAccount.getFeedbacks() : Collections.emptySet();
    }

    public void addTraining(TrainingSchedule training) {
        this.trainingSchedules.add(training);
    }
}