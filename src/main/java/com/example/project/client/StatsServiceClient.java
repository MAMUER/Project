package com.example.project.client;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Component
public class StatsServiceClient {

    private final RestTemplate restTemplate;
    private final String statsServiceUrl;

    public StatsServiceClient(
            @Value("${stats.service.url:http://localhost:8081}") String statsServiceUrl) {
        this.restTemplate = new RestTemplate();
        this.statsServiceUrl = statsServiceUrl;
        log.info("StatsServiceClient initialized with URL: {}", statsServiceUrl);
    }

    /**
     * Получить количество посещений за сегодня для конкретного клуба
     */
    public Integer getTodayVisits(Integer clubId) {
        if (clubId == null) {
            log.warn("ClubId is null, returning default value 0");
            return 0;
        }

        String url = statsServiceUrl + "/api/stats/club/" + clubId + "/today";
        log.debug("Calling stats service: {}", url);

        try {
            ParameterizedTypeReference<Map<String, Object>> responseType = 
                new ParameterizedTypeReference<Map<String, Object>>() {};
            
            @SuppressWarnings("null")
            ResponseEntity<Map<String, Object>> response = 
                restTemplate.exchange(url, HttpMethod.GET, null, responseType);
            
            Map<String, Object> body = response.getBody();
            
            if (body != null && body.containsKey("todayVisits")) {
                Object value = body.get("todayVisits");
                if (value instanceof Number number) {
                    return number.intValue();
                }
            }
            
            log.warn("Unexpected response from stats service: {}", body);
            return 0;
            
        } catch (HttpClientErrorException.NotFound | ResourceAccessException e) {
            log.error("Stats service error ({}): {}", url, e.getMessage());
            return 0;
        } catch (RestClientException e) {
            log.error("Error calling stats service: {}", e.getMessage(), e);
            return 0;
        }
    }

    /**
     * Получить топ активных участников
     */
    public List<Map<String, Object>> getTopActiveMembers() {
        String url = statsServiceUrl + "/api/stats/members/top";
        log.debug("Calling stats service: {}", url);

        try {
            ParameterizedTypeReference<List<Map<String, Object>>> responseType = 
                new ParameterizedTypeReference<List<Map<String, Object>>>() {};
            
            @SuppressWarnings("null")
            ResponseEntity<List<Map<String, Object>>> response = 
                restTemplate.exchange(url, HttpMethod.GET, null, responseType);
            
            List<Map<String, Object>> body = response.getBody();
            return body != null ? body : Collections.emptyList();
            
        } catch (HttpClientErrorException.NotFound | ResourceAccessException e) {
            log.error("Stats service error ({}): {}", url, e.getMessage());
            return Collections.emptyList();
        } catch (RestClientException e) {
            log.error("Error getting top active members from stats service: {}", e.getMessage());
            return Collections.emptyList();
        }
    }

    /**
     * Получить статистику посещаемости за период
     */
    public Map<String, Object> getVisitsStats(Integer clubId, String period) {
        String url = statsServiceUrl + "/api/stats/club/" + clubId + "/visits?period=" + period;
        log.debug("Calling stats service: {}", url);

        try {
            ParameterizedTypeReference<Map<String, Object>> responseType = 
                new ParameterizedTypeReference<Map<String, Object>>() {};
            
            @SuppressWarnings("null")
            ResponseEntity<Map<String, Object>> response = 
                restTemplate.exchange(url, HttpMethod.GET, null, responseType);
            
            return response.getBody() != null ? response.getBody() : new HashMap<>();
            
        } catch (HttpClientErrorException.NotFound | ResourceAccessException e) {
            log.error("Stats service error ({}): {}", url, e.getMessage());
            return new HashMap<>();
        } catch (RestClientException e) {
            log.error("Error getting visits stats: {}", e.getMessage());
            return new HashMap<>();
        }
    }

    /**
     * Проверить здоровье сервиса статистики
     */
    public boolean isServiceHealthy() {
        String url = statsServiceUrl + "/health";
        log.debug("Checking stats service health: {}", url);

        try {
            ResponseEntity<HealthResponse> response = 
                restTemplate.getForEntity(url, HealthResponse.class);
            
            HealthResponse body = response.getBody();
            return body != null && "healthy".equals(body.getStatus());
            
        } catch (RestClientException e) {
            log.warn("Stats service health check failed: {}", e.getMessage());
            return false;
        }
    }

    /**
     * DTO для ответа health check
     */
    public static class HealthResponse {
        private String status;
        
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
    }
}