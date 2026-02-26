package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "nutrition_plan")
public class NutritionPlan {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_plan", nullable = false)
    private int idPlan;

    @Column(name = "plan_name", nullable = false, length = 100)
    private String planName;

    @Column(name = "nutrition_description", length = 200)
    private String nutritionDescription;

    @Column(name = "goal_type", length = 50)
    private String goalType; // цель

    @Column(name = "difficulty_level", length = 20)
    private String difficultyLevel; // легкий, средний, сложный

    @Column(name = "calories_per_day")
    private Integer caloriesPerDay;

    @Column(name = "protein_percent")
    private Integer proteinPercent;

    @Column(name = "carbs_percent")
    private Integer carbsPercent;

    @Column(name = "fat_percent")
    private Integer fatPercent;
}