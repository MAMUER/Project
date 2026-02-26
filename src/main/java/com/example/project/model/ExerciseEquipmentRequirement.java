package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "exercise_equipment_requirements")
public class ExerciseEquipmentRequirement {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_requirement")
    private Integer idRequirement;

    @ManyToOne
    @JoinColumn(name = "id_exercise", nullable = false)
    private Exercise exercise;

    @ManyToOne
    @JoinColumn(name = "id_equipment_type", nullable = false)
    private EquipmentType equipmentType;

    @Column(name = "quantity_required")
    private Integer quantityRequired;

    @Column(name = "is_required")
    private Boolean isRequired;
}
