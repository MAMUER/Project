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
@Table(name = "gym_types")
public class GymTypes {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "gym_type_id", nullable = false)
    private Integer gymTypeId;

    @Column(name = "type_name", nullable = false, length = 100)
    private String typeName;

    @Column(name = "description", length = 200)
    private String description;

    @OneToMany(mappedBy = "gymType", cascade = CascadeType.ALL)
    private Set<Gyms> gyms = new HashSet<>();
}