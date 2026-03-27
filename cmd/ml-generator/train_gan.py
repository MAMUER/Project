import numpy as np
import tensorflow as tf
from tensorflow import keras # type: ignore
import os

def build_generator(latent_dim=100, output_dim=100):
    """Генератор: латентный вектор -> параметры программы тренировки"""
    model = keras.Sequential([
        keras.layers.Dense(128, activation='relu', input_shape=(latent_dim,)),
        keras.layers.BatchNormalization(),
        keras.layers.Dense(256, activation='relu'),
        keras.layers.BatchNormalization(),
        keras.layers.Dense(512, activation='relu'),
        keras.layers.BatchNormalization(),
        keras.layers.Dense(output_dim, activation='tanh')
    ])
    return model

def build_discriminator(input_dim=100):
    """Дискриминатор: параметры программы -> реальная/фейковая"""
    model = keras.Sequential([
        keras.layers.Dense(512, activation='relu', input_shape=(input_dim,)),
        keras.layers.Dropout(0.3),
        keras.layers.Dense(256, activation='relu'),
        keras.layers.Dropout(0.3),
        keras.layers.Dense(128, activation='relu'),
        keras.layers.Dense(1, activation='sigmoid')
    ])
    return model

def build_gan(generator, discriminator):
    """GAN: генератор + замороженный дискриминатор"""
    discriminator.trainable = False
    gan = keras.Sequential([generator, discriminator])
    gan.compile(optimizer=keras.optimizers.Adam(0.0002, 0.5), loss='binary_crossentropy')
    return gan

def generate_synthetic_training_plans(n_samples=5000, output_dim=100):
    """Генерация синтетических программ тренировок для обучения GAN"""
    # Создаём шаблонные программы с разной структурой
    plans = []
    for _ in range(n_samples):
        # Тип программы (0-3)
        prog_type = np.random.randint(0, 4)
        
        # Генерация параметров программы
        plan = np.zeros(output_dim)
        
        # Базовые параметры
        plan[0] = prog_type / 3.0  # тип
        plan[1] = np.random.rand()  # интенсивность
        plan[2] = np.random.randint(4, 12) / 12.0  # длительность (недели)
        plan[3] = np.random.randint(3, 7) / 7.0  # тренировок в неделю
        
        # Упражнения (просто случайные значения)
        for i in range(4, min(output_dim, 20)):
            plan[i] = np.random.rand()
        
        plans.append(plan)
    
    return np.array(plans)

def train_gan():
    print("=" * 50)
    print("TRAINING GAN FOR TRAINING PROGRAM GENERATION")
    print("=" * 50)
    
    latent_dim = 100
    output_dim = 100
    epochs = 5000
    batch_size = 32
    
    # Создание моделей
    generator = build_generator(latent_dim, output_dim)
    discriminator = build_discriminator(output_dim)
    discriminator.compile(optimizer=keras.optimizers.Adam(0.0002, 0.5), 
                          loss='binary_crossentropy', metrics=['accuracy'])
    
    gan = build_gan(generator, discriminator)
    
    # Синтетические реальные данные
    real_data = generate_synthetic_training_plans(10000, output_dim)
    
    print(f"Training data shape: {real_data.shape}")
    
    for epoch in range(epochs):
        # Обучение дискриминатора
        noise = np.random.randn(batch_size, latent_dim)
        generated = generator.predict(noise, verbose=0)
        
        idx = np.random.randint(0, real_data.shape[0], batch_size)
        real_batch = real_data[idx]
        
        X = np.concatenate([real_batch, generated])
        y = np.concatenate([np.ones(batch_size), np.zeros(batch_size)])
        
        d_loss = discriminator.train_on_batch(X, y)
        
        # Обучение генератора
        noise = np.random.randn(batch_size, latent_dim)
        y_gan = np.ones(batch_size)
        g_loss = gan.train_on_batch(noise, y_gan)
        
        if epoch % 500 == 0:
            print(f"Epoch {epoch}: D loss: {d_loss[0]:.4f}, D acc: {d_loss[1]:.4f}, G loss: {g_loss:.4f}")
    
    # Сохранение модели
    os.makedirs('models', exist_ok=True)
    generator.save('models/generator.h5')
    print("\n✅ GAN model saved to models/generator.h5")

if __name__ == "__main__":
    train_gan()