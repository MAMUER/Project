package com.example.project.model;

import lombok.*;

import java.time.LocalDate;
import java.util.HashSet;
import java.util.Set;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "visits_history")
public class VisitsHistory {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_visit", nullable = false)
    private int idVisit;

    @Column(name = "visit_date")
    private LocalDate visitDate;

    @ManyToMany(mappedBy = "visitsHistory")
    private Set<Members> members = new HashSet<>();
}