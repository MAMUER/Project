import numpy as np
import tensorflow as tf
from tensorflow import keras

def build_generator(latent_dim=100):
    """Построение генератора GAN"""
    model = keras.Sequential([
        keras.layers.Dense(128, activation='relu', input_shape=(latent_dim,)),
        keras.layers.Dense(256, activation='relu'),
        keras.layers.Dense(512, activation='relu'),
        keras.layers.Dense(100, activation='tanh')  # 100 параметров плана
    ])
    return model

def build_discriminator():
    """Построение дискриминатора GAN"""
    model = keras.Sequential([
        keras.layers.Dense(512, activation='relu', input_shape=(100,)),
        keras.layers.Dropout(0.3),
        keras.layers.Dense(256, activation='relu'),
        keras.layers.Dropout(0.3),
        keras.layers.Dense(1, activation='sigmoid')
    ])
    return model

def build_gan(generator, discriminator):
    """Построение GAN"""
    discriminator.trainable = False
    gan = keras.Sequential([generator, discriminator])
    gan.compile(optimizer='adam', loss='binary_crossentropy')
    return gan

def train_gan():
    latent_dim = 100
    epochs = 10000
    batch_size = 32
    
    generator = build_generator(latent_dim)
    discriminator = build_discriminator()
    discriminator.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
    
    gan = build_gan(generator, discriminator)
    
    # Синтетические данные для обучения
    real_data = np.random.randn(1000, 100)
    
    for epoch in range(epochs):
        # Обучение дискриминатора
        noise = np.random.randn(batch_size, latent_dim)
        generated = generator.predict(noise, verbose=0)
        
        real_batch = real_data[np.random.randint(0, real_data.shape[0], batch_size)]
        X = np.concatenate([real_batch, generated])
        y = np.concatenate([np.ones(batch_size), np.zeros(batch_size)])
        
        d_loss = discriminator.train_on_batch(X, y)
        
        # Обучение генератора
        noise = np.random.randn(batch_size, latent_dim)
        y_gan = np.ones(batch_size)
        g_loss = gan.train_on_batch(noise, y_gan)
        
        if epoch % 1000 == 0:
            print(f"Epoch {epoch}: D loss: {d_loss[0]}, G loss: {g_loss}")
    
    generator.save('models/generator.h5')
    print("GAN model saved to models/generator.h5")

if __name__ == "__main__":
    train_gan()