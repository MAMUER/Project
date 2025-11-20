package com.example.project.controller;

import lombok.AllArgsConstructor;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.*;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;

import com.example.project.model.Clubs;
import com.example.project.model.Equipment;
import com.example.project.model.EquipmentType;
import com.example.project.service.ClubsService;
import com.example.project.service.EquipmentService;
import com.example.project.service.EquipmentTypeService;
import com.example.project.service.CustomUserDetailsService;
import com.example.project.service.StaffScheduleService;

import java.util.List;
import java.util.Map;
import java.net.URLDecoder;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Comparator;
import java.util.HashMap;

@Controller
@AllArgsConstructor
@RequestMapping("/admin/equipment")
public class EquipmentAdminController {

    private final ClubsService clubsService;
    private final EquipmentService equipmentService;
    private final EquipmentTypeService equipmentTypeService;
    private final CustomUserDetailsService userDetailsService;
    private final StaffScheduleService staffScheduleService;

    private boolean isStaffUser() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        return authentication != null &&
                authentication.getAuthorities().stream()
                        .anyMatch(grantedAuthority -> grantedAuthority.getAuthority().equals("STAFF"));
    }

    // ДОБАВИТЬ этот метод
    private Integer getCurrentStaffId() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        if (authentication != null && authentication.getPrincipal() instanceof UserDetails) {
            String username = ((UserDetails) authentication.getPrincipal()).getUsername();
            return userDetailsService.getUserId(username);
        }
        return null;
    }

    // Главная страница управления оборудованием
    @GetMapping
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    public String equipmentManagement(Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }

        // Получаем клубы, в которых работает текущий staff
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer staffId = userDetailsService.getUserId(username);

        // Получаем клубы через расписание staff
        var staffSchedules = staffScheduleService.getStaffScheduleByStaff(staffId);
        Map<String, Clubs> staffClubs = new HashMap<>();
        for (var schedule : staffSchedules) {
            Clubs club = schedule.getClub();
            staffClubs.put(club.getClubName(), club);
        }

        model.addAttribute("clubs", staffClubs.values());
        model.addAttribute("staffId", staffId); // ДОБАВИТЬ эту строку
        return "admin/equipment-management";
    }

    // Страница редактирования оборудования конкретного клуба
    @GetMapping("/{clubName}")
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    public String editClubEquipment(@PathVariable String clubName, Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }

        String actualClubName;

        try {
            // Пробуем декодировать название клуба
            actualClubName = URLDecoder.decode(clubName, StandardCharsets.UTF_8);
        } catch (Exception e) {
            // Если декодирование не удалось, используем оригинальное название
            actualClubName = clubName;
        }

        // Создаем effectively final копию для использования в лямбде
        final String finalClubName = actualClubName;

        // Проверяем, что staff имеет доступ к этому клубу
        if (!hasAccessToClub(finalClubName)) {
            return "redirect:/error/403";
        }

        Clubs club = clubsService.getClub(finalClubName);
        List<EquipmentType> equipmentTypes = equipmentTypeService.getAllEquipmentTypes();

        // Фильтруем оборудование по клубу с инициализацией связей И СОРТИРОВКОЙ
        List<Equipment> allEquipment = equipmentService.getAllEquipment();
        List<Equipment> clubEquipment = allEquipment.stream()
                .filter(e -> e.getClub() != null && finalClubName.equals(e.getClub().getClubName()))
                .peek(e -> {
                    // Принудительно инициализируем ленивые связи
                    if (e.getEquipmentType() != null) {
                        e.getEquipmentType().getTypeName();
                    }
                })
                .sorted(Comparator.comparing(e -> e.getEquipmentType().getTypeName())) // ДОБАВИТЬ СОРТИРОВКУ
                .toList();

        model.addAttribute("club", club);
        model.addAttribute("clubEquipment", clubEquipment);
        model.addAttribute("equipmentTypes", equipmentTypes);
        model.addAttribute("newEquipment", new Equipment());
        model.addAttribute("staffId", getCurrentStaffId());

        return "admin/equipment-edit";
    }

    // Обновление количества оборудования
    @PostMapping("/{clubName}/update")
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    public String updateEquipment(@PathVariable String clubName,
            @RequestParam Integer equipmentId,
            @RequestParam Integer quantity,
            Model model) {
        if (!isStaffUser() || !hasAccessToClub(clubName)) {
            return "redirect:/error/403";
        }

        try {
            Equipment equipment = equipmentService.getEquipment(equipmentId);
            if (equipment != null && equipment.getClub() != null &&
                    clubName.equals(equipment.getClub().getClubName())) {
                equipment.setQuantity(quantity);
                equipmentService.saveEquipment(equipment);
            }

            // Кодируем кириллическое название клуба в URL
            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?success=Equipment updated";
        } catch (Exception e) {
            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?error=Error updating equipment";
        }
    }

    // Добавление нового оборудования
    @PostMapping("/{clubName}/add")
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    public String addEquipment(@PathVariable String clubName,
            @ModelAttribute Equipment newEquipment,
            @RequestParam Integer equipmentTypeId,
            Model model) {
        if (!isStaffUser() || !hasAccessToClub(clubName)) {
            return "redirect:/error/403";
        }

        try {
            Clubs club = clubsService.getClub(clubName);
            EquipmentType equipmentType = equipmentTypeService.getEquipmentType(equipmentTypeId);

            if (club != null && equipmentType != null) {
                Equipment equipment = new Equipment();
                equipment.setEquipmentType(equipmentType);
                equipment.setClub(club);
                equipment.setQuantity(newEquipment.getQuantity() != null ? newEquipment.getQuantity() : 1);

                equipmentService.saveEquipment(equipment);
            }

            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?success=Equipment added";
        } catch (Exception e) {
            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?error=Error adding equipment";
        }
    }

    // Удаление оборудования
    @PostMapping("/{clubName}/delete")
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    public String deleteEquipment(@PathVariable String clubName,
            @RequestParam Integer equipmentId,
            Model model) {
        if (!isStaffUser() || !hasAccessToClub(clubName)) {
            return "redirect:/error/403";
        }

        try {
            Equipment equipment = equipmentService.getEquipment(equipmentId);
            if (equipment != null && equipment.getClub() != null &&
                    clubName.equals(equipment.getClub().getClubName())) {

                // Реальное удаление оборудования
                equipmentService.deleteEquipment(equipmentId);
            }

            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?success=Equipment deleted";
        } catch (Exception e) {
            String encodedClubName = URLEncoder.encode(clubName, StandardCharsets.UTF_8);
            return "redirect:/admin/equipment/" + encodedClubName + "?error=Error deleting equipment";
        }
    }

    // Проверка доступа staff к клубу
    private boolean hasAccessToClub(String clubName) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer staffId = userDetailsService.getUserId(username);

        var staffSchedules = staffScheduleService.getStaffScheduleByStaff(staffId);
        return staffSchedules.stream()
                .anyMatch(schedule -> schedule.getClub().getClubName().equals(clubName));
    }
}