package com.example.project.service;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Equipment;
import com.example.project.repository.EquipmentRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class EquipmentService {
    private final EquipmentRepository equipmentRepository;

    public Equipment getEquipment(Integer id) {
        return equipmentRepository.findById(id).orElse(null);
    }

    public List<Equipment> getAllEquipment() {
        return equipmentRepository.findAll();
    }

    public Set<Equipment> getAvailableEquipment() {
        return equipmentRepository.findAvailableEquipment();
    }

    public Set<Equipment> getEquipmentByName(String name) {
        return equipmentRepository.findByNameContaining(name);
    }

    public Set<Equipment> getEquipmentByType(Integer typeId) {
        return equipmentRepository.findByEquipmentTypeIdEquipmentType(typeId);
    }

    public Equipment saveEquipment(Equipment equipment) {
        return equipmentRepository.save(equipment);
    }
}