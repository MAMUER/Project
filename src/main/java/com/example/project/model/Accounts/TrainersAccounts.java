package com.example.project.model.Accounts;

import lombok.*;

import java.time.LocalDate;

import com.example.project.model.Trainers;
import com.example.project.model.TrainersPhoto;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "trainers_accounts")
public class TrainersAccounts {
    @Id
    @Column(name = "username", nullable = false, length = 45)
    private String username;

    @OneToOne
    @JoinColumn(name = "id_trainer", nullable = false)
    private Trainers trainer;

    @ManyToOne
    @JoinColumn(name = "id_trainers_photo")
    private TrainersPhoto trainersPhoto;

    @Column(name = "password", nullable = false, length = 100)
    private String password;

    @Column(name = "last_login")
    private LocalDate lastLogin;

    @Column(name = "account_creation_date", nullable = false)
    private LocalDate accountCreationDate;

    @Column(name = "user_role", nullable = false, length = 45)
    private String userRole;
}