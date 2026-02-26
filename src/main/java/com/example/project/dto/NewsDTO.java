package com.example.project.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.stream.Collectors;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class NewsDTO {
    private int idNews;
    private String newsTitle;
    private String newsText;
    private List<String> clubNames;

    public static NewsDTO fromEntity(com.example.project.model.News news) {
        List<String> clubNames = news.getClubs().stream()
                .map(com.example.project.model.Clubs::getClubName)
                .collect(Collectors.toList());

        return new NewsDTO(
                news.getIdNews(),
                news.getNewsTitle(),
                news.getNewsText(),
                clubNames);
    }
}