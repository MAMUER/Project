package com.example.project.model;

import lombok.*;

import java.time.LocalDate;

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

    @ManyToOne
    @JoinColumn(name = "id_member", nullable = false)
    private Members member;

    @Column(name = "nutrition_description", length = 100)
    private String nutritionDescription;

    @Column(name = "start_date", nullable = false)
    private LocalDate startDate;

    @Column(name = "end_date")
    private LocalDate endDate;
}