package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.TrainersPhoto;

@Repository
public interface TrainersPhotoRepository extends JpaRepository<TrainersPhoto, Integer> {
    
    Set<TrainersPhoto> findByImageUrlContaining(String keyword);
}