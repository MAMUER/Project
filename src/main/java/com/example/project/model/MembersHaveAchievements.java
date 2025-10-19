package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;

import java.time.LocalDate;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "members_have_achievements")
public class MembersHaveAchievements {
    
    @EmbeddedId
    private MembersHaveAchievementsId id;

    @ManyToOne
    @MapsId("memberId")
    @JoinColumn(name = "id_member", nullable = false)
    private Members member;

    @ManyToOne
    @MapsId("achievementId")
    @JoinColumn(name = "id_achievement", nullable = false)
    private Achievements achievement;

    @Column(name = "receipt_date", nullable = false)
    private LocalDate receiptDate;
}