package com.example.project.controller;

import java.net.URLDecoder;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;

import com.example.project.model.Clubs;
import com.example.project.model.Equipment;
import com.example.project.model.EquipmentType;
import com.example.project.service.ClubsService;
import com.example.project.service.CustomUserDetailsService;
import com.example.project.service.EquipmentService;
import com.example.project.service.EquipmentTypeService;
import com.example.project.service.StaffScheduleService;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.AllArgsConstructor;

@Controller
@AllArgsConstructor
@RequestMapping("/admin/equipment")
@Tag(name = "Управление оборудованием", description = "API для управления оборудованием в фитнес-клубах (только для сотрудников STAFF)")
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
    @Operation(summary = "Панель управления оборудованием", description = "Отображает панель управления оборудованием с доступными клубами для текущего сотрудника STAFF")
    @ApiResponse(responseCode = "200", description = "Панель управления успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен - требуется роль STAFF")
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
        model.addAttribute("staffId", staffId);
        return "admin/equipment-management";
    }

    // Страница редактирования оборудования конкретного клуба
    @GetMapping("/{clubName}")
    @PreAuthorize("hasAuthority('STAFF')")
    @Transactional
    @Operation(summary = "Редактирование оборудования клуба", description = "Отображает страницу для редактирования оборудования в указанном клубе")
    @ApiResponse(responseCode = "200", description = "Страница редактирования оборудования успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен - нет прав доступа к этому клубу")
    public String editClubEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'", schema = @Schema(type = "string", format = "uri")) @PathVariable String clubName,
            Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }

        String actualClubName;

        try {
            actualClubName = URLDecoder.decode(clubName, StandardCharsets.UTF_8);
        } catch (Exception e) {
            actualClubName = clubName;
        }

        final String finalClubName = actualClubName;

        if (!hasAccessToClub(finalClubName)) {
            return "redirect:/error/403";
        }

        Clubs club = clubsService.getClub(finalClubName);
        List<EquipmentType> equipmentTypes = equipmentTypeService.getAllEquipmentTypes();

        List<Equipment> allEquipment = equipmentService.getAllEquipment();
        List<Equipment> clubEquipment = allEquipment.stream()
                .filter(e -> e.getClub() != null && finalClubName.equals(e.getClub().getClubName()))
                .peek(e -> {
                    if (e.getEquipmentType() != null) {
                        e.getEquipmentType().getTypeName();
                    }
                })
                .sorted(Comparator.comparing(e -> e.getEquipmentType().getTypeName()))
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
    @Operation(summary = "Обновление количества оборудования", description = "Обновляет количество единиц указанного оборудования в клубе")
    @ApiResponse(responseCode = "200", description = "Количество оборудования успешно обновлено")
    @ApiResponse(responseCode = "400", description = "Ошибка обновления оборудования")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String updateEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            @Parameter(description = "ID оборудования", example = "1", required = true) @RequestParam Integer equipmentId,
            @Parameter(description = "Новое количество", example = "10", required = true) @RequestParam Integer quantity,
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
    @Operation(summary = "Добавление нового оборудования", description = "Добавляет новое оборудование в указанный клуб")
    @ApiResponse(responseCode = "200", description = "Оборудование успешно добавлено")
    @ApiResponse(responseCode = "400", description = "Ошибка добавления оборудования")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String addEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            @Parameter(description = "Данные нового оборудования") @ModelAttribute Equipment newEquipment,
            @Parameter(description = "ID типа оборудования", example = "1", required = true) @RequestParam Integer equipmentTypeId,
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
                equipment.setQuantity(Optional.ofNullable(newEquipment.getQuantity()).orElse(1));

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
    @Operation(summary = "Удаление оборудования", description = "Удаляет оборудование из указанного клуба")
    @ApiResponse(responseCode = "200", description = "Оборудование успешно удалено")
    @ApiResponse(responseCode = "400", description = "Ошибка удаления оборудования")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String deleteEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            @Parameter(description = "ID оборудования", example = "1", required = true) @RequestParam Integer equipmentId,
            Model model) {
        if (!isStaffUser() || !hasAccessToClub(clubName)) {
            return "redirect:/error/403";
        }

        try {
            Equipment equipment = equipmentService.getEquipment(equipmentId);
            if (equipment != null && equipment.getClub() != null &&
                    clubName.equals(equipment.getClub().getClubName())) {

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