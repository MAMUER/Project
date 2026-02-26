package com.example.project.model;

import lombok.*;

import java.util.HashSet;
import java.util.Set;

import com.example.project.model.accounts.MembersAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "users_photo")
public class UsersPhoto {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_photo", nullable = false)
    private int idPhoto;

    @Column(name = "image_url", nullable = false, length = 200)
    private String imageUrl;

    @OneToMany(mappedBy = "userPhoto", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<MembersAccounts> membersAccounts = new HashSet<>();
}