package com.example.project.service;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import com.example.project.model.Clubs;
import com.example.project.model.Gyms;
import com.example.project.model.News;
import com.example.project.model.StaffSchedule;
import com.example.project.repository.ClubsRepository;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Set;

@Service
@AllArgsConstructor
public class ClubsService {
    private final ClubsRepository clubsRepository;

    public Clubs getClub(String id) {
        return clubsRepository.findById(id).orElse(null);
    }

    public List<Clubs> getAllClubs() {
        return clubsRepository.findAll();
    }

    public Set<Clubs> getClubsByAddress(String address) {
        return clubsRepository.findByAddressContaining(address);
    }

    public List<News> getListOfClubNews(String id) {
        Clubs club = clubsRepository.findById(id).orElse(null);
        return club != null ? new ArrayList<>(club.getNews()) : Collections.emptyList();
    }

    public Set<Gyms> getSetOfGyms(String id) {
        Clubs club = clubsRepository.findById(id).orElse(null);
        return club != null ? club.getGyms() : Collections.emptySet();
    }

    public Set<StaffSchedule> getSetOfStaffSchedules(String clubsId) {
        Clubs club = clubsRepository.findById(clubsId).orElse(null);
        return club != null ? club.getStaffSchedules() : Collections.emptySet();
    }

    public Clubs saveClub(Clubs club) {
        return clubsRepository.save(club);
    }

    public void deleteClub(String id) {
        clubsRepository.deleteById(id);
    }

    public Clubs getClubByName(String clubName) {
        return clubsRepository.findById(clubName).orElse(null);
    }

    public Clubs updateClub(String clubName, Clubs updatedClub) {
        Clubs existingClub = clubsRepository.findById(clubName)
                .orElseThrow(() -> new RuntimeException("Club not found: " + clubName));

        // Обновляем поля
        if (updatedClub.getAddress() != null) {
            existingClub.setAddress(updatedClub.getAddress());
        }
        if (updatedClub.getSchedule() != null) {
            existingClub.setSchedule(updatedClub.getSchedule());
        }

        return clubsRepository.save(existingClub);
    }

    public boolean existsById(String clubName) {
        return clubsRepository.existsById(clubName);
    }
}