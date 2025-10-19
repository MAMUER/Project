package com.example.project.repository;

import java.util.Set;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.UsersPhoto;

@Repository
public interface UsersPhotoRepository extends JpaRepository<UsersPhoto, Integer> {
    
    Set<UsersPhoto> findByImageUrlContaining(String keyword);
}