package com.example.project.model;

import lombok.*;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

import com.example.project.model.Accounts.TrainersAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "trainers")
public class Trainers {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_trainer", nullable = false)
    private int idTrainer;

    @Column(name = "first_name", nullable = false, length = 45)
    private String firstName;

    @Column(name = "second_name", nullable = false, length = 45)
    private String secondName;

    @Column(name = "speciality", length = 45)
    private String speciality;

    @Column(name = "experience")
    private Integer experience;

    @Column(name = "certifications")
    private Integer certifications;

    @Column(name = "phone_number", nullable = false, length = 11)
    private String phoneNumber;

    @Column(name = "email", nullable = false, length = 45)
    private String email;

    @Column(name = "hire_date", nullable = false)
    private LocalDate hireDate;

    @OneToOne(mappedBy = "trainer", cascade = CascadeType.ALL, orphanRemoval = true)
    private TrainersAccounts trainersAccount;

    @OneToMany(mappedBy = "trainer", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<TrainingSchedule> trainingSchedules = new ArrayList<>();
}