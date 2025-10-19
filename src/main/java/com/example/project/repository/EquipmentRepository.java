package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
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
}
