package com.example.project.model.accounts;

import lombok.*;

import java.time.LocalDate;
import java.util.HashSet;
import java.util.Set;

import com.example.project.model.Feedback;
import com.example.project.model.Members;
import com.example.project.model.UsersPhoto;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "members_accounts")
public class MembersAccounts {
    @Id
    @Column(name = "username", nullable = false, length = 45)
    private String username;

    @OneToOne
    @JoinColumn(name = "id_member", nullable = false)
    private Members member;

    @ManyToOne
    @JoinColumn(name = "id_photo")
    private UsersPhoto userPhoto;

    @Column(name = "password", nullable = false, length = 100)
    private String password;

    @Column(name = "account_creation_date", nullable = false)
    private LocalDate accountCreationDate;

    @Column(name = "last_login")
    private LocalDate lastLogin;

    @Column(name = "user_role", nullable = false, length = 45)
    private String userRole;

    @OneToMany(mappedBy = "memberAccount", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Feedback> feedbacks = new HashSet<>();
}