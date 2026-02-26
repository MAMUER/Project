package com.example.project.repository;

import java.util.Set;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import com.example.project.model.Gyms;

@Repository
public interface GymsRepository extends JpaRepository<Gyms, Integer> {
    
    Set<Gyms> findByClubClubName(String clubName);
    
    Set<Gyms> findByCapacityGreaterThanEqual(int minCapacity);
    
    Set<Gyms> findByAvailableHoursGreaterThan(int minHours);
    
    Set<Gyms> findByCapacityGreaterThanEqualAndAvailableHoursGreaterThan(int minCapacity, int minHours);
    
    // ДОБАВИТЬ: поиск по типу зала
    Set<Gyms> findByGymTypeGymTypeId(Integer gymTypeId);
}