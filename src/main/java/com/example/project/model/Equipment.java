package com.example.project.model;

import lombok.*;
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
    private Integer idEquipment;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "id_equipment_type")
    private EquipmentType equipmentType;

    @Column(name = "quantity")
    private Integer quantity;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "club_name")
    private Clubs club;
}