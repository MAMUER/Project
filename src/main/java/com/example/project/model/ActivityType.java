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
@Table(name = "activity_type")
public class ActivityType {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_activity", nullable = false)
    private int idActivity;

    @Column(name = "activity_name", nullable = false)
    private String activityName;

    @OneToMany(mappedBy = "activityType", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<EquipmentStatistics> equipmentStatistics = new HashSet<>();
}