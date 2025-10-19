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
@Table(name = "inbody_analyses")
public class InbodyAnalyses {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_inbody_analys", nullable = false)
    private int idInbodyAnalys;

    @Column(name = "height")
    private Float height;

    @Column(name = "weight")
    private Float weight;

    @Column(name = "bmi")
    private Float bmi;

    @Column(name = "fat_percent")
    private Float fatPercent;

    @Column(name = "muscle_persent")
    private Float musclePersent;

    @ManyToMany(mappedBy = "inbodyAnalyses")
    private Set<Members> members = new HashSet<>();
}