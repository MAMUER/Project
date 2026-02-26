package com.example.project.model.accounts;

import jakarta.persistence.*;

import java.time.LocalDate;

import com.example.project.model.Staff;
import com.example.project.model.StaffPhoto;

import lombok.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "staff_accounts")
public class StaffAccounts {
    @Id
    @Column(name = "username", nullable = false, length = 45)
    private String username;

    @OneToOne
    @JoinColumn(name = "id_staff", nullable = false)
    private Staff staff;

    @ManyToOne
    @JoinColumn(name = "id_staff_photo")
    private StaffPhoto staffPhoto;

    @Column(name = "password", nullable = false, length = 100)
    private String password;

    @Column(name = "last_login")
    private LocalDate lastLogin;

    @Column(name = "account_creation_date", nullable = false)
    private LocalDate accountCreationDate;

    @Column(name = "user_role", nullable = false, length = 45)
    private String userRole;
}