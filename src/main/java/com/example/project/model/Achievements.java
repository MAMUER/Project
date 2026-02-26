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
@Table(name = "achievements")
public class Achievements {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_achievement", nullable = false)
    private Integer idAchievement;

    @Column(name = "achievement_description", length = 45)
    private String achievementDescription;

    @Column(name = "achievement_title", nullable = false, length = 45)
    private String achievementTitle;

    @Column(name = "achievement_icon_url", nullable = false, length = 100)
    private String achievementIconUrl;

    @OneToMany(mappedBy = "achievement", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<MembersHaveAchievements> membersHaveAchievements = new HashSet<>();
}