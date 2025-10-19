package com.example.project.service;

import lombok.AllArgsConstructor;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Clubs;
import com.example.project.model.Equipment;
import com.example.project.model.Gyms;
import com.example.project.model.News;
import com.example.project.model.StaffSchedule;
import com.example.project.repository.ClubsRepository;

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

    public Set<Equipment> getSetOfEquipments(String clubsId, int gymId) {
        Set<Gyms> gyms = getSetOfGyms(clubsId);
        for (Gyms gym : gyms) {
            if (gym.getIdGym() == gymId) {
                return new HashSet<>(gym.getEquipment());
            }
        }
        return Collections.emptySet();
    }

    public String getEquipmentTypeName(String clubsId, int gymId, int equipmentId) {
        Set<Equipment> equipments = getSetOfEquipments(clubsId, gymId);
        for (Equipment equipment : equipments) {
            if (equipment.getIdEquipment() == equipmentId) {
                return equipment.getEquipmentType().getTypeName();
            }
        }
        return null;
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
}