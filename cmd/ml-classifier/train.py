"""
Training script for ML Classifier
Classifies training types based on physiological parameters
Based on: Recovery, Endurance E1-E2, Threshold E3, Strength/HIIT
"""
import os
import json
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import classification_report, confusion_matrix
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import layers, models, callbacks
import joblib
from datetime import datetime

# Suppress TF warnings
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'
tf.get_logger().setLevel('ERROR')

# Training type classes (from document)
TRAINING_CLASSES = {
    0: {
        'name': 'recovery',
        'name_ru': 'Восстановление',
        'hr_range': '50-65% HRmax',
        'hrv': 'Высокий',
        'spo2': '96-99%',
        'marker': 'Низкая нагрузка + высокий HRV + хорошее восстановление'
    },
    1: {
        'name': 'endurance_e1e2',
        'name_ru': 'Базовая выносливость (E1-E2)',
        'hr_range': '65-80% HRmax',
        'hrv': 'Умеренный',
        'spo2': '95-98%',
        'marker': 'Работа ниже лактатного порога'
    },
    2: {
        'name': 'threshold_e3',
        'name_ru': 'Пороговая выносливость (E3)',
        'hr_range': '80-90% HRmax',
        'hrv': 'Сниженный',
        'spo2': '93-96%',
        'marker': 'Нагрузка вблизи анаэробного порога'
    },
    3: {
        'name': 'strength_hiit',
        'name_ru': 'Силовая/HIIT',
        'hr_range': '90-100% HRmax',
        'hrv': 'Резкое падение',
        'spo2': '90-94%',
        'marker': 'Высокая вариабельность + постнагрузочная гипертензия'
    }
}


def load_real_data():
    """
    Загрузка данных из предобработанного CSV
    Пробует несколько путей к файлу
    """
    possible_paths = [
        '../../datasets/processed/training_data_real.csv',
        '../../datasets/processed/training_data.csv',
        'datasets/processed/training_data_real.csv',
        'datasets/processed/training_data.csv'
    ]
    
    data_path = None
    for path in possible_paths:
        if os.path.exists(path):
            data_path = path
            break
    
    if data_path is None:
        print("⚠️  Real data not found. Falling back to synthetic data...")
        return generate_synthetic_data(5000)
    
    print(f"✅ Loading real data from: {data_path}")
    df = pd.read_csv(data_path)
    
    # Проверка колонок
    required_cols = ['hr', 'hrv', 'spo2', 'temp', 'bp_s', 'bp_d', 'sleep', 'label']
    missing_cols = [col for col in required_cols if col not in df.columns]
    if missing_cols:
        raise ValueError(f"Missing columns: {missing_cols}")
    
    # Формирование матриц (порядок как в main.py)
    X = df[['hr', 'hrv', 'spo2', 'temp', 'bp_s', 'bp_d', 'sleep']].values
    y = df['label'].values
    
    print(f"✅ Loaded {len(df)} samples")
    print(f"   Class distribution: {np.bincount(y.astype(int))}")
    
    return X, y


def generate_synthetic_data(n_samples=5000):
    """
    Generate synthetic physiological data for training
    Based on the training type parameters from the document
    """
    np.random.seed(42)
    features = []
    labels = []
    
    samples_per_class = n_samples // 4
    
    for class_id in range(4):
        for _ in range(samples_per_class):
            if class_id == 0:  # Recovery
                hr = np.random.uniform(50, 65)
                hrv = np.random.uniform(60, 100)
                spo2 = np.random.uniform(96, 99)
                temp = np.random.uniform(36.5, 37.0)
                bp_systolic = np.random.uniform(110, 130)
                bp_diastolic = np.random.uniform(70, 85)
                sleep_hours = np.random.uniform(7, 9)
                
            elif class_id == 1:  # Endurance E1-E2
                hr = np.random.uniform(65, 80)
                hrv = np.random.uniform(40, 70)
                spo2 = np.random.uniform(95, 98)
                temp = np.random.uniform(37.0, 37.8)
                bp_systolic = np.random.uniform(130, 150)
                bp_diastolic = np.random.uniform(80, 90)
                sleep_hours = np.random.uniform(6, 8)
                
            elif class_id == 2:  # Threshold E3
                hr = np.random.uniform(80, 90)
                hrv = np.random.uniform(20, 50)
                spo2 = np.random.uniform(93, 96)
                temp = np.random.uniform(37.5, 38.2)
                bp_systolic = np.random.uniform(150, 170)
                bp_diastolic = np.random.uniform(85, 100)
                sleep_hours = np.random.uniform(5, 7)
                
            else:  # Strength/HIIT
                hr = np.random.uniform(90, 100)
                hrv = np.random.uniform(10, 30)
                spo2 = np.random.uniform(90, 94)
                temp = np.random.uniform(38.0, 39.0)
                bp_systolic = np.random.uniform(170, 200)
                bp_diastolic = np.random.uniform(95, 110)
                sleep_hours = np.random.uniform(4, 6)
            
            features.append([hr, hrv, spo2, temp, bp_systolic, bp_diastolic, sleep_hours])
            labels.append(class_id)
    
    indices = np.random.permutation(len(features))
    features = np.array(features)[indices]
    labels = np.array(labels)[indices]
    
    return features, labels


def create_classifier_model(input_shape=7, num_classes=4):
    """
    Create the classifier neural network model
    """
    model = models.Sequential([
        layers.Input(shape=(input_shape,)),
        
        layers.Dense(64, activation='relu'),
        layers.BatchNormalization(),
        layers.Dropout(0.3),
        
        layers.Dense(32, activation='relu'),
        layers.BatchNormalization(),
        layers.Dropout(0.3),
        
        layers.Dense(16, activation='relu'),
        layers.BatchNormalization(),
        layers.Dropout(0.2),
        
        layers.Dense(num_classes, activation='softmax')
    ])
    
    model.compile(
        optimizer=keras.optimizers.Adam(learning_rate=0.001),
        loss='sparse_categorical_crossentropy',
        metrics=['accuracy']
    )
    
    return model


def train_model():
    """
    Main training function
    """
    print("=" * 60)
    print("Starting ML Classifier Training")
    print("=" * 60)
    
    # Load/generate data
    print("\n[1/5] Loading physiological data...")
    X, y = load_real_data()
    print(f"Total samples: {len(X)} with {len(np.unique(y))} classes")
    print(f"Feature shape: {X.shape}")
    
    # Split data
    print("\n[2/5] Splitting data...")
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )
    print(f"Train: {len(X_train)}, Test: {len(X_test)}")
    
    # Scale features
    print("\n[3/5] Scaling features...")
    scaler = StandardScaler()
    X_train_scaled = scaler.fit_transform(X_train)
    X_test_scaled = scaler.transform(X_test)
    
    # Save scaler
    os.makedirs('../../models', exist_ok=True)
    joblib.dump(scaler, '../../models/scaler.pkl')
    print("Scaler saved to ../../models/scaler.pkl")
    
    # Create model
    print("\n[4/5] Creating model...")
    model = create_classifier_model(input_shape=X_train_scaled.shape[1])
    model.summary()
    
    # Callbacks
    early_stop = callbacks.EarlyStopping(
        monitor='val_loss',
        patience=10,
        restore_best_weights=True,
        verbose=1
    )
    
    reduce_lr = callbacks.ReduceLROnPlateau(
        monitor='val_loss',
        factor=0.5,
        patience=5,
        min_lr=1e-6,
        verbose=1
    )
    
    checkpoint = callbacks.ModelCheckpoint(
        '../../models/classifier.keras',
        monitor='val_accuracy',
        save_best_only=True,
        verbose=1
    )
    
    # Train
    print("\n[5/5] Training model...")
    history = model.fit(
        X_train_scaled, y_train,
        validation_data=(X_test_scaled, y_test),
        epochs=50,
        batch_size=32,
        callbacks=[early_stop, reduce_lr, checkpoint],
        verbose=1
    )
    
    # Evaluate
    print("\n" + "=" * 60)
    print("Evaluation Results")
    print("=" * 60)
    
    y_pred = np.argmax(model.predict(X_test_scaled, verbose=0), axis=1)
    print("\nClassification Report:")
    print(classification_report(y_test, y_pred, 
                                target_names=[TRAINING_CLASSES[i]['name_ru'] for i in range(4)]))
    
    print("\nConfusion Matrix:")
    print(confusion_matrix(y_test, y_pred))
    
    # Save model
    model.save('../../models/classifier.keras')
    print("\nModel saved to ../../models/classifier.keras")
    
    # Save training history
    training_history = {
        'accuracy': [float(a) for a in history.history['accuracy']],
        'val_accuracy': [float(a) for a in history.history['val_accuracy']],
        'loss': [float(l) for l in history.history['loss']],
        'val_loss': [float(l) for l in history.history['val_loss']],
        'timestamp': datetime.now().isoformat(),
        'classes': TRAINING_CLASSES
    }
    
    with open('../../models/training_history.json', 'w', encoding='utf-8') as f:
        json.dump(training_history, f, indent=2, ensure_ascii=False)
    print("Training history saved to ../../models/training_history.json")
    
    print("\n" + "=" * 60)
    print("Training Complete!")
    print("=" * 60)
    
    return model, scaler


if __name__ == '__main__':
    train_model()