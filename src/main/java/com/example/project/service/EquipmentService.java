package com.example.project.service;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

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

    public Set<Equipment> getEquipmentByType(Integer typeId) {
        return equipmentRepository.findByEquipmentTypeIdEquipmentType(typeId);
    }

    public Equipment saveEquipment(Equipment equipment) {
        return equipmentRepository.save(equipment);
    }

    // ДОБАВИТЬ: метод для удаления оборудования
    @Transactional
    public boolean deleteEquipment(Integer equipmentId) {
        try {
            if (equipmentRepository.existsById(equipmentId)) {
                equipmentRepository.deleteByIdEquipment(equipmentId);
                return true;
            }
            return false;
        } catch (Exception e) {
            return false;
        }
    }

    // ДОБАВИТЬ: метод для проверки существования оборудования
    public boolean existsById(Integer equipmentId) {
        return equipmentRepository.existsById(equipmentId);
    }
}