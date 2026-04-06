# cmd/ml-generator/train_gan.py
"""
Training script for GAN-based Training Plan Generator - Keras 3 compatible
"""
import os
import sys
import json
import numpy as np
from datetime import datetime
import keras
from keras import layers, models
import matplotlib.pyplot as plt

os.environ['KERAS_BACKEND'] = 'tensorflow'
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'

import tensorflow as tf
tf.get_logger().setLevel('ERROR')

# Constants
TRAINING_TEMPLATES = {
    'recovery': {'duration_range': (20, 45), 'intensity_range': (0.3, 0.5), 'exercises': ['walking', 'yoga', 'stretching']},
    'endurance_e1e2': {'duration_range': (45, 90), 'intensity_range': (0.5, 0.7), 'exercises': ['running', 'cycling', 'swimming']},
    'threshold_e3': {'duration_range': (30, 60), 'intensity_range': (0.7, 0.85), 'exercises': ['tempo_run', 'intervals']},
    'strength_hiit': {'duration_range': (20, 45), 'intensity_range': (0.85, 1.0), 'exercises': ['hiit', 'strength', 'sprints']}
}

class TrainingPlanGAN:
    """GAN for generating training plans"""
    
    def __init__(self, latent_dim=32, user_profile_dim=10, training_class_dim=4, plan_dim=16):
        self.latent_dim = latent_dim
        self.user_profile_dim = user_profile_dim
        self.training_class_dim = training_class_dim
        self.plan_dim = plan_dim
        
        self.generator = self.build_generator()
        self.discriminator = self.build_discriminator()
    
    def build_generator(self):
        """Build the generator network"""
        input_latent = layers.Input(shape=(self.latent_dim,), name='latent_input')
        input_profile = layers.Input(shape=(self.user_profile_dim,), name='profile_input')
        input_class = layers.Input(shape=(self.training_class_dim,), name='class_input')
        
        combined = layers.concatenate([input_latent, input_profile, input_class])
        
        x = layers.Dense(512, activation='relu')(combined)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(256, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(128, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        
        output = layers.Dense(self.plan_dim, activation='sigmoid', name='plan_output')(x)
        
        model = models.Model(
            inputs=[input_latent, input_profile, input_class],
            outputs=output,
            name='generator'
        )
        
        model.compile(
            optimizer=keras.optimizers.Adam(learning_rate=0.0002, beta_1=0.5),
            loss='binary_crossentropy'
        )
        
        return model
    
    def build_discriminator(self):
        """Build the discriminator network"""
        input_plan = layers.Input(shape=(self.plan_dim,), name='plan_input')
        
        x = layers.Dense(256, activation='relu')(input_plan)
        x = layers.Dropout(0.3)(x)
        
        x = layers.Dense(128, activation='relu')(x)
        x = layers.Dropout(0.3)(x)
        
        x = layers.Dense(64, activation='relu')(x)
        
        output = layers.Dense(1, activation='sigmoid', name='discriminator_output')(x)
        
        model = models.Model(inputs=input_plan, outputs=output, name='discriminator')
        
        model.compile(
            optimizer=keras.optimizers.Adam(learning_rate=0.0002, beta_1=0.5),
            loss='binary_crossentropy',
            metrics=['accuracy']
        )
        
        return model
    
    def generate_training_data(self, n_samples=5000):
        """Generate synthetic training plan data"""
        np.random.seed(42)
        plans = []
        
        for _ in range(n_samples):
            training_class = np.random.randint(0, 4)
            class_name = list(TRAINING_TEMPLATES.keys())[training_class]
            template = TRAINING_TEMPLATES[class_name]
            
            duration = np.random.uniform(*template['duration_range']) / 100
            intensity = np.random.uniform(*template['intensity_range'])
            rest_ratio = np.random.uniform(max(0, template.get('rest_ratio', 0.5) - 0.1),
                                          min(1, template.get('rest_ratio', 0.5) + 0.1))
            weekly_freq = np.random.uniform(2, 6) / 7
            
            exercise_probs = np.random.dirichlet(np.ones(5)) * 0.5
            
            warmup = np.random.uniform(5, 15) / 100
            cooldown = np.random.uniform(5, 15) / 100
            
            plan = np.array([
                duration, intensity, rest_ratio, weekly_freq,
                exercise_probs[0], exercise_probs[1], exercise_probs[2],
                exercise_probs[3], exercise_probs[4],
                warmup, cooldown,
                np.random.uniform(0.1, 0.3),  # progression
                np.random.uniform(0.5, 1.0),  # age_factor
                np.random.uniform(0.5, 1.0),  # fitness_factor
                np.random.uniform(0.5, 1.0),  # health_factor
                np.random.uniform(0.5, 1.0)   # goal_factor
            ])
            
            plans.append(plan)
        
        return np.array(plans)
    
    def train(self, epochs=500, batch_size=32, save_interval=50):
        """Train the GAN"""
        print("=" * 60)
        print("STARTING GAN TRAINING")
        print("=" * 60)
        
        print("\n[1/4] Generating training data...")
        real_plans = self.generate_training_data(5000)
        print(f"Generated {len(real_plans)} training plans with {real_plans.shape[1]} features")
        
        user_profiles = np.random.uniform(0, 1, (5000, self.user_profile_dim))
        training_classes = np.random.randint(0, 4, 5000)
        training_classes_onehot = keras.utils.to_categorical(training_classes, self.training_class_dim)
        
        valid = np.ones((batch_size, 1))
        fake = np.zeros((batch_size, 1))
        
        history = {'d_loss': [], 'g_loss': [], 'd_acc': []}
        
        print("\n[2/4] Training GAN...")
        for epoch in range(epochs):
            # Select random batch
            idx = np.random.randint(0, real_plans.shape[0], batch_size)
            real_plan_batch = real_plans[idx]
            profile_batch = user_profiles[idx]
            class_batch = training_classes_onehot[idx]
            
            # Sample noise
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            
            # TRAIN DISCRIMINATOR
            self.discriminator.trainable = True
            
            generated_plans = self.generator.predict([noise, profile_batch, class_batch], verbose=0)
            
            d_loss_real = self.discriminator.train_on_batch(real_plan_batch, valid)
            d_loss_fake = self.discriminator.train_on_batch(generated_plans, fake)
            
            d_loss = 0.5 * (d_loss_real[0] + d_loss_fake[0])
            d_acc = 0.5 * (d_loss_real[1] + d_loss_fake[1])
            
            # TRAIN GENERATOR
            self.discriminator.trainable = False
            
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            
            generated_plans = self.generator.predict([noise, profile_batch, class_batch], verbose=0)
            g_loss = self.discriminator.train_on_batch(generated_plans, valid)
            
            # Restore discriminator trainable
            self.discriminator.trainable = True
            
            history['d_loss'].append(float(d_loss))
            history['g_loss'].append(float(g_loss[0]) if isinstance(g_loss, list) else float(g_loss))
            history['d_acc'].append(float(d_acc))
            
            if epoch % save_interval == 0:
                print(f"Epoch {epoch}: D loss: {history['d_loss'][-1]:.4f}, "
                      f"G loss: {history['g_loss'][-1]:.4f}, D acc: {history['d_acc'][-1]:.4f}")
        
        print("\n[3/4] Saving models...")
        os.makedirs('../../models', exist_ok=True)
        self.generator.save('../../models/generator.keras')
        print("Generator saved to ../../models/generator.keras")
        
        history['timestamp'] = datetime.now().isoformat()
        history['plan_dim'] = self.plan_dim
        history['user_profile_dim'] = self.user_profile_dim
        history['training_class_dim'] = self.training_class_dim
        
        with open('../../models/gan_training_history.json', 'w', encoding='utf-8') as f:
            json.dump(history, f, indent=2, ensure_ascii=False)
        print("Training history saved")
        
        # Визуализация
        plt.figure(figsize=(12, 5))
        
        plt.subplot(1, 2, 1)
        plt.plot(history['d_loss'], label='D Loss')
        plt.plot(history['g_loss'], label='G Loss')
        plt.xlabel('Epoch')
        plt.ylabel('Loss')
        plt.legend()
        plt.grid(True, alpha=0.3)
        
        plt.subplot(1, 2, 2)
        plt.plot(history['d_acc'], label='D Accuracy')
        plt.xlabel('Epoch')
        plt.ylabel('Accuracy')
        plt.legend()
        plt.grid(True, alpha=0.3)
        
        plt.tight_layout()
        plt.savefig('../../models/gan_training_history.png', dpi=150)
        print("GAN training plot saved")
        
        print("\n[4/4] Training Complete!")
        print("=" * 60)
        
        return self.generator, history


def train_and_save():
    """Main training function"""
    gan = TrainingPlanGAN(
        latent_dim=32,
        user_profile_dim=10,
        training_class_dim=4,
        plan_dim=16
    )
    generator, history = gan.train(epochs=500, batch_size=32, save_interval=50)
    
    print("\n" + "=" * 60)
    print("Testing Plan Generation")
    print("=" * 60)
    
    test_noise = np.random.normal(0, 1, (1, 32))
    test_profile = np.random.uniform(0, 1, (1, 10))
    test_class = keras.utils.to_categorical([1], 4)
    
    generated_plan = generator.predict([test_noise, test_profile, test_class], verbose=0)
    print(f"\nSample generated plan vector: {generated_plan[0][:5]}...")
    
    return generator


if __name__ == '__main__':
    train_and_save()