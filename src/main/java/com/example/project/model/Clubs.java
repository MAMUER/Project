package com.example.project.model;

import lombok.*;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.util.HashSet;
import java.util.Set;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@ToString // ДОБАВИТЬ: для корректной работы в Thymeleaf
@Entity
@Table(name = "clubs")
public class Clubs {
    @Id
    @Column(name = "club_name", nullable = false, length = 45)
    private String clubName;

    @Column(name = "address", nullable = false, length = 45)
    private String address;

    @Column(name = "schedule", columnDefinition = "jsonb")
    @JdbcTypeCode(SqlTypes.JSON)
    private String schedule;

    // ИСПРАВЛЕНО: оставить LAZY, но добавить @ToString.Exclude
    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.LAZY)
    @ToString.Exclude
    private Set<Members> members = new HashSet<>();

    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.LAZY)
    @ToString.Exclude
    private Set<Gyms> gyms = new HashSet<>();

    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.LAZY)
    @ToString.Exclude
    private Set<StaffSchedule> staffSchedules = new HashSet<>();

    @ManyToMany(fetch = FetchType.LAZY)
    @JoinTable(name = "clubs_have_news", joinColumns = @JoinColumn(name = "club_name"), inverseJoinColumns = @JoinColumn(name = "id_news"))
    @ToString.Exclude
    private Set<News> news = new HashSet<>();

    @ManyToMany(fetch = FetchType.LAZY)
    @JoinTable(name = "clubs_have_equipment", joinColumns = @JoinColumn(name = "club_name"), inverseJoinColumns = @JoinColumn(name = "id_equipment"))
    @ToString.Exclude
    private Set<Equipment> equipment = new HashSet<>();
}