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
@Table(name = "exercises") // Нужно создать таблицу в БД
public class Exercise {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_exercise") // ДОБАВИТЬ явное указание имени столбца
    private Integer idExercise;

    @Column(name = "exercise_name", nullable = false, length = 100)
    private String exerciseName;

    @Column(name = "description", length = 300)
    private String description;

    @Column(name = "muscle_group", length = 100)
    private String muscleGroup; // грудные, ноги, спина и т.д.

    @Column(name = "difficulty_level")
    private Integer difficultyLevel; // 1-легкий, 2-средний, 3-сложный

    @Column(name = "equipment_required")
    private String equipmentRequired; // тип необходимого оборудования

    @Column(name = "estimated_calories")
    private Integer estimatedCalories;

    @OneToMany(mappedBy = "exercise", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<ProgramExercise> programExercises = new HashSet<>();

    @OneToMany(mappedBy = "exercise", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<ExerciseEquipmentRequirement> equipmentRequirements = new HashSet<>();
}