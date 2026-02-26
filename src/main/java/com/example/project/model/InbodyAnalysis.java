package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;
import java.util.HashSet;
import java.util.Set;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "inbody_analysis") // ИСПРАВЛЕНО: правильное название таблицы
public class InbodyAnalysis {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_inbody_analysis", nullable = false)
    private Integer idInbodyAnalysis;

    @Column(name = "height")
    private Float height;

    @Column(name = "weight")
    private Float weight;

    @Column(name = "bmi")
    private Float bmi;

    @Column(name = "fat_percent")
    private Float fatPercent;

    @Column(name = "muscle_percent")
    private Float musclePercent;

    @ManyToMany
    @JoinTable(name = "members_have_inbody_analysis", // ИСПРАВЛЕНО: правильное название
               joinColumns = @JoinColumn(name = "id_inbody_analysis"), 
               inverseJoinColumns = @JoinColumn(name = "id_member"))
    private Set<Members> members = new HashSet<>();
}