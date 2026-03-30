"""
Training script for GAN-based Training Plan Generator
Generates personalized training plans based on training class and user profile
"""

import os
import json
import numpy as np
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import layers, models
from datetime import datetime

# Suppress TF warnings
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'
tf.get_logger().setLevel('ERROR')

# Training plan templates (from document)
TRAINING_TEMPLATES = {
    'recovery': {
        'duration_range': (20, 45),
        'intensity_range': (0.3, 0.5),
        'exercises': ['walking', 'yoga', 'stretching', 'light_swimming', 'mobility'],
        'rest_ratio': 0.7,
        'name_ru': 'Восстановление'
    },
    'endurance_e1e2': {
        'duration_range': (45, 90),
        'intensity_range': (0.5, 0.7),
        'exercises': ['running', 'cycling', 'swimming', 'rowing', 'hiking'],
        'rest_ratio': 0.4,
        'name_ru': 'Базовая выносливость'
    },
    'threshold_e3': {
        'duration_range': (30, 60),
        'intensity_range': (0.7, 0.85),
        'exercises': ['tempo_run', 'threshold_intervals', 'fartlek', 'critical_power'],
        'rest_ratio': 0.3,
        'name_ru': 'Пороговая выносливость'
    },
    'strength_hiit': {
        'duration_range': (20, 45),
        'intensity_range': (0.85, 1.0),
        'exercises': ['hiit', 'strength', 'sprints', 'crossfit', 'plyometrics'],
        'rest_ratio': 0.5,
        'name_ru': 'Силовая/HIIT'
    }
}


class TrainingPlanGAN:
    """GAN for generating training plans"""
    
    def __init__(self, latent_dim=32, user_profile_dim=10, training_class_dim=4, plan_dim=16):
        self.latent_dim = latent_dim
        self.user_profile_dim = user_profile_dim
        self.training_class_dim = training_class_dim
        self.plan_dim = plan_dim  # Output dimension
        
        self.generator = self.build_generator()
        self.discriminator = self.build_discriminator()
        self.gan = self.build_gan()
        
    def build_generator(self):
        """Build the generator network"""
        # Input layers
        input_latent = layers.Input(shape=(self.latent_dim,), name='latent_input')
        input_profile = layers.Input(shape=(self.user_profile_dim,), name='profile_input')
        input_class = layers.Input(shape=(self.training_class_dim,), name='class_input')
        
        # Concatenate all inputs
        combined = layers.concatenate([input_latent, input_profile, input_class])
        
        # Generator network
        x = layers.Dense(128, activation='relu')(combined)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(256, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(128, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        
        # Output layer - must match plan_dim
        output = layers.Dense(self.plan_dim, activation='sigmoid', name='plan_output')(x)
        
        model = models.Model(
            inputs=[input_latent, input_profile, input_class],
            outputs=output,
            name='generator'
        )
        
        return model
    
    def build_discriminator(self):
        """Build the discriminator network"""
        # Input must match generator output dimension
        input_plan = layers.Input(shape=(self.plan_dim,), name='plan_input')
        
        x = layers.Dense(128, activation='relu')(input_plan)
        x = layers.Dropout(0.3)(x)
        
        x = layers.Dense(64, activation='relu')(x)
        x = layers.Dropout(0.3)(x)
        
        x = layers.Dense(32, activation='relu')(x)
        
        output = layers.Dense(1, activation='sigmoid', name='discriminator_output')(x)
        
        model = models.Model(inputs=input_plan, outputs=output, name='discriminator')
        
        model.compile(
            optimizer=keras.optimizers.Adam(learning_rate=0.0002),
            loss='binary_crossentropy',
            metrics=['accuracy']
        )
        
        return model
    
    def build_gan(self):
        """Build the combined GAN"""
        self.discriminator.trainable = False
        
        # Input layers
        input_latent = layers.Input(shape=(self.latent_dim,), name='latent_input')
        input_profile = layers.Input(shape=(self.user_profile_dim,), name='profile_input')
        input_class = layers.Input(shape=(self.training_class_dim,), name='class_input')
        
        # Generate plan
        generated_plan = self.generator([input_latent, input_profile, input_class])
        
        # Discriminate
        validity = self.discriminator(generated_plan)
        
        model = models.Model(
            inputs=[input_latent, input_profile, input_class],
            outputs=validity,
            name='gan'
        )
        
        model.compile(
            optimizer=keras.optimizers.Adam(learning_rate=0.0002),
            loss='binary_crossentropy'
        )
        
        return model
    
    def generate_training_data(self, n_samples=2000):
        """Generate synthetic training plan data"""
        np.random.seed(42)
        
        plans = []
        
        for _ in range(n_samples):
            training_class = np.random.randint(0, 4)
            class_name = list(TRAINING_TEMPLATES.keys())[training_class]
            template = TRAINING_TEMPLATES[class_name]
            
            # Generate plan parameters (16 dimensions)
            duration = np.random.uniform(*template['duration_range']) / 100
            intensity = np.random.uniform(*template['intensity_range'])
            rest_ratio = np.random.uniform(max(0, template['rest_ratio'] - 0.1), 
                                          min(1, template['rest_ratio'] + 0.1))
            weekly_freq = np.random.uniform(2, 6) / 7
            
            # Exercise probabilities (5 exercises)
            exercise_probs = np.random.dirichlet(np.ones(5)) * 0.5
            
            # Additional parameters
            warmup = np.random.uniform(5, 15) / 100
            cooldown = np.random.uniform(5, 15) / 100
            progression = np.random.uniform(0.1, 0.3)
            
            # User adaptation factors (4 dimensions)
            age_factor = np.random.uniform(0.5, 1.0)
            fitness_factor = np.random.uniform(0.5, 1.0)
            health_factor = np.random.uniform(0.5, 1.0)
            goal_factor = np.random.uniform(0.5, 1.0)
            
            plan = np.array([
                duration, intensity, rest_ratio, weekly_freq,
                exercise_probs[0], exercise_probs[1], exercise_probs[2], 
                exercise_probs[3], exercise_probs[4],
                warmup, cooldown, progression,
                age_factor, fitness_factor, health_factor, goal_factor
            ])
            
            plans.append(plan)
        
        return np.array(plans)
    
    def train(self, epochs=500, batch_size=32, save_interval=50):
        """Train the GAN"""
        print("=" * 60)
        print("Starting GAN Training")
        print("=" * 60)
        
        # Generate training data
        print("\n[1/4] Generating training data...")
        real_plans = self.generate_training_data(2000)
        print(f"Generated {len(real_plans)} training plans with {real_plans.shape[1]} features")
        
        # Generate user profiles (10 dimensions)
        user_profiles = np.random.uniform(0, 1, (2000, self.user_profile_dim))
        
        # Generate training class one-hot encoding (4 dimensions)
        training_classes = np.random.randint(0, 4, 2000)
        training_classes_onehot = keras.utils.to_categorical(training_classes, self.training_class_dim)
        
        valid = np.ones((batch_size, 1))
        fake = np.zeros((batch_size, 1))
        
        history = {
            'd_loss': [],
            'g_loss': [],
            'd_acc': []
        }
        
        print("\n[2/4] Training GAN...")
        for epoch in range(epochs):
            # Select random batch
            idx = np.random.randint(0, real_plans.shape[0], batch_size)
            real_plan_batch = real_plans[idx]
            profile_batch = user_profiles[idx]
            class_batch = training_classes_onehot[idx]
            
            # Sample noise
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            
            # Train discriminator
            generated_plans = self.generator.predict(
                [noise, profile_batch, class_batch], 
                verbose=0
            )
            
            d_loss_real = self.discriminator.train_on_batch(real_plan_batch, valid)
            d_loss_fake = self.discriminator.train_on_batch(generated_plans, fake)
            d_loss = 0.5 * np.add(d_loss_real[0], d_loss_fake[0])
            d_acc = 0.5 * np.add(d_loss_real[1], d_loss_fake[1])
            
            # Train generator
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            g_loss = self.gan.train_on_batch([noise, profile_batch, class_batch], valid)[0]
            
            history['d_loss'].append(float(d_loss))
            history['g_loss'].append(float(g_loss))
            history['d_acc'].append(float(d_acc))
            
            if epoch % save_interval == 0:
                print(f"Epoch {epoch}: D loss: {d_loss:.4f}, G loss: {g_loss:.4f}, D acc: {d_acc:.4f}")
        
        print("\n[3/4] Saving models...")
        os.makedirs('../../models', exist_ok=True)
        self.generator.save('../../models/generator.h5')
        print("Generator saved to ../../models/generator.h5")
        
        # Save history
        history['timestamp'] = datetime.now().isoformat()
        history['plan_dim'] = self.plan_dim
        history['user_profile_dim'] = self.user_profile_dim
        history['training_class_dim'] = self.training_class_dim
        
        with open('../../models/gan_training_history.json', 'w', encoding='utf-8') as f:
            json.dump(history, f, indent=2, ensure_ascii=False)
        print("Training history saved")
        
        print("\n[4/4] Training Complete!")
        print("=" * 60)
        
        return self.generator, history


def decode_plan(plan_vector, training_class, user_profile):
    """Decode GAN output to human-readable training plan"""
    class_names = list(TRAINING_TEMPLATES.keys())
    class_name = class_names[training_class]
    template = TRAINING_TEMPLATES[class_name]
    
    duration = int(plan_vector[0] * 100)
    intensity = plan_vector[1]
    rest_ratio = plan_vector[2]
    weekly_freq = int(plan_vector[3] * 7)
    
    exercise_probs = plan_vector[4:9]
    primary_exercise_idx = np.argmax(exercise_probs)
    primary_exercise = template['exercises'][primary_exercise_idx % len(template['exercises'])]
    
    warmup = int(plan_vector[9] * 100)
    cooldown = int(plan_vector[10] * 100)
    
    # Build plan
    plan = {
        'training_type': class_name,
        'training_type_ru': template['name_ru'],
        'duration_minutes': max(20, min(120, duration)),
        'intensity': round(intensity, 2),
        'weekly_frequency': max(1, min(7, weekly_freq)),
        'primary_exercise': primary_exercise,
        'warmup_minutes': max(5, min(20, warmup)),
        'cooldown_minutes': max(5, min(20, cooldown)),
        'rest_ratio': round(rest_ratio, 2),
        'exercises': template['exercises'],
        'notes': []
    }
    
    # Add personalized notes based on user profile
    if user_profile.get('fitness_level') == 'beginner':
        plan['notes'].append("Начните с 50% от рекомендованной интенсивности")
        plan['duration_minutes'] = int(plan['duration_minutes'] * 0.7)
    
    if user_profile.get('age', 0) > 50:
        plan['notes'].append("Увеличьте время разминки и заминки")
        plan['warmup_minutes'] += 5
        plan['cooldown_minutes'] += 5
    
    if user_profile.get('health_conditions'):
        plan['notes'].append("Проконсультируйтесь с врачом перед началом")
    
    if user_profile.get('goals'):
        goals_lower = [g.lower() for g in user_profile['goals']]
        if 'похудение' in goals_lower or 'weight_loss' in goals_lower:
            plan['notes'].append("Добавьте 10-15 минут кардио после основной тренировки")
        if 'набор массы' in goals_lower or 'muscle_gain' in goals_lower:
            plan['notes'].append("Сфокусируйтесь на силовых упражнениях")
        if 'реабилитация' in goals_lower or 'rehabilitation' in goals_lower:
            plan['notes'].append("Следите за техникой выполнения упражнений")
    
    return plan


def train_and_save():
    """Main training function"""
    # plan_dim=16 must match discriminator input
    gan = TrainingPlanGAN(
        latent_dim=32, 
        user_profile_dim=10, 
        training_class_dim=4, 
        plan_dim=16
    )
    generator, history = gan.train(epochs=500, batch_size=32, save_interval=50)
    
    # Test generation
    print("\n" + "=" * 60)
    print("Testing Plan Generation")
    print("=" * 60)
    
    test_noise = np.random.normal(0, 1, (1, 32))
    test_profile = np.random.uniform(0, 1, (1, 10))
    test_class = keras.utils.to_categorical([1], 4)  # Endurance
    
    generated_plan = generator.predict([test_noise, test_profile, test_class], verbose=0)
    decoded = decode_plan(
        generated_plan[0], 
        1, 
        {'fitness_level': 'intermediate', 'age': 30, 'goals': ['похудение']}
    )
    
    print("\nSample Generated Plan:")
    print(json.dumps(decoded, indent=2, ensure_ascii=False))
    
    return generator


if __name__ == '__main__':
    train_and_save()