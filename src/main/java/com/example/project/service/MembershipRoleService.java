package com.example.project.service;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.project.model.MembershipRole;
import com.example.project.repository.MembershipRoleRepository;

import lombok.AllArgsConstructor;

@Service
@AllArgsConstructor
public class MembershipRoleService {
    private final MembershipRoleRepository membershipRoleRepository;

    public MembershipRole getMembershipRole(Integer id) {
        return membershipRoleRepository.findById(id).orElse(null);
    }

    public MembershipRole getMembershipRoleByName(String roleName) {
        return membershipRoleRepository.findByRoleName(roleName).orElse(null);
    }

    public List<MembershipRole> getAllMembershipRoles() {
        return membershipRoleRepository.findAll();
    }

    public MembershipRole saveMembershipRole(MembershipRole role) {
        return membershipRoleRepository.save(role);
    }
}