package com.example.project.repository;

import java.time.LocalDate;
import java.util.Set;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import com.example.project.model.Members;

@Repository
public interface MembersRepository extends JpaRepository<Members, Integer> {

    Set<Members> findByFirstNameContaining(String firstName);

    Set<Members> findBySecondNameContaining(String secondName);

    Set<Members> findByFirstNameContainingAndSecondNameContaining(String firstName, String secondName);

    Set<Members> findByClubClubName(String clubName);

    Set<Members> findByBirthDateBetween(LocalDate startDate, LocalDate endDate);

    Set<Members> findByGender(Integer gender);

    // ИСПРАВЛЕНО: правильное название таблицы в БД
    @Query(value = "SELECT COUNT(*) > 0 FROM members_have_training_schedule WHERE id_member = :memberId AND id_session = :trainingId", nativeQuery = true)
    boolean existsTrainingForMember(@Param("memberId") Integer memberId,
            @Param("trainingId") Integer trainingId);
}