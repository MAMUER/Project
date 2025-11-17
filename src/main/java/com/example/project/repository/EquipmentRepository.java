package com.example.project.repository;

import java.util.Set;
import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.project.model.Equipment;

@Repository
public interface EquipmentRepository extends JpaRepository<Equipment, Integer> {

    Set<Equipment> findByNameContaining(String name);

    Set<Equipment> findByQuantityGreaterThan(int quantity);

    Set<Equipment> findByEquipmentTypeIdEquipmentType(Integer equipmentTypeId);

    default Set<Equipment> findAvailableEquipment() {
        return findByQuantityGreaterThan(0);
    }

    // ИСПРАВЛЯЕМ метод - убираем неиспользуемый параметр gymId
    @Query("SELECT e FROM Equipment e WHERE e.gym.idGym = :gymId AND e.equipmentType.idEquipmentType = :equipmentTypeId")
    List<Equipment> findByGymIdAndEquipmentTypeId(@Param("gymId") Integer gymId,
            @Param("equipmentTypeId") Integer equipmentTypeId);

    // ОСНОВНОЙ метод для получения оборудования клуба по типу
    @Query("SELECT e FROM Equipment e WHERE e.gym.club.clubName = :clubName AND e.equipmentType.idEquipmentType = :equipmentTypeId")
    List<Equipment> findByClubNameAndEquipmentTypeId(@Param("clubName") String clubName,
            @Param("equipmentTypeId") Integer equipmentTypeId);

    // ПРАВИЛЬНО - используем g.club.clubName
    @Query("SELECT e FROM Equipment e " +
            "JOIN e.gyms g " +
            "WHERE g.club.clubName = :clubName")
    List<Equipment> findByClubName(@Param("clubName") String clubName);

    // ДОБАВЛЯЕМ метод для получения оборудования по клубу и названию типа
    @Query("SELECT e FROM Equipment e WHERE e.gym.club.clubName = :clubName AND e.equipmentType.typeName = :typeName")
    List<Equipment> findByClubNameAndEquipmentTypeName(@Param("clubName") String clubName,
            @Param("typeName") String typeName);
}