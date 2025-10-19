package com.example.project.repository;

import java.util.List;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.project.model.News;

@Repository
public interface NewsRepository extends JpaRepository<News, Integer> {

    Set<News> findByNewsTitleContaining(String title);

    @Query("SELECT n FROM News n WHERE LENGTH(n.newsText) > :minLength")
    Set<News> findByNewsTextLengthGreaterThan(@Param("minLength") int minLength);

    // Метод с JOIN FETCH для загрузки клубов
    @Query("SELECT DISTINCT n FROM News n LEFT JOIN FETCH n.clubs")
    List<News> findAllWithClubs();

    // Метод для фильтрации по клубу с JOIN FETCH
    @Query("SELECT DISTINCT n FROM News n LEFT JOIN FETCH n.clubs c WHERE c.clubName = :clubName")
    List<News> findByClubsClubNameWithClubs(@Param("clubName") String clubName);
}