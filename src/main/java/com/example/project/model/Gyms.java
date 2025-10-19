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
@Table(name = "gyms")
public class Gyms {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_gym", nullable = false)
    private int idGym;

    @ManyToOne
    @JoinColumn(name = "club_name", nullable = false)
    private Clubs club;

    @Column(name = "gym_name", nullable = false, length = 45)
    private String gymName;

    @Column(name = "capacity", nullable = false)
    private int capacity;

    @Column(name = "available_hours", nullable = false)
    private int availableHours;

    @ManyToMany
    @JoinTable(
        name = "gyms_have_equipment",
        joinColumns = @JoinColumn(name = "id_gym"),
        inverseJoinColumns = @JoinColumn(name = "id_equipment")
    )
    private Set<Equipment> equipment = new HashSet<>();
}