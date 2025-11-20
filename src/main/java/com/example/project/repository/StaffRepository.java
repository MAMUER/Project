package com.example.project.repository;

import java.time.LocalDate;
import java.util.Optional;
import java.util.Set;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import com.example.project.model.Staff;

@Repository
public interface StaffRepository extends JpaRepository<Staff, Integer> {

    @Query("SELECT s FROM Staff s LEFT JOIN FETCH s.staffSchedules WHERE s.idStaff = :staffId")
    Optional<Staff> findByIdWithSchedules(@Param("staffId") Integer staffId);

    Set<Staff> findByFirstNameContaining(String firstName);
    Set<Staff> findBySecondNameContaining(String secondName);
    Set<Staff> findByPositionIdPosition(Integer positionId);
    Set<Staff> findByHireDateAfter(LocalDate hireDate);
    
    // ИСПРАВЛЕНО: убрать - поля staffAbout может быть null, но метод работает
    Set<Staff> findByStaffAboutIsNotNull();
    
    Set<Staff> findByFirstNameContainingAndSecondNameContaining(String firstName, String secondName);
}