package com.example.project.repository;

import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.ActivityType;

@Repository
public interface ActivityTypeRepository extends JpaRepository<ActivityType, Integer> {

    Optional<ActivityType> findByActivityName(String name);

    Set<ActivityType> findByActivityNameContaining(String name);
}