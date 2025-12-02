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

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import io.swagger.v3.oas.annotations.media.Content;
import java.util.Map;

@Controller
@AllArgsConstructor
@RequestMapping("/admin/clubs")
@Tag(name = "Управление клубами", description = "API для управления фитнес-клубами и их возможностями (только для сотрудников STAFF)")
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
    @Operation(summary = "Панель управления клубами", description = "Отображает список всех фитнес-клубов для управления")
    @ApiResponse(responseCode = "200", description = "Список клубов успешно загружен")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен - требуется роль STAFF")
    public String clubsAdmin(Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        model.addAttribute("clubs", clubsService.getAllClubs());
        model.addAttribute("staffId", getCurrentStaffId());
        return "admin/admin-clubs";
    }

    @GetMapping("/{clubName}/capabilities")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Анализ возможностей клуба", description = "Отображает детальный анализ возможностей и оснащенности указанного клуба")
    @ApiResponse(responseCode = "200", description = "Анализ возможностей успешно загружен")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String clubCapabilities(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Clubs club = clubsService.getClub(clubName);
        Map<String, Object> capabilities = clubCapabilityService.analyzeClubCapabilities(clubName);
        Map<String, Object> recommendations = clubCapabilityService.getClubImprovementRecommendations(clubName);

        model.addAttribute("club", club);
        model.addAttribute("capabilities", capabilities);
        model.addAttribute("recommendations", recommendations);
        model.addAttribute("staffId", getCurrentStaffId());

        return "admin/club-capabilities";
    }

    @GetMapping("/{clubName}/equipment-page")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Сводка по оборудованию клуба", description = "Отображает сводную информацию об оборудовании в указанном клубе")
    @ApiResponse(responseCode = "200", description = "Сводка по оборудованию успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String clubEquipmentPage(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Map<String, Integer> equipmentSummary = clubCapabilityService.getClubEquipmentSummary(clubName);

        model.addAttribute("clubName", clubName);
        model.addAttribute("equipmentSummary", equipmentSummary);
        model.addAttribute("staffId", getCurrentStaffId());

        return "admin/equipment";
    }

    @GetMapping("/create")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Форма создания клуба", description = "Отображает форму для создания нового фитнес-клуба")
    @ApiResponse(responseCode = "200", description = "Форма создания клуба успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String createClubForm(Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        model.addAttribute("club", new Clubs());
        model.addAttribute("staffId", getCurrentStaffId());
        return "admin/club-create";
    }

    @GetMapping("/{clubName}/edit")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Форма редактирования клуба", description = "Отображает форму для редактирования информации о клубе")
    @ApiResponse(responseCode = "200", description = "Форма редактирования успешно загружена")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    @ApiResponse(responseCode = "404", description = "Клуб не найден")
    public String editClubForm(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        Clubs club = clubsService.getClub(clubName);
        model.addAttribute("club", club);
        model.addAttribute("staffId", getCurrentStaffId());
        return "admin/club-edit";
    }

    @GetMapping("/{clubName}/manage-equipment")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Перенаправление на управление оборудованием", description = "Перенаправляет на страницу управления оборудованием для указанного клуба")
    @ApiResponse(responseCode = "302", description = "Перенаправление на страницу оборудования")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String manageClubEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            Model model) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        return "redirect:/admin/equipment/" + clubName;
    }

    @GetMapping("/{clubName}/equipment")
    @ResponseBody
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Получить оборудование клуба (API)", description = "Возвращает JSON с оборудованием указанного клуба")
    @ApiResponse(responseCode = "200", description = "Оборудование успешно получено", content = @Content(mediaType = "application/json"))
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public Object getClubEquipment(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName) {
        if (!isStaffUser()) {
            return Map.of("error", "Access denied");
        }
        return clubCapabilityService.getClubEquipmentSummary(clubName);
    }

    @GetMapping("/{clubName}/recommendations")
    @ResponseBody
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Получить рекомендации по улучшению клуба (API)", description = "Возвращает JSON с рекомендациями по улучшению указанного клуба")
    @ApiResponse(responseCode = "200", description = "Рекомендации успешно получены", content = @Content(mediaType = "application/json"))
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public Map<String, Object> getClubRecommendations(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName) {
        if (!isStaffUser()) {
            return Map.of("error", "Access denied");
        }
        return clubCapabilityService.getClubImprovementRecommendations(clubName);
    }

    @PostMapping("/create")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Создать новый клуб", description = "Создает новый фитнес-клуб с указанными параметрами")
    @ApiResponse(responseCode = "200", description = "Клуб успешно создан")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации данных")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    public String createClub(
            @Parameter(description = "Данные нового клуба") @ModelAttribute Clubs club) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.saveClub(club);
        return "redirect:/admin/clubs";
    }

    @PostMapping("/{clubName}/edit")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Обновить информацию о клубе", description = "Обновляет информацию о существующем фитнес-клубе")
    @ApiResponse(responseCode = "200", description = "Информация о клубе успешно обновлена")
    @ApiResponse(responseCode = "400", description = "Ошибка обновления данных")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    @ApiResponse(responseCode = "404", description = "Клуб не найден")
    public String updateClub(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName,
            @Parameter(description = "Обновленные данные клуба") @ModelAttribute Clubs club) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.updateClub(clubName, club);
        return "redirect:/admin/clubs";
    }

    @PostMapping("/{clubName}/delete")
    @PreAuthorize("hasAuthority('STAFF')")
    @Operation(summary = "Удалить клуб", description = "Удаляет указанный фитнес-клуб из системы")
    @ApiResponse(responseCode = "200", description = "Клуб успешно удален")
    @ApiResponse(responseCode = "403", description = "Доступ запрещен")
    @ApiResponse(responseCode = "404", description = "Клуб не найден")
    public String deleteClub(
            @Parameter(description = "Название клуба", example = "Фитнес Центр 'Энергия'") @PathVariable String clubName) {
        if (!isStaffUser()) {
            return "redirect:/error/403";
        }
        clubsService.deleteClub(clubName);
        return "redirect:/admin/clubs";
    }
}