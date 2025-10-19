package com.example.project.repository;

import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.EquipmentType;

@Repository
public interface EquipmentTypeRepository extends JpaRepository<EquipmentType, Integer> {
    
    Optional<EquipmentType> findByTypeName(String typeName);
    
    Set<EquipmentType> findByTypeNameContaining(String name);
}