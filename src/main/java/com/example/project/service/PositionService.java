package com.example.project.service;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Position;
import com.example.project.repository.PositionRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class PositionService {
    private final PositionRepository positionRepository;

    @SuppressWarnings("null")
    public Position getPosition(Integer id) {
        return positionRepository.findById(id).orElse(null);
    }

    public Position getPositionByName(String roleName) {
        return positionRepository.findByRoleName(roleName).orElse(null);
    }

    public List<Position> getAllPositions() {
        return positionRepository.findAll();
    }

    public Set<Position> searchPositions(String roleName) {
        return positionRepository.findByRoleNameContaining(roleName);
    }

    @SuppressWarnings("null")
    public Position savePosition(Position position) {
        return positionRepository.save(position);
    }
}