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
@Table(name = "equipment")
public class Equipment {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_equipment", nullable = false)
    private int idEquipment;

    @ManyToOne
    @JoinColumn(name = "id_equipment_type")
    private EquipmentType equipmentType;

    @Column(name = "name", nullable = false, length = 45)
    private String name;

    @Column(name = "quantity")
    private Integer quantity;

    @ManyToMany(mappedBy = "equipment")
    private Set<Gyms> gyms = new HashSet<>();
}