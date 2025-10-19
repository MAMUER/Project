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
@Table(name = "training_type")
public class TrainingType {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_training_type", nullable = false)
    private int idTrainingType;

    @Column(name = "training_type_name", nullable = false, length = 45)
    private String trainingTypeName;

    @Column(name = "workout_description", length = 300)
    private String workoutDescription;

    @OneToMany(mappedBy = "trainingType", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<TrainingSchedule> trainingSchedules = new HashSet<>();
}