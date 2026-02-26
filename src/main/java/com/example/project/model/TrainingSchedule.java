package com.example.project.model;

import lombok.*;

import java.time.LocalDateTime;
import java.util.HashSet;
import java.util.Set;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "training_schedule")
public class TrainingSchedule {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_session", nullable = false)
    private int idSession;

    @ManyToOne
    @JoinColumn(name = "id_trainer", nullable = false)
    private Trainers trainer;

    @ManyToOne
    @JoinColumn(name = "id_training_type")
    private TrainingType trainingType;

    @Column(name = "session_date", nullable = false)
    private LocalDateTime sessionDate;

    @Column(name = "session_time", nullable = false)
    private int sessionTime;

    @ManyToMany(mappedBy = "trainingSchedules", fetch = FetchType.LAZY)
    private Set<Members> members = new HashSet<>();

    public Set<Members> getMembers() {
        return members;
    }
    
    public void setMembers(Set<Members> members) {
        this.members = members;
    }
}