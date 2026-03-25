import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
import tensorflow as tf
from tensorflow import keras # type: ignore
import os

def load_and_prepare_data(data_path):
    """Загрузка и подготовка данных из датасета"""
    # Заглушка - в реальности загружаем из файлов
    # Для примера создаем синтетические данные
    np.random.seed(42)
    n_samples = 10000
    
    # 6 входных признаков
    X = np.random.randn(n_samples, 6)
    
    # 4 класса: endurance, strength, recovery, interval
    y = np.random.randint(0, 4, n_samples)
    
    return X, y

def build_model():
    """Построение нейросети"""
    model = keras.Sequential([
        keras.layers.Input(shape=(6,)),
        keras.layers.Dense(32, activation='relu'),
        keras.layers.Dropout(0.2),
        keras.layers.Dense(16, activation='relu'),
        keras.layers.Dense(4, activation='softmax')
    ])
    
    model.compile(
        optimizer='adam',
        loss='sparse_categorical_crossentropy',
        metrics=['accuracy']
    )
    return model

def train():
    print("Loading data...")
    X, y = load_and_prepare_data('datasets/processed/')
    
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
    
    scaler = StandardScaler()
    X_train = scaler.fit_transform(X_train)
    X_test = scaler.transform(X_test)
    
    model = build_model()
    
    print("Training model...")
    history = model.fit(
        X_train, y_train,
        epochs=50,
        batch_size=32,
        validation_data=(X_test, y_test),
        verbose=1
    )
    
    # Сохранение модели
    model.save('models/classifier.h5')
    print("Model saved to models/classifier.h5")
    
    # Сохранение scaler
    import joblib
    joblib.dump(scaler, 'models/scaler.pkl')
    
    # Оценка
    loss, acc = model.evaluate(X_test, y_test)
    print(f"Test accuracy: {acc:.4f}")

if __name__ == "__main__":
    train()