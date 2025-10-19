package com.example.project.repository;

import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.Position;

@Repository
public interface PositionRepository extends JpaRepository<Position, Integer> {
    
    Optional<Position> findByRoleName(String roleName);
    
    Set<Position> findByRoleNameContaining(String roleName);
}