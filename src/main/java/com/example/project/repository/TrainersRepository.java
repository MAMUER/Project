package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.Trainers;

@Repository
public interface TrainersRepository extends JpaRepository<Trainers, Integer> {

    Set<Trainers> findByFirstNameContaining(String firstName);

    Set<Trainers> findBySecondNameContaining(String secondName);

    Set<Trainers> findBySpecialityContaining(String speciality);

    Set<Trainers> findByExperienceGreaterThanEqual(int minExperience);

    Set<Trainers> findByCertificationsGreaterThan(int minCertifications);

    Set<Trainers> findByHireDateAfter(LocalDate hireDate);
}