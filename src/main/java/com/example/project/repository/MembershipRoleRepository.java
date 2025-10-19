package com.example.project.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.project.model.MembershipRole;

@Repository
public interface MembershipRoleRepository extends JpaRepository<MembershipRole, Integer> {

    Optional<MembershipRole> findByRoleName(String roleName);
}