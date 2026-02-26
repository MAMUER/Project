package com.example.project.service;

import lombok.AllArgsConstructor;

import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.dto.NewsDTO;
import com.example.project.model.News;
import com.example.project.repository.NewsRepository;

@Service
@AllArgsConstructor
@Transactional
public class NewsService {
    private final NewsRepository newsRepository;

    @SuppressWarnings("null")
    public News getNews(Integer id) {
        return newsRepository.findById(id).orElse(null);
    }

    @Transactional(readOnly = true)
    public List<News> getAllNews() {
        return newsRepository.findAll();
    }

    @Transactional(readOnly = true)
    public List<NewsDTO> getAllNewsWithClubsDTO() {
        List<News> news = newsRepository.findAllWithClubs();
        return news.stream()
                .map(NewsDTO::fromEntity)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<NewsDTO> getNewsByClubDTO(String clubName) {
        List<News> news = newsRepository.findByClubsClubNameWithClubs(clubName);
        return news.stream()
                .map(NewsDTO::fromEntity)
                .collect(Collectors.toList());
    }

    public Set<News> getNewsByTitle(String title) {
        return newsRepository.findByNewsTitleContaining(title);
    }

    public Set<News> getDetailedNews(int minLength) {
        return newsRepository.findByNewsTextLengthGreaterThan(minLength);
    }

    @SuppressWarnings("null")
    public News saveNews(News news) {
        return newsRepository.save(news);
    }

    @SuppressWarnings("null")
    public void deleteNews(Integer id) {
        newsRepository.deleteById(id);
    }
}