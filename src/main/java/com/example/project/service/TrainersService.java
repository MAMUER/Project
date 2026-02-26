package com.example.project.service;

import lombok.AllArgsConstructor;

import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Set;

import org.springframework.stereotype.Service;

import com.example.project.model.Trainers;
import com.example.project.model.TrainingSchedule;
import com.example.project.model.Accounts.TrainersAccounts;
import com.example.project.repository.TrainersRepository;
@Service
@AllArgsConstructor
public class TrainersService {
    private final TrainersRepository trainersRepository;

    @SuppressWarnings("null")
    public Trainers getTrainer(Integer trainerId) {
        return trainersRepository.findById(trainerId).orElse(null);
    }

    public List<Trainers> getAllTrainers() {
        return trainersRepository.findAll();
    }

    public Set<Trainers> getTrainersBySpeciality(String speciality) {
        return trainersRepository.findBySpecialityContaining(speciality);
    }

    public Set<Trainers> getTrainersByExperience(int minExperience) {
        return trainersRepository.findByExperienceGreaterThanEqual(minExperience);
    }

    public String getTrainerFirstName(Integer trainerId) {
        @SuppressWarnings("null")
        Trainers trainer = trainersRepository.findById(trainerId).orElse(null);
        return trainer != null ? trainer.getFirstName() : null;
    }

    public String getTrainerSecondName(Integer trainerId) {
        @SuppressWarnings("null")
        Trainers trainer = trainersRepository.findById(trainerId).orElse(null);
        return trainer != null ? trainer.getSecondName() : null;
    }

    public List<TrainingSchedule> getTrainingSchedules(int trainerId) {
        Trainers trainer = trainersRepository.findById(trainerId).orElse(null);
        return trainer != null ? trainer.getTrainingSchedules() : Collections.emptyList();
    }

    public TrainersAccounts getTrainerAccount(int trainerId) {
        Trainers trainer = trainersRepository.findById(trainerId).orElse(null);
        return trainer != null ? trainer.getTrainersAccount() : null;
    }

    public List<TrainingSchedule> getSetOfTrainingSchedule(int trainerId) {
        Trainers trainer = trainersRepository.findById(trainerId).orElse(null);
        if (trainer != null) {
            List<TrainingSchedule> trainerTrainings = trainer.getTrainingSchedules();
            trainerTrainings.sort(Comparator.comparing(TrainingSchedule::getSessionDate));
            return trainerTrainings;
        }
        return Collections.emptyList();
    }

    public String getPhotoUrl(int trainersId) {
        TrainersAccounts trainerAccount = getTrainerAccount(trainersId);
        try {
            return trainerAccount.getTrainersPhoto().getImageUrl();
        } catch (Exception e) {
            return "https://i.postimg.cc/Wbznd0qn/1674365371-3-34.jpg";
        }
    }

    @SuppressWarnings("null")
    public Trainers saveTrainer(Trainers trainer) {
        return trainersRepository.save(trainer);
    }

    @SuppressWarnings("null")
    public void deleteTrainer(Integer id) {
        trainersRepository.deleteById(id);
    }
}