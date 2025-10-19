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
    private int idFeedback;

    @ManyToOne
    @JoinColumn(name = "username", nullable = false)
    private MembersAccounts memberAccount;

    @Column(name = "feedback_text", length = 45)
    private String feedbackText;

    @Column(name = "feedback_date", nullable = false)
    private LocalDate feedbackDate;

    @Column(name = "rating", nullable = false)
    private short rating;
}