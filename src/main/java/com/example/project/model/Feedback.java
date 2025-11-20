package com.example.project.model;

import lombok.*;

import java.time.LocalDate;

import com.example.project.model.Accounts.MembersAccounts;

import jakarta.persistence.*;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "feedback")
public class Feedback {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id_feedback", nullable = false)
    private Integer idFeedback;

    // ИСПРАВЛЕНО: используем строку вместо объекта
    @Column(name = "username", nullable = false, length = 45)
    private String username;

    @Column(name = "feedback_text", length = 45)
    private String feedbackText;

    @Column(name = "feedback_date", nullable = false)
    private LocalDate feedbackDate;

    @Column(name = "rating", nullable = false)
    private Short rating; // Используем Short вместо short для совместимости с NULL

    // Опционально: связь для получения данных аккаунта
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "username", referencedColumnName = "username", insertable = false, updatable = false)
    private MembersAccounts memberAccount;
}