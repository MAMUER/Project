package com.example.project.repository;

import java.util.Set;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.Clubs;

@Repository
public interface ClubsRepository extends JpaRepository<Clubs, String> {

    Optional<Clubs> findByClubName(String clubName);

    Set<Clubs> findByAddressContaining(String address);

    Set<Clubs> findByClubNameContaining(String clubName);
}
