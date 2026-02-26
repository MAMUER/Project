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

    @SuppressWarnings("null")
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

    @SuppressWarnings("null")
    public void saveEquipment(Equipment equipment) {
        equipmentRepository.save(equipment);
    }

    // ДОБАВИТЬ: метод для удаления оборудования
    @SuppressWarnings("null")
    @Transactional
    public void deleteEquipment(Integer equipmentId) {
        try {
            if (equipmentRepository.existsById(equipmentId)) {
                equipmentRepository.deleteByIdEquipment(equipmentId);
            }
        } catch (Exception e) {
        }
    }

    // ДОБАВИТЬ: метод для проверки существования оборудования
    @SuppressWarnings("null")
    public boolean existsById(Integer equipmentId) {
        return equipmentRepository.existsById(equipmentId);
    }
}