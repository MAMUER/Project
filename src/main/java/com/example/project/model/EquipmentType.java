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
@Table(name = "equipment_type")
public class EquipmentType {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_equipment_type", nullable = false)
    private int idEquipmentType;

    @Column(name = "type_name", nullable = false, length = 45)
    private String typeName;

    @OneToMany(mappedBy = "equipmentType", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Equipment> equipment = new HashSet<>();
}