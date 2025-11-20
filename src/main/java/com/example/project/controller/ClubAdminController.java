package com.example.project.controller;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.*;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;

import com.example.project.model.Clubs;
import com.example.project.service.ClubCapabilityService;
import com.example.project.service.ClubsService;
import com.example.project.service.CustomUserDetailsService;

import java.util.Map;

@Controller
@AllArgsConstructor
@RequestMapping("/admin/clubs")
public class ClubAdminController {

    private final ClubsService clubsService;
    private final ClubCapabilityService clubCapabilityService;
    private final CustomUserDetailsService userDetailsService;

    private boolean isStaffUser() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        return authentication != null &&
                authentication.getAuthorities().stream()
                        .anyMatch(grantedAuthority -> grantedAuthority.getAuthority().equals("STAFF"));
    }

    private Integer getCurrentStaffId() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        if (authentication != null && authentication.getPrincipal() instanceof UserDetails) {
            String username = ((UserDetails) authentication.getPrincipal()).getUsername();
            return userDetailsService.getUserId(username);
        }
        return null;
    }

    @GetMapping
    @PreAuthorize("hasAuthority('STAFF')")
    public String clubsAdmin(Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        model.addAttribute("clubs", clubsService.getAllClubs());
        model.addAttribute("staffId", getCurrentStaffId()); // ДОБАВЛЕНО
        return "admin/admin-clubs";
    }

    @GetMapping("/{clubName}/capabilities")
    @PreAuthorize("hasAuthority('STAFF')")
    public String clubCapabilities(@PathVariable String clubName, Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Clubs club = clubsService.getClub(clubName);
        Map<String, Object> capabilities = clubCapabilityService.analyzeClubCapabilities(clubName);
        Map<String, Object> recommendations = clubCapabilityService.getClubImprovementRecommendations(clubName);

        model.addAttribute("club", club);
        model.addAttribute("capabilities", capabilities);
        model.addAttribute("recommendations", recommendations);
        model.addAttribute("staffId", getCurrentStaffId()); // ДОБАВЛЕНО

        return "admin/club-capabilities";
    }

    @GetMapping("/{clubName}/equipment-page")
    @PreAuthorize("hasAuthority('STAFF')")
    public String clubEquipmentPage(@PathVariable String clubName, Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Map<String, Integer> equipmentSummary = clubCapabilityService.getClubEquipmentSummary(clubName);

        model.addAttribute("clubName", clubName);
        model.addAttribute("equipmentSummary", equipmentSummary);
        model.addAttribute("staffId", getCurrentStaffId()); // ДОБАВЛЕНО

        return "admin/equipment";
    }

    // Остальные методы остаются без изменений, но нужно добавить staffId везде где есть Model
    @GetMapping("/create")
    @PreAuthorize("hasAuthority('STAFF')")
    public String createClubForm(Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        model.addAttribute("club", new Clubs());
        model.addAttribute("staffId", getCurrentStaffId()); // ДОБАВЛЕНО
        return "admin/club-create";
    }

    @GetMapping("/{clubName}/edit")
    @PreAuthorize("hasAuthority('STAFF')")
    public String editClubForm(@PathVariable String clubName, Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Clubs club = clubsService.getClub(clubName);
        model.addAttribute("club", club);
        model.addAttribute("staffId", getCurrentStaffId()); // ДОБАВЛЕНО
        return "admin/club-edit";
    }

    // Добавить в ClubAdminController новый метод
    @GetMapping("/{clubName}/manage-equipment")
    @PreAuthorize("hasAuthority('STAFF')")
    public String manageClubEquipment(@PathVariable String clubName, Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        return "redirect:/admin/equipment/" + clubName;
    }

    // Остальные методы без изменений...
    @GetMapping("/{clubName}/equipment")
    @ResponseBody
    @PreAuthorize("hasAuthority('STAFF')")
    public Object getClubEquipment(@PathVariable String clubName) {
        if (!isStaffUser()) {
            return Map.of("error", "Access denied");
        }
        return clubCapabilityService.getClubEquipmentSummary(clubName);
    }

    @GetMapping("/{clubName}/recommendations")
    @ResponseBody
    @PreAuthorize("hasAuthority('STAFF')")
    public Map<String, Object> getClubRecommendations(@PathVariable String clubName) {
        if (!isStaffUser()) {
            return Map.of("error", "Access denied");
        }
        return clubCapabilityService.getClubImprovementRecommendations(clubName);
    }

    @PostMapping("/create")
    @PreAuthorize("hasAuthority('STAFF')")
    public String createClub(@ModelAttribute Clubs club) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.saveClub(club);
        return "redirect:/admin/clubs";
    }

    @PostMapping("/{clubName}/edit")
    @PreAuthorize("hasAuthority('STAFF')")
    public String updateClub(@PathVariable String clubName, @ModelAttribute Clubs club) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.updateClub(clubName, club);
        return "redirect:/admin/clubs";
    }

    @PostMapping("/{clubName}/delete")
    @PreAuthorize("hasAuthority('STAFF')")
    public String deleteClub(@PathVariable String clubName) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.deleteClub(clubName);
        return "redirect:/admin/clubs";
    }
}