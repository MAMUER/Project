package com.example.project.model;

import lombok.*;
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
    private Integer idGym;

    @ManyToOne
    @JoinColumn(name = "club_name", nullable = false)
    private Clubs club;

    @ManyToOne
    @JoinColumn(name = "gym_type_id", nullable = false)
    private GymTypes gymType;

    @Column(name = "capacity", nullable = false)
    private Integer capacity;

    @Column(name = "available_hours", nullable = false)
    private Integer availableHours;
}