package com.example.project.service;

import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.ActivityType;
import com.example.project.repository.ActivityTypeRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class ActivityTypeService {
    private final ActivityTypeRepository activityTypeRepository;

    public ActivityType getActivityType(Integer id) {
        return activityTypeRepository.findById(id).orElse(null);
    }

    public ActivityType getActivityTypeByName(String name) {
        return activityTypeRepository.findByActivityName(name).orElse(null);
    }

    public List<ActivityType> getAllActivityTypes() {
        return activityTypeRepository.findAll();
    }

    public Set<ActivityType> getActivityTypesByNameContaining(String name) {
        return activityTypeRepository.findByActivityNameContaining(name);
    }
}