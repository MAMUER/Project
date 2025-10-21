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
@Table(name = "program_days")
public class ProgramDay {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_day")
    private Integer idDay;
    
    @ManyToOne
    @JoinColumn(name = "id_program", nullable = false)
    private TrainingProgram trainingProgram;
    
    @Column(name = "day_number")
    private Integer dayNumber;
    
    @Column(name = "day_name", length = 20)
    private String dayName;
    
    @Column(name = "muscle_groups", length = 100)
    private String muscleGroups;
    
    @OneToMany(mappedBy = "programDay", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<ProgramExercise> exercises = new HashSet<>();
}