import numpy as np
import tensorflow as tf
from tensorflow.keras import layers, models

class TrainingProgramGAN:
    def __init__(self, latent_dim=100, condition_dim=20, output_dim=210):
        self.latent_dim = latent_dim
        self.condition_dim = condition_dim
        self.output_dim = output_dim
        
        self.generator = self._build_generator()
        self.discriminator = self._build_discriminator()
        self.gan = self._build_gan()
    
    def _build_generator(self):
        """Генератор: шум + условия -> программа тренировок"""
        noise_input = layers.Input(shape=(self.latent_dim,))
        condition_input = layers.Input(shape=(self.condition_dim,))
        
        combined = layers.Concatenate()([noise_input, condition_input])
        
        x = layers.Dense(256, activation='relu')(combined)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(512, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        x = layers.Dropout(0.2)(x)
        
        x = layers.Dense(1024, activation='relu')(x)
        x = layers.BatchNormalization()(x)
        
        output = layers.Dense(self.output_dim, activation='tanh')(x)
        
        model = models.Model([noise_input, condition_input], output)
        return model
    
    def _build_discriminator(self):
        """Дискриминатор: программа -> real/fake"""
        program_input = layers.Input(shape=(self.output_dim,))
        condition_input = layers.Input(shape=(self.condition_dim,))
        
        combined = layers.Concatenate()([program_input, condition_input])
        
        x = layers.Dense(512, activation='relu')(combined)
        x = layers.Dropout(0.3)(x)
        x = layers.Dense(256, activation='relu')(x)
        x = layers.Dropout(0.3)(x)
        x = layers.Dense(128, activation='relu')(x)
        x = layers.Dropout(0.3)(x)
        
        output = layers.Dense(1, activation='sigmoid')(x)
        
        model = models.Model([program_input, condition_input], output)
        model.compile(optimizer='adam', loss='binary_crossentropy')
        return model
    
    def _build_gan(self):
        """GAN: генератор + дискриминатор (дискриминатор заморожен)"""
        self.discriminator.trainable = False
        
        noise_input = layers.Input(shape=(self.latent_dim,))
        condition_input = layers.Input(shape=(self.condition_dim,))
        
        generated_program = self.generator([noise_input, condition_input])
        validity = self.discriminator([generated_program, condition_input])
        
        model = models.Model([noise_input, condition_input], validity)
        model.compile(optimizer='adam', loss='binary_crossentropy')
        return model
    
    def train(self, real_programs, conditions, epochs=10000, batch_size=32):
        """Обучение GAN"""
        for epoch in range(epochs):
            # Обучение дискриминатора
            idx = np.random.randint(0, real_programs.shape[0], batch_size)
            real_batch = real_programs[idx]
            condition_batch = conditions[idx]
            
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            fake_batch = self.generator.predict([noise, condition_batch], verbose=0)
            
            real_labels = np.ones((batch_size, 1))
            fake_labels = np.zeros((batch_size, 1))
            
            d_loss_real = self.discriminator.train_on_batch(
                [real_batch, condition_batch], real_labels)
            d_loss_fake = self.discriminator.train_on_batch(
                [fake_batch, condition_batch], fake_labels)
            d_loss = 0.5 * np.add(d_loss_real, d_loss_fake)
            
            # Обучение генератора
            noise = np.random.normal(0, 1, (batch_size, self.latent_dim))
            condition_batch = conditions[np.random.randint(0, conditions.shape[0], batch_size)]
            
            g_loss = self.gan.train_on_batch([noise, condition_batch], np.ones((batch_size, 1)))
            
            if epoch % 1000 == 0:
                print(f"Epoch {epoch}: D loss: {d_loss:.4f}, G loss: {g_loss:.4f}")
    
    def generate(self, conditions, num_samples=1):
        """Генерация программ тренировок"""
        noise = np.random.normal(0, 1, (num_samples, self.latent_dim))
        generated = self.generator.predict([noise, conditions], verbose=0)
        return generated
    
    def save(self, path_prefix='models/gan'):
        """Сохранение модели"""
        self.generator.save(f'{path_prefix}_generator.h5')
        self.discriminator.save(f'{path_prefix}_discriminator.h5')
    
    def load(self, path_prefix='models/gan'):
        """Загрузка модели"""
        self.generator = tf.keras.models.load_model(f'{path_prefix}_generator.h5')
        self.discriminator = tf.keras.models.load_model(f'{path_prefix}_discriminator.h5')
        self.gan = self._build_gan()