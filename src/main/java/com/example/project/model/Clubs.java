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
@Table(name = "clubs")
public class Clubs {
    @Id
    @Column(name = "club_name", nullable = false, length = 45)
    private String clubName;

    @Column(name = "address", nullable = false, length = 45)
    private String address;

    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Members> members = new HashSet<>();

    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Gyms> gyms = new HashSet<>();

    @OneToMany(mappedBy = "club", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<StaffSchedule> staffSchedules = new HashSet<>();

    @ManyToMany
    @JoinTable(
        name = "clubs_have_news",
        joinColumns = @JoinColumn(name = "club_name"),
        inverseJoinColumns = @JoinColumn(name = "id_news")
    )
    private Set<News> news = new HashSet<>();
}