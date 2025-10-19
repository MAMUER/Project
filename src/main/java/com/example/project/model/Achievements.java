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
@Table(name = "achievements")
public class Achievements {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_achievement", nullable = false)
    private int idAchievement;

    @Column(name = "achievement_description")
    private String achievementDescription;

    @Column(name = "achievement_title", nullable = false)
    private String achievementTitle;

    @Column(name = "achievement_icon_url", nullable = false)
    private String achievementIconUrl;

    @OneToMany(mappedBy = "achievement", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<MembersHaveAchievements> membersHaveAchievements = new HashSet<>();
}