package com.example.project.service;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.project.model.EquipmentType;
import com.example.project.repository.EquipmentTypeRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class EquipmentTypeService {
    private final EquipmentTypeRepository equipmentTypeRepository;

    @SuppressWarnings("null")
    public EquipmentType getEquipmentType(Integer id) {
        return equipmentTypeRepository.findById(id).orElse(null);
    }

    public EquipmentType getEquipmentTypeByName(String name) {
        return equipmentTypeRepository.findByTypeName(name).orElse(null);
    }

    public List<EquipmentType> getAllEquipmentTypes() {
        return equipmentTypeRepository.findAll();
    }

    @SuppressWarnings("null")
    public EquipmentType saveEquipmentType(EquipmentType equipmentType) {
        return equipmentTypeRepository.save(equipmentType);
    }
}