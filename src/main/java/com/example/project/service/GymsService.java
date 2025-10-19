package com.example.project.service;

import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Equipment;
import com.example.project.model.Gyms;
import com.example.project.repository.GymsRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class GymsService {
    private final GymsRepository gymsRepository;

    public Gyms getGym(Integer id) {
        return gymsRepository.findById(id).orElse(null);
    }

    public List<Gyms> getAllGyms() {
        return gymsRepository.findAll();
    }

    public Set<Gyms> getGymsByClub(String clubName) {
        return gymsRepository.findByClubClubName(clubName);
    }

    public Set<Gyms> getGymsByCapacity(int minCapacity) {
        return gymsRepository.findByCapacityGreaterThanEqual(minCapacity);
    }

    public Set<Gyms> getGymsByAvailableHours(int minHours) {
        return gymsRepository.findByAvailableHoursGreaterThan(minHours);
    }

    public Set<Equipment> getGymEquipment(Integer gymId) {
        Gyms gym = gymsRepository.findById(gymId).orElse(null);
        return gym != null ? new HashSet<>(gym.getEquipment()) : Collections.emptySet();
    }

    public Gyms saveGym(Gyms gym) {
        return gymsRepository.save(gym);
    }
}