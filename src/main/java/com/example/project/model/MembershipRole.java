package com.example.project.model;

import lombok.*;

import java.util.HashSet;
import java.util.Set;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "membership_role")
public class MembershipRole {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_role", nullable = false)
    private int idRole;

    @Column(name = "role_name", nullable = false, length = 45)
    private String roleName;

    @OneToMany(mappedBy = "membershipRole", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Members> members = new HashSet<>();
}