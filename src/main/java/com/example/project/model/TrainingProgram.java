package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;
import java.time.LocalDate;
import java.util.HashSet;
import java.util.Set;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "training_programs")
public class TrainingProgram {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_program")
    private Integer idProgram;
    
    @ManyToOne
    @JoinColumn(name = "id_member", nullable = false)
    private Members member;
    
    // ДОБАВИТЬ: связь с планом питания
    @ManyToOne
    @JoinColumn(name = "id_nutrition_plan")
    private NutritionPlan nutritionPlan;
    
    @Column(name = "program_name", nullable = false, length = 100)
    private String programName;
    
    @Column(name = "goal", length = 100)
    private String goal;
    
    @Column(name = "level")
    private String level;
    
    @Column(name = "duration_weeks")
    private Integer durationWeeks;
    
    @Column(name = "created_date")
    private LocalDate createdDate;
    
    @Column(name = "is_active")
    private Boolean isActive;
    
    @OneToMany(mappedBy = "trainingProgram", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<ProgramDay> programDays = new HashSet<>();
}