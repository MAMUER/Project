import numpy as np
import tensorflow as tf
from tensorflow.keras import layers, models
import joblib

class TrainingClassifier:
    def __init__(self):
        self.model = None
        self.classes = ['cardio', 'strength', 'flexibility', 'recovery', 'hiit', 'endurance']
        self._build_model()
    
    def _build_model(self):
        # Нейросеть: 6 входных параметров -> 6 классов
        self.model = models.Sequential([
            layers.Dense(32, activation='relu', input_shape=(6,)),
            layers.Dropout(0.2),
            layers.Dense(16, activation='relu'),
            layers.Dropout(0.2),
            layers.Dense(8, activation='relu'),
            layers.Dense(len(self.classes), activation='softmax')
        ])
        
        self.model.compile(
            optimizer='adam',
            loss='categorical_crossentropy',
            metrics=['accuracy']
        )
    
    def train(self, X_train, y_train, epochs=50, batch_size=32, validation_split=0.2):
        """Обучение модели"""
        history = self.model.fit(
            X_train, y_train,
            epochs=epochs,
            batch_size=batch_size,
            validation_split=validation_split,
            verbose=1
        )
        return history
    
    def predict(self, features):
        """Предсказание класса"""
        input_data = np.array([[
            features['heart_rate'],
            features['ecg_mean'],
            features['systolic'],
            features['diastolic'],
            features['spo2'],
            features['temperature']
        ]])
        
        pred = self.model.predict(input_data, verbose=0)
        class_idx = np.argmax(pred[0])
        confidence = float(pred[0][class_idx])
        
        return self.classes[class_idx], confidence
    
    def save(self, path='models/classifier_model.h5'):
        """Сохранение модели"""
        self.model.save(path)
    
    def load(self, path='models/classifier_model.h5'):
        """Загрузка модели"""
        self.model = tf.keras.models.load_model(path)


# Функция для предобработки данных
def preprocess_features(heart_rate, ecg_str, systolic, diastolic, spo2, temperature):
    ecg_values = [int(x) for x in ecg_str.split(',')] if ecg_str else [0]
    ecg_mean = sum(ecg_values) / len(ecg_values) if ecg_values else 0
    
    return {
        'heart_rate': heart_rate / 200.0,  # нормализация
        'ecg_mean': ecg_mean / 100.0,
        'systolic': systolic / 200.0,
        'diastolic': diastolic / 120.0,
        'spo2': spo2 / 100.0,
        'temperature': (temperature - 35) / 5  # нормализация
    }