package com.example.project.repository;

import java.util.List;
import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.project.model.Equipment;

@Repository
public interface EquipmentRepository extends JpaRepository<Equipment, Integer> {

    Set<Equipment> findByQuantityGreaterThan(int quantity);

    Set<Equipment> findByEquipmentTypeIdEquipmentType(Integer equipmentTypeId);

    default Set<Equipment> findAvailableEquipment() {
        return findByQuantityGreaterThan(0);
    }

    // ПРАВИЛЬНЫЕ методы для оборудования по клубу
    @Query("SELECT e FROM Equipment e WHERE e.club.clubName = :clubName")
    List<Equipment> findByClubName(@Param("clubName") String clubName);

    @Query("SELECT e FROM Equipment e WHERE e.club.clubName = :clubName AND e.equipmentType.idEquipmentType = :equipmentTypeId")
    List<Equipment> findByClubNameAndEquipmentTypeId(@Param("clubName") String clubName,
            @Param("equipmentTypeId") Integer equipmentTypeId);

    @Query("SELECT e FROM Equipment e WHERE e.club.clubName = :clubName AND e.equipmentType.typeName = :typeName")
    List<Equipment> findByClubNameAndEquipmentTypeName(@Param("clubName") String clubName,
            @Param("typeName") String typeName);

    // ДОБАВИТЬ: метод для удаления оборудования по ID
    @Modifying
    @Query("DELETE FROM Equipment e WHERE e.idEquipment = :equipmentId")
    void deleteByIdEquipment(@Param("equipmentId") Integer equipmentId);
}