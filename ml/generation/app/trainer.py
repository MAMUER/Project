import numpy as np
import pandas as pd
from sklearn.preprocessing import LabelEncoder, StandardScaler
from gan_model import TrainingProgramGAN
import joblib
import os

class ProgramTrainer:
    def __init__(self):
        self.gan = None
        self.scaler = StandardScaler()
        self.label_encoder = LabelEncoder()
        self.classes = ['cardio', 'strength', 'flexibility', 'recovery', 'hiit', 'endurance']
        
    def load_dataset(self, filepath):
        """Загрузка датасета программ тренировок"""
        if os.path.exists(filepath):
            df = pd.read_csv(filepath)
            return df
        return None
    
    def prepare_training_data(self, programs_df):
        """Подготовка данных для обучения GAN"""
        # Кодирование классов
        class_encoded = self.label_encoder.fit_transform(programs_df['training_class'])
        
        # One-hot encoding для классов (6 классов)
        class_onehot = np.zeros((len(programs_df), len(self.classes)))
        class_onehot[np.arange(len(programs_df)), class_encoded] = 1
        
        # Кодирование других условий
        conditions = []
        for _, row in programs_df.iterrows():
            condition_vec = np.zeros(20)
            
            # Класс тренировки (6)
            condition_vec[:6] = class_onehot[_]
            
            # Противопоказания (5)
            contraindications = ['heart_issues', 'joint_problems', 'hypertension', 'asthma', 'diabetes']
            contra = eval(row.get('contraindications', '[]'))
            for i, cond in enumerate(contraindications):
                if cond in contra:
                    condition_vec[6 + i] = 1
            
            # Цели (5)
            goals = ['weight_loss', 'muscle_gain', 'endurance', 'flexibility', 'rehabilitation']
            user_goals = eval(row.get('goals', '[]'))
            for i, goal in enumerate(goals):
                if goal in user_goals:
                    condition_vec[11 + i] = 1
            
            # Уровень подготовки
            level_map = {'beginner': 0, 'intermediate': 0.5, 'advanced': 1}
            condition_vec[16] = level_map.get(row.get('fitness_level', 'beginner'), 0)
            
            # Возрастная группа
            age_map = {'young': 0, 'adult': 0.33, 'senior': 0.66, 'elderly': 1}
            condition_vec[17] = age_map.get(row.get('age_group', 'adult'), 0.33)
            
            # Пол
            condition_vec[18] = 1 if row.get('gender') == 'male' else 0
            
            # Наличие травм
            condition_vec[19] = 1 if row.get('has_injury', False) else 0
            
            conditions.append(condition_vec)
        
        conditions = np.array(conditions)
        
        # Программы тренировок (210 параметров)
        programs = []
        for _, row in programs_df.iterrows():
            program = eval(row.get('program_data', '[]'))
            programs.append(program[:210] if len(program) >= 210 else program + [0]*(210-len(program)))
        
        programs = self.scaler.fit_transform(np.array(programs))
        
        return programs, conditions
    
    def train(self, programs, conditions, latent_dim=100, epochs=10000, batch_size=32):
        """Обучение GAN"""
        self.gan = TrainingProgramGAN(latent_dim=latent_dim)
        self.gan.train(programs, conditions, epochs=epochs, batch_size=batch_size)
    
    def generate_programs(self, conditions, num_programs=1):
        """Генерация программ"""
        programs_raw = self.gan.generate(conditions, num_programs)
        programs = self.scaler.inverse_transform(programs_raw)
        return programs
    
    def save_models(self, path='models/'):
        """Сохранение моделей"""
        os.makedirs(path, exist_ok=True)
        self.gan.save(f'{path}gan')
        joblib.dump(self.scaler, f'{path}scaler.pkl')
        joblib.dump(self.label_encoder, f'{path}label_encoder.pkl')
    
    def load_models(self, path='models/'):
        """Загрузка моделей"""
        self.gan = TrainingProgramGAN()
        self.gan.load(f'{path}gan')
        self.scaler = joblib.load(f'{path}scaler.pkl')
        self.label_encoder = joblib.load(f'{path}label_encoder.pkl')