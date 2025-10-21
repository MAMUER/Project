package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "program_exercises")
public class ProgramExercise {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_program_exercise")
    private Integer idProgramExercise;
    
    @ManyToOne
    @JoinColumn(name = "id_day", nullable = false)
    private ProgramDay programDay;
    
    @ManyToOne
    @JoinColumn(name = "id_exercise", nullable = false)
    private Exercise exercise;
    
    @Column(name = "sets")
    private Integer sets;
    
    @Column(name = "reps")
    private Integer reps;
    
    @Column(name = "weight")
    private Double weight;
    
    @Column(name = "rest_seconds")
    private Integer restSeconds;
    
    @Column(name = "order_index")
    private Integer orderIndex;
}