package com.example.project.model;

import lombok.*;

import java.util.HashSet;
import java.util.Set;

import com.example.project.model.accounts.StaffAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "staff_photo")
public class StaffPhoto {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_staff_photo", nullable = false)
    private int idStaffPhoto;

    @Column(name = "image_url", nullable = false, length = 250)
    private String imageUrl;

    @OneToMany(mappedBy = "staffPhoto", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<StaffAccounts> staffAccounts = new HashSet<>();
}