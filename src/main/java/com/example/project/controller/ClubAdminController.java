package com.example.project.controller;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.*;

import com.example.project.model.Clubs;
import com.example.project.service.ClubCapabilityService;
import com.example.project.service.ClubsService;

import java.util.Map;

@Controller
@AllArgsConstructor
@RequestMapping("/admin/clubs")
public class ClubAdminController {
    
    private final ClubsService clubsService;
    private final ClubCapabilityService clubCapabilityService;

    @GetMapping
    public String clubsAdmin(Model model) {
        model.addAttribute("clubs", clubsService.getAllClubs());
        return "admin-clubs";
    }

    @GetMapping("/{clubName}/capabilities")
    public String clubCapabilities(@PathVariable String clubName, Model model) {
        Clubs club = clubsService.getClub(clubName);
        Map<String, Object> capabilities = clubCapabilityService.analyzeClubCapabilities(clubName);
        Map<String, Object> recommendations = clubCapabilityService.getClubImprovementRecommendations(clubName);
        
        model.addAttribute("club", club);
        model.addAttribute("capabilities", capabilities);
        model.addAttribute("recommendations", recommendations);
        
        return "club-capabilities";
    }

    @GetMapping("/{clubName}/equipment")
    @ResponseBody
    public Map<String, Integer> getClubEquipment(@PathVariable String clubName) {
        return clubCapabilityService.getClubEquipmentSummary(clubName);
    }

    @GetMapping("/{clubName}/recommendations")
    @ResponseBody
    public Map<String, Object> getClubRecommendations(@PathVariable String clubName) {
        return clubCapabilityService.getClubImprovementRecommendations(clubName);
    }
}