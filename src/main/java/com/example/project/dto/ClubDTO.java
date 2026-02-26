package com.example.project.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ClubDTO {
    private String clubName;
    private String address;
    private String schedule; // добавим если нужно
    
    public static ClubDTO fromEntity(com.example.project.model.Clubs club) {
        if (club == null) {
            return null;
        }
        // ПРИНУДИТЕЛЬНО инициализируем прокси перед созданием DTO
        if (club instanceof org.hibernate.proxy.HibernateProxy hibernateProxy) {
            club = (com.example.project.model.Clubs) hibernateProxy
                    .getHibernateLazyInitializer().getImplementation();
        }
        return new ClubDTO(club.getClubName(), club.getAddress(), club.getSchedule());
    }
}