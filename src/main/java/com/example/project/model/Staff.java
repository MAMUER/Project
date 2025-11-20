package com.example.project.model;

import lombok.*;
import jakarta.persistence.*;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "staff")
public class Staff {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_staff", nullable = false)
    private int idStaff;

    @ManyToOne
    @JoinColumn(name = "id_position")
    private Position position;

    @Column(name = "first_name", nullable = false, length = 45)
    private String firstName;

    @Column(name = "second_name", nullable = false, length = 45)
    private String secondName;

    @Column(name = "hire_date", nullable = false)
    private LocalDate hireDate;

    @Column(name = "staff_about", length = 100)
    private String staffAbout;

    @OneToOne(mappedBy = "staff", cascade = CascadeType.ALL, orphanRemoval = true)
    private com.example.project.model.Accounts.StaffAccounts staffAccount;

    @OneToMany(mappedBy = "staff", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<StaffSchedule> staffSchedules = new ArrayList<>();
}