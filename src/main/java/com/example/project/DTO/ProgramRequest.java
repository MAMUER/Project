package com.example.project.dto;

import lombok.Data;

@Data
public class ProgramRequest {
    private String goal; // похудение, набор_массы, поддержание
    private String level; // начальный, средний, продвинутый
    private Integer durationWeeks;
    // УБИРАЕМ clubName - используем клуб пользователя из профиля
}