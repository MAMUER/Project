package com.example.project.service;

import lombok.*;

@Getter
@AllArgsConstructor
@Setter
public class Event {
    private String title;
    private String start;
    private String end;
    private int id;
    private String color;
}