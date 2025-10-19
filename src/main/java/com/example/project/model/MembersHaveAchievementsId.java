package com.example.project.model;

import lombok.*;
import jakarta.persistence.Embeddable;
import java.io.Serializable;

@Embeddable
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@EqualsAndHashCode
public class MembersHaveAchievementsId implements Serializable {
    private Integer memberId;
    private Integer achievementId;
}