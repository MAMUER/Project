package com.example.project.service;

import lombok.AllArgsConstructor;

import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Staff;
import com.example.project.model.StaffSchedule;
import com.example.project.model.Accounts.StaffAccounts;
import com.example.project.repository.StaffRepository;

@Service
@AllArgsConstructor
public class StaffService {
    private final StaffRepository staffRepository;

    @SuppressWarnings("null")
    public Staff getStaff(Integer staffId) {
        return staffRepository.findById(staffId).orElse(null);
    }

    public List<Staff> getAllStaff() {
        return staffRepository.findAll();
    }

    public Set<Staff> getStaffByPosition(Integer positionId) {
        return staffRepository.findByPositionIdPosition(positionId);
    }

    public Set<Staff> searchStaffByName(String firstName, String secondName) {
        if (firstName != null && secondName != null) {
            return staffRepository.findByFirstNameContainingAndSecondNameContaining(firstName, secondName);
        } else if (firstName != null) {
            return staffRepository.findByFirstNameContaining(firstName);
        } else if (secondName != null) {
            return staffRepository.findBySecondNameContaining(secondName);
        }
        return Collections.emptySet();
    }

    public List<StaffSchedule> getStaffSchedule(int staffId) {
        Staff staff = staffRepository.findById(staffId).orElse(null);
        return staff != null ? staff.getStaffSchedules() : Collections.emptyList();
    }

    public StaffAccounts getStaffAccount(int staffId) {
        Staff staff = staffRepository.findById(staffId).orElse(null);
        return staff != null ? staff.getStaffAccount() : null;
    }

    public String getPositionName(int staffId) {
        Staff staff = staffRepository.findById(staffId).orElse(null);
        return staff != null && staff.getPosition() != null ? staff.getPosition().getRoleName() : null;
    }

    public String getPhotoUrl(Integer staffId) {
        StaffAccounts staffAccount = getStaffAccount(staffId);
        try {
            return staffAccount.getStaffPhoto().getImageUrl();
        } catch (Exception e) {
            return "https://i.postimg.cc/Wbznd0qn/1674365371-3-34.jpg";
        }
    }

    public List<StaffSchedule> getSetOfStaffSchedule(Integer staffId) {
        Staff staff = staffRepository.findByIdWithSchedules(staffId).orElse(null);
        if (staff != null) {
            List<StaffSchedule> staffSchedule = staff.getStaffSchedules();
            staffSchedule.sort(Comparator.comparing(StaffSchedule::getDate));
            return staffSchedule;
        }
        return Collections.emptyList();
    }

    @SuppressWarnings("null")
    public Staff saveStaff(Staff staff) {
        return staffRepository.save(staff);
    }
}
