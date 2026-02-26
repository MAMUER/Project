package com.example.project.repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.StaffSchedule;

@Repository
public interface StaffScheduleRepository extends JpaRepository<StaffSchedule, Integer> {

    List<StaffSchedule> findByStaffIdStaff(Integer staffId);

    Set<StaffSchedule> findByClubClubName(String clubName);

    Set<StaffSchedule> findByDateBetween(LocalDate start, LocalDate end);

    Set<StaffSchedule> findByShift(Integer shift);

    Set<StaffSchedule> findByStaffIdStaffAndDateBetween(Integer staffId, LocalDate startDate, LocalDate endDate);
}