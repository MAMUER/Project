package com.example.project.dto;

import lombok.*;
import java.time.LocalDate;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class MemberAchievementDTO {
    private Integer memberId;
    private String memberName;
    private Integer achievementId;
    private String achievementTitle; // ИСПРАВЛЕНО: achievementTitle вместо achievementName
    private String achievementDescription;
    private String achievementIconUrl;
    private LocalDate receiptDate;
    
    public static MemberAchievementDTO fromEntity(com.example.project.model.MembersHaveAchievements entity) {
        return new MemberAchievementDTO(
            entity.getMember().getIdMember(),
            entity.getMember().getFirstName() + " " + entity.getMember().getSecondName(),
            entity.getAchievement().getIdAchievement(),
            entity.getAchievement().getAchievementTitle(), // ИСПРАВЛЕНО
            entity.getAchievement().getAchievementDescription(),
            entity.getAchievement().getAchievementIconUrl(),
            entity.getReceiptDate()
        );
    }
}