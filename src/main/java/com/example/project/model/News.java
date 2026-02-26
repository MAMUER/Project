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
@Table(name = "news")
public class News {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_news", nullable = false)
    private int idNews;

    @Column(name = "news_title", nullable = false, length = 45)
    private String newsTitle;

    @Column(name = "news_text", length = 200)
    private String newsText;

    @ManyToMany(mappedBy = "news")
    private Set<Clubs> clubs = new HashSet<>();
}