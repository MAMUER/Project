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
@Table(name = "equipment_statistics")
public class EquipmentStatistics {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_statistics", nullable = false)
    private int idStatistics;

    @ManyToOne
    @JoinColumn(name = "id_activity", nullable = false)
    private ActivityType activityType;

    @Column(name = "approaches")
    private Integer approaches;

    @Column(name = "kilocalories")
    private Integer kilocalories;

    @ManyToMany(mappedBy = "equipmentStatistics")
    private Set<Members> members = new HashSet<>();
}