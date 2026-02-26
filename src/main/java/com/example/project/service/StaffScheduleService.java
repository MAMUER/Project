package com.example.project.service;

import lombok.AllArgsConstructor;

import java.time.LocalDate;
import java.util.*;

import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.project.model.Staff;
import com.example.project.model.StaffSchedule;
import com.example.project.repository.StaffScheduleRepository;

@Service
@AllArgsConstructor
@Transactional
public class StaffScheduleService {
    private final StaffScheduleRepository staffScheduleRepository;
    private final CustomUserDetailsService userDetailsService;
    private final StaffService staffService;

    @SuppressWarnings("null")
    public StaffSchedule getStaffSchedule(Integer id) {
        return staffScheduleRepository.findById(id).orElse(null);
    }

    public List<StaffSchedule> getAllStaffSchedules() {
        return staffScheduleRepository.findAll();
    }

    public List<StaffSchedule> getStaffScheduleByStaff(Integer staffId) {
        return staffScheduleRepository.findByStaffIdStaff(staffId);
    }

    public Set<StaffSchedule> getStaffScheduleByClub(String clubName) {
        return staffScheduleRepository.findByClubClubName(clubName);
    }

    public Set<StaffSchedule> getStaffScheduleByDateRange(LocalDate startDate, LocalDate endDate) {
        return staffScheduleRepository.findByDateBetween(startDate, endDate);
    }

    public Set<StaffSchedule> getStaffScheduleByShift(Integer shift) {
        return staffScheduleRepository.findByShift(shift);
    }

    public Set<StaffSchedule> getStaffScheduleByStaffAndDateRange(Integer staffId, LocalDate startDate,
            LocalDate endDate) {
        return staffScheduleRepository.findByStaffIdStaffAndDateBetween(staffId, startDate, endDate);
    }

    public List<Event> staffScheduleToEvents(List<StaffSchedule> staffScheduleSet) {
        List<Event> eventsSet = new ArrayList<>();

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer staffId = userDetailsService.getUserId(username);
        Staff staff = staffService.getStaff(staffId);

        // Получаем расписание текущего сотрудника для сравнения
        Set<Integer> currentStaffScheduleIds = new HashSet<>();
        if (staff != null && staff.getStaffSchedules() != null) {
            staff.getStaffSchedules().forEach(schedule -> currentStaffScheduleIds.add(schedule.getIdSchedule()));
        }

        for (StaffSchedule work : staffScheduleSet) {
            String color;
            if (currentStaffScheduleIds.contains(work.getIdSchedule())) {
                color = "#3e4684"; // Собственное расписание
            } else {
                color = "#b2b4d4"; // Расписание других сотрудников
            }

            String staffName = work.getStaff().getFirstName() + " " +
                    work.getStaff().getSecondName() + " - смена " + work.getShift();

            String startDateTime = formatDateTime(work.getDate(), work.getShift(), true);
            String endDateTime = formatDateTime(work.getDate(), work.getShift(), false);

            int workId = work.getIdSchedule();

            eventsSet.add(new Event(staffName, startDateTime, endDateTime, workId, color));
        }
        return eventsSet;
    }

    /**
     * Форматирует дату и время для календаря
     */
    private String formatDateTime(LocalDate date, Integer shift, boolean isStart) {
        String dateStr = date.toString();
        return switch (shift) {
            case 1 -> isStart ? dateStr + "T06:00:00" : dateStr + "T12:00:00";
            case 2 -> isStart ? dateStr + "T12:00:00" : dateStr + "T18:00:00";
            case 3 -> isStart ? dateStr + "T18:00:00" : dateStr + "T23:59:00";
            default -> isStart ? dateStr + "T00:00:00" : dateStr + "T23:59:00";
        };
    }

    /**
     * Альтернативный метод с использованием Stream API
     */
    public Set<Event> convertToEvents(Set<StaffSchedule> staffScheduleSet) {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentStaffId = userDetailsService.getUserId(username);

        return staffScheduleSet.stream()
                .map(schedule -> createEvent(schedule, currentStaffId))
                .collect(HashSet::new, HashSet::add, HashSet::addAll);
    }

    private Event createEvent(StaffSchedule schedule, Integer currentStaffId) {
        boolean isOwnSchedule = Integer.valueOf(schedule.getStaff().getIdStaff()).equals(currentStaffId);
        String color = isOwnSchedule ? "#3e4684" : "#b2b4d4";

        String title = schedule.getStaff().getFirstName() + " " +
                schedule.getStaff().getSecondName() + " - смена " + schedule.getShift();

        String start = formatDateTime(schedule.getDate(), schedule.getShift(), true);
        String end = formatDateTime(schedule.getDate(), schedule.getShift(), false);

        return new Event(title, start, end, schedule.getIdSchedule(), color);
    }

    @SuppressWarnings("null")
    public StaffSchedule saveStaffSchedule(StaffSchedule schedule) {
        return staffScheduleRepository.save(schedule);
    }

    @SuppressWarnings("null")
    public List<StaffSchedule> saveAllStaffSchedules(Set<StaffSchedule> schedules) {
        return staffScheduleRepository.saveAll(schedules);
    }

    @SuppressWarnings("null")
    public void deleteStaffSchedule(Integer id) {
        staffScheduleRepository.deleteById(id);
    }

    @SuppressWarnings("null")
    public void deleteStaffSchedulesByStaff(Integer staffId) {
        List<StaffSchedule> schedules = staffScheduleRepository.findByStaffIdStaff(staffId);
        staffScheduleRepository.deleteAll(schedules);
    }

    /**
     * Получить расписание для текущего аутентифицированного сотрудника
     */
    public List<StaffSchedule> getCurrentStaffSchedule() {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer staffId = userDetailsService.getUserId(username);
        return getStaffScheduleByStaff(staffId);
    }

    /**
     * Получить события для текущего сотрудника
     */
    public List<Event> getCurrentStaffEvents() {
        List<StaffSchedule> schedules = getCurrentStaffSchedule();
        return staffScheduleToEvents(schedules);
    }

    /**
     * Проверить, принадлежит ли расписание текущему сотруднику
     */
    public boolean isScheduleBelongsToCurrentStaff(Integer scheduleId) {
        StaffSchedule schedule = getStaffSchedule(scheduleId);
        if (schedule == null)
            return false;

        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        String username = ((UserDetails) principal).getUsername();
        Integer currentStaffId = userDetailsService.getUserId(username);

        return Integer.valueOf(schedule.getStaff().getIdStaff()).equals(currentStaffId);
    }
}