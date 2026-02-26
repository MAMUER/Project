package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.StaffPhoto;

@Repository
public interface StaffPhotoRepository extends JpaRepository<StaffPhoto, Integer> {
    
    Set<StaffPhoto> findByImageUrlContaining(String keyword);
}