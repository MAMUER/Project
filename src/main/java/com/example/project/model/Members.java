package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;
import java.time.LocalDate;
import java.util.HashSet;
import java.util.Set;

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

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "club_name")
    private Clubs club;

    @Column(name = "first_name", nullable = false, length = 45)
    private String firstName;

    @Column(name = "second_name", nullable = false, length = 45)
    private String secondName;

    @Column(name = "birth_date", nullable = false)
    private LocalDate birthDate;

    @Column(name = "gender", nullable = false)
    private Integer gender;

    @OneToOne(mappedBy = "member", cascade = CascadeType.ALL, orphanRemoval = true)
    private com.example.project.model.Accounts.MembersAccounts membersAccount;

    @OneToMany(mappedBy = "member", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<MembersHaveAchievements> membersHaveAchievements = new HashSet<>();

    @ManyToMany(mappedBy = "members")
    private Set<InbodyAnalysis> inbodyAnalysis = new HashSet<>();

    @ManyToMany(fetch = FetchType.LAZY)
    @JoinTable(name = "members_have_training_schedule", 
               joinColumns = @JoinColumn(name = "id_member"), 
               inverseJoinColumns = @JoinColumn(name = "id_session"))
    private Set<TrainingSchedule> trainingSchedules = new HashSet<>();

    @ManyToMany
    @JoinTable(name = "members_have_visits_history", 
               joinColumns = @JoinColumn(name = "id_member"), 
               inverseJoinColumns = @JoinColumn(name = "id_visit"))
    private Set<VisitsHistory> visitsHistory = new HashSet<>();
}