package com.example.project.dto;

import lombok.Data;
import java.util.List;

@Data
public class ProgramRequest {
    private String goal; // цель
    private String level; // начальный, средний, продвинутый
    private Integer durationWeeks;
    private List<String> trainingDays; // выбранные дни тренировок
    private String preferredTime; // предпочтительное время: УТРО, ДЕНЬ, ВЕЧЕР
}