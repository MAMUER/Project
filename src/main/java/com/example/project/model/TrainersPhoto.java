package com.example.project.model;

import lombok.*;

import java.util.HashSet;
import java.util.Set;

import com.example.project.model.accounts.TrainersAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "trainers_photo")
public class TrainersPhoto {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_trainers_photo", nullable = false)
    private int idTrainersPhoto;

    @Column(name = "image_url", nullable = false, length = 250)
    private String imageUrl;

    @OneToMany(mappedBy = "trainersPhoto", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<TrainersAccounts> trainersAccounts = new HashSet<>();
}