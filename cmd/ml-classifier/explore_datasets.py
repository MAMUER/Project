# cmd/ml-classifier/explore_datasets.py
"""
Dataset Exploration Script - FIXED for ACTUAL folder names
"""
import os
import pandas as pd
import numpy as np
import pickle
import json
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any

# Configuration - ACTUAL FOLDER NAMES
RAW_DATA_DIR = '../../datasets/raw'
OUTPUT_DIR = '../../datasets/processed'
REPORT_FILE = f'{OUTPUT_DIR}/dataset_exploration_report.json'

# Mapping of actual folder names
DATASET_FOLDERS = {
    'bidmc': 'bidmc-ppg-and-respiration-dataset-1.0.0',
    'wesad': 'WESAD',
    'capnobase_ieee': 'CapnoBase IEEE TBME Respiratory Rate Benchmark',
    'capnobase_invivo': 'CapnoBase InVivo Dataset',
    'capnobase_event': 'CapnoBase Respiratory Event Benchmark',
    'capnobase_long': 'CapnoBase 8-minute (long) Dataset',
    'capnobase_sim': 'CapnoBase Simulation Dataset',
    'spd': 'SPD',
    'wesd': 'WESD_a-wearable-exam-stress-dataset-for-predicting-cognitive-performance-in-real-world-settings-1.0.0',
    'ppg_dalia': 'PPG_DaLiA_FieldStudy',
    'sleep_edf': 'sleep-edf-database-1.0.0',
    'toadstool': 'Toadstool',
    'ue4w': 'ue4w',
    'stress_nurses': 'stress_detection_nurses_hospital',
    'e4selflearning': 'in-gauge-and-en-gage-understanding-occupants-behaviour-engagement-emotion-and-comfort-indoors-with-heterogeneous-sensors-and-wearables-1.0.0',
    'big_ideas_lab': 'big-ideas-lab-glycemic-variability-and-wearable-device-data-1.1.1',
    'adarp': 'ADARP',
    'weee': 'WEEE',
    'csl': 'CSL Pulse Oximetry Artifact Labels'
}

class DatasetExplorer:
    def __init__(self, raw_data_dir: str):
        self.raw_data_dir = raw_data_dir
        self.report = {
            'timestamp': datetime.now().isoformat(),
            'datasets': {},
            'summary': {},
            'recommendations': []
        }
    
    def explore_all(self):
        print("=" * 80)
        print("🔍 COMPREHENSIVE DATASET EXPLORATION")
        print("=" * 80)
        print(f"Raw data directory: {self.raw_data_dir}")
        print(f"Directory exists: {os.path.exists(self.raw_data_dir)}")
        print("=" * 80)
        
        if not os.path.exists(self.raw_data_dir):
            print("❌ ERROR: Raw data directory not found!")
            return self.report
        
        # Explore each dataset
        self.explore_bidmc()
        self.explore_wesad()
        self.explore_capnobase_all()
        self.explore_spd()
        self.explore_wesd()
        self.explore_ppg_dalia()
        self.explore_toadstool()
        self.explore_ue4w()
        self.explore_sleep_edf()
        self.explore_stress_nurses()
        self.explore_e4selflearning()
        self.explore_big_ideas_lab()
        self.explore_adarp()
        self.explore_weee()
        
        # Generate summary
        self.generate_summary()
        
        # Save report
        self.save_report()
        
        return self.report
    
    def explore_bidmc(self):
        """Explore BIDMC PPG and Respiration Dataset"""
        print("\n" + "=" * 80)
        print("📁 BIDMC DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'BIDMC',
            'folder': DATASET_FOLDERS['bidmc'],
            'subjects': 0,
            'signals': [],
            'sample_data': {}
        }
        
        bidmc_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['bidmc'])
        if not os.path.exists(bidmc_path):
            print(f"⚠️  BIDMC folder not found at: {bidmc_path}")
            self.report['datasets']['bidmc'] = dataset_info
            return
        
        # Count subjects
        all_files = os.listdir(bidmc_path)
        subjects = set()
        for f in all_files:
            if f.startswith('bidmc_') and f.endswith('.csv'):
                parts = f.split('_')
                if len(parts) >= 2:
                    subjects.add(parts[1])
        
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects")
        print(f"📄 Total files: {len(all_files)}")
        
        # Explore first subject
        if subjects:
            first_subj = sorted(subjects)[0]
            print(f"\n--- Exploring subject bidmc_{first_subj} ---")
            
            # Read Numerics.csv
            numerics_file = os.path.join(bidmc_path, f'bidmc_{first_subj}_Numerics.csv')
            if os.path.exists(numerics_file):
                try:
                    df = pd.read_csv(numerics_file, nrows=10)
                    print(f"📊 Numerics shape: {df.shape}")
                    print(f"📋 Columns: {list(df.columns)[:15]}...")
                    
                    # Check for key signals
                    signals_found = []
                    for col in ['HR', 'SpO2', 'RespRate', 'ABP']:
                        if col in df.columns or col.lower() in [c.lower() for c in df.columns]:
                            signals_found.append(col)
                            dataset_info['signals'].append(col)
                    
                    print(f"💓 Available signals: {signals_found}")
                    
                    # Sample stats
                    for col in ['HR', 'SpO2']:
                        if col in df.columns:
                            dataset_info['sample_data'][col.lower()] = {
                                'mean': float(df[col].mean()),
                                'std': float(df[col].std()),
                                'min': float(df[col].min()),
                                'max': float(df[col].max())
                            }
                    
                except Exception as e:
                    print(f"❌ Error reading Numerics: {e}")
            
            # Read Fix.txt for metadata
            fix_file = os.path.join(bidmc_path, f'bidmc_{first_subj}_Fix.txt')
            if os.path.exists(fix_file):
                try:
                    with open(fix_file, 'r', encoding='utf-8') as f:
                        content = f.read(500)
                        print(f"📝 Fix.txt preview:\n{content[:200]}...")
                        
                        # Extract age/gender if available
                        for line in content.split('\n'):
                            if line.startswith('Age:'):
                                dataset_info['sample_data']['age'] = line.split(':')[1].strip()
                            elif line.startswith('Gender:'):
                                dataset_info['sample_data']['gender'] = line.split(':')[1].strip()
                except Exception as e:
                    print(f"❌ Error reading Fix: {e}")
        
        self.report['datasets']['bidmc'] = dataset_info
        print(f"✅ BIDMC exploration complete")
    
    def explore_wesad(self):
        """Explore WESAD Dataset"""
        print("\n" + "=" * 80)
        print("📁 WESAD DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'WESAD',
            'folder': DATASET_FOLDERS['wesad'],
            'subjects': 0,
            'signals': [],
            'sample_data': {}
        }
        
        wesad_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['wesad'])
        if not os.path.exists(wesad_path):
            print(f"⚠️  WESAD folder not found at: {wesad_path}")
            self.report['datasets']['wesad'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(wesad_path) 
                   if d.startswith('S') and os.path.isdir(os.path.join(wesad_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects: {subjects}")
        
        # Explore first 2 subjects
        for subject in subjects[:2]:
            subject_path = os.path.join(wesad_path, subject)
            files = os.listdir(subject_path)
            
            print(f"\n--- Subject {subject} ---")
            print(f"📄 Files: {files}")
            
            # Read PKL
            pkl_file = os.path.join(subject_path, f"{subject}.pkl")
            if os.path.exists(pkl_file):
                try:
                    with open(pkl_file, 'rb') as f:
                        data = pickle.load(f, encoding='latin1')
                    
                    if isinstance(data, dict):
                        print(f"📊 PKL keys: {list(data.keys())[:10]}")
                        
                        if 'signal' in data:
                            signal = data['signal']
                            for sensor in ['chest', 'wrist']:
                                if sensor in signal:
                                    sensor_data = signal[sensor]
                                    if isinstance(sensor_data, dict):
                                        signals = list(sensor_data.keys())
                                        dataset_info['signals'].extend(signals)
                                        print(f"📡 {sensor} signals: {signals[:10]}")
                                        
                                        # Sample HR data
                                        if 'hr' in sensor_data or 'HR' in sensor_data:
                                            hr_data = sensor_data.get('hr') or sensor_data.get('HR')
                                            if hasattr(hr_data, '__len__') and len(hr_data) > 0:
                                                dataset_info['sample_data']['hr'] = {
                                                    'mean': float(np.mean(hr_data[:1000])),
                                                    'std': float(np.std(hr_data[:1000])),
                                                    'length': len(hr_data)
                                                }
                except Exception as e:
                    print(f"❌ Error loading PKL: {e}")
            
            # Read CSV files (S2 has separate CSVs)
            csv_files = [f for f in files if f.endswith('.csv')]
            for csv_file in csv_files[:3]:
                csv_path = os.path.join(subject_path, csv_file)
                try:
                    df = pd.read_csv(csv_path, nrows=5)
                    print(f"📊 {csv_file}: shape={df.shape}, columns={list(df.columns)[:5]}")
                except Exception as e:
                    print(f"❌ Error reading {csv_file}: {e}")
        
        dataset_info['signals'] = list(set(dataset_info['signals']))
        self.report['datasets']['wesad'] = dataset_info
        print(f"✅ WESAD exploration complete")
    
    def explore_capnobase_all(self):
        """Explore all CapnoBase subsets"""
        print("\n" + "=" * 80)
        print("📁 CAPNOBASE DATASETS")
        print("=" * 80)
        
        capno_mapping = {
            'ieee': 'capnobase_ieee',
            'invivo': 'capnobase_invivo',
            'event': 'capnobase_event',
            'long': 'capnobase_long',
            'sim': 'capnobase_sim'
        }
        
        for subset_key, folder_name in capno_mapping.items():
            actual_folder = DATASET_FOLDERS.get(folder_name, folder_name)
            subset_path = os.path.join(self.raw_data_dir, actual_folder)
            
            print(f"\n--- {folder_name} ---")
            
            if not os.path.exists(subset_path):
                print(f"⚠️  Not found: {subset_path}")
                continue
            
            files = [f for f in os.listdir(subset_path) if f.endswith(('.csv', '.tab'))]
            print(f"📄 Files: {len(files)}")
            
            if files:
                # Read sample file
                sample_file = os.path.join(subset_path, files[0])
                try:
                    if sample_file.endswith('.tab'):
                        df = pd.read_csv(sample_file, sep='\t', nrows=5)
                    else:
                        df = pd.read_csv(sample_file, nrows=5)
                    
                    print(f"📊 Sample shape: {df.shape}")
                    print(f"📋 Columns: {list(df.columns)[:10]}")
                    
                    # Check for key signals
                    for col in ['HR', 'SpO2', 'RespRate', 'EtCO2']:
                        if any(col.lower() in c.lower() for c in df.columns):
                            print(f"✅ Found {col}")
                    
                except Exception as e:
                    print(f"❌ Error reading sample: {e}")
        
        self.report['datasets']['capnobase'] = {
            'name': 'CapnoBase',
            'subsets': len(capno_mapping),
            'status': 'explored'
        }
        print(f"✅ CapnoBase exploration complete")
    
    def explore_spd(self):
        """Explore SPD (Stress Detection) Dataset"""
        print("\n" + "=" * 80)
        print("📁 SPD DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'SPD',
            'folder': DATASET_FOLDERS['spd'],
            'subjects': 0,
            'signals': []
        }
        
        spd_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['spd'])
        if not os.path.exists(spd_path):
            print(f"⚠️  SPD folder not found at: {spd_path}")
            self.report['datasets']['spd'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(spd_path) 
                   if d.startswith('S') and os.path.isdir(os.path.join(spd_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects (S01-S{len(subjects):02d})")
        
        # Explore first subject
        if subjects:
            subject_path = os.path.join(spd_path, subjects[0])
            files = os.listdir(subject_path)
            print(f"📄 Files in {subjects[0]}: {files}")
            
            for csv_file in ['HR.csv', 'IBI.csv', 'TEMP.csv', 'EDA.csv', 'BVP.csv', 'ACC.csv']:
                csv_path = os.path.join(subject_path, csv_file)
                if os.path.exists(csv_path):
                    try:
                        df = pd.read_csv(csv_path, nrows=5)
                        print(f"📊 {csv_file}: shape={df.shape}")
                        dataset_info['signals'].append(csv_file.split('.')[0])
                    except Exception as e:
                        print(f"❌ Error reading {csv_file}: {e}")
        
        dataset_info['signals'] = list(set(dataset_info['signals']))
        self.report['datasets']['spd'] = dataset_info
        print(f"✅ SPD exploration complete")
    
    def explore_wesd(self):
        """Explore WESD Dataset"""
        print("\n" + "=" * 80)
        print("📁 WESD DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'WESD',
            'folder': DATASET_FOLDERS['wesd'],
            'subjects': 0,
            'sessions': []
        }
        
        wesd_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['wesd'])
        if not os.path.exists(wesd_path):
            print(f"⚠️  WESD folder not found at: {wesd_path}")
            self.report['datasets']['wesd'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(wesd_path) 
                   if d.startswith('S') and os.path.isdir(os.path.join(wesd_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects")
        
        # Explore first subject
        if subjects:
            subject_path = os.path.join(wesd_path, subjects[0])
            sessions = [d for d in os.listdir(subject_path) if os.path.isdir(os.path.join(subject_path, d))]
            dataset_info['sessions'] = sessions
            print(f"📚 Sessions: {sessions}")
            
            if sessions:
                session_path = os.path.join(subject_path, sessions[0])
                files = os.listdir(session_path)
                print(f"📄 Files in {sessions[0]}: {files}")
        
        self.report['datasets']['wesd'] = dataset_info
        print(f"✅ WESD exploration complete")
    
    def explore_ppg_dalia(self):
        """Explore PPG DaLiA Dataset"""
        print("\n" + "=" * 80)
        print("📁 PPG DaLiA DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'PPG_DaLiA',
            'folder': DATASET_FOLDERS['ppg_dalia'],
            'subjects': 0
        }
        
        dalia_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['ppg_dalia'])
        if not os.path.exists(dalia_path):
            print(f"⚠️  PPG_DaLiA folder not found at: {dalia_path}")
            self.report['datasets']['ppg_dalia'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(dalia_path) 
                   if d.startswith('S') and os.path.isdir(os.path.join(dalia_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects")
        
        self.report['datasets']['ppg_dalia'] = dataset_info
        print(f"✅ PPG DaLiA exploration complete")
    
    def explore_toadstool(self):
        """Explore Toadstool Dataset"""
        print("\n" + "=" * 80)
        print("📁 TOADSTOOL DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'Toadstool',
            'folder': DATASET_FOLDERS['toadstool'],
            'participants': 0
        }
        
        toad_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['toadstool'])
        if not os.path.exists(toad_path):
            print(f"⚠️  Toadstool folder not found at: {toad_path}")
            self.report['datasets']['toadstool'] = dataset_info
            return
        
        participants_path = os.path.join(toad_path, 'participants')
        if os.path.exists(participants_path):
            participants = [d for d in os.listdir(participants_path) 
                          if d.startswith('participant_') and os.path.isdir(os.path.join(participants_path, d))]
            dataset_info['participants'] = len(participants)
            print(f"✅ Found {len(participants)} participants")
        
        self.report['datasets']['toadstool'] = dataset_info
        print(f"✅ Toadstool exploration complete")
    
    def explore_ue4w(self):
        """Explore UE4W Dataset"""
        print("\n" + "=" * 80)
        print("📁 UE4W DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'UE4W',
            'folder': DATASET_FOLDERS['ue4w'],
            'sessions': 0
        }
        
        ue4w_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['ue4w'])
        if not os.path.exists(ue4w_path):
            print(f"⚠️  UE4W folder not found at: {ue4w_path}")
            self.report['datasets']['ue4w'] = dataset_info
            return
        
        sessions = [d for d in os.listdir(ue4w_path) 
                   if os.path.isdir(os.path.join(ue4w_path, d))]
        dataset_info['sessions'] = len(sessions)
        print(f"✅ Found {len(sessions)} sessions")
        
        self.report['datasets']['ue4w'] = dataset_info
        print(f"✅ UE4W exploration complete")
    
    def explore_sleep_edf(self):
        """Explore Sleep EDF Dataset"""
        print("\n" + "=" * 80)
        print("📁 SLEEP EDF DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'Sleep_EDF',
            'folder': DATASET_FOLDERS['sleep_edf'],
            'records': 0
        }
        
        sleep_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['sleep_edf'])
        if not os.path.exists(sleep_path):
            print(f"⚠️  Sleep_EDF folder not found at: {sleep_path}")
            self.report['datasets']['sleep_edf'] = dataset_info
            return
        
        records = [f for f in os.listdir(sleep_path) if f.endswith('.rec')]
        dataset_info['records'] = len(records)
        print(f"✅ Found {len(records)} records")
        
        self.report['datasets']['sleep_edf'] = dataset_info
        print(f"✅ Sleep EDF exploration complete")
    
    def explore_stress_nurses(self):
        """Explore Stress Detection Nurses Dataset"""
        print("\n" + "=" * 80)
        print("📁 STRESS NURSES DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'Stress_Nurses',
            'folder': DATASET_FOLDERS['stress_nurses'],
            'subjects': 0,
            'samples': 0
        }
        
        nurses_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['stress_nurses'])
        if not os.path.exists(nurses_path):
            print(f"⚠️  Stress Nurses folder not found at: {nurses_path}")
            self.report['datasets']['stress_nurses'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(nurses_path) 
                   if os.path.isdir(os.path.join(nurses_path, d))]
        dataset_info['subjects'] = len(subjects)
        
        total_zips = 0
        for subj in subjects:
            subj_path = os.path.join(nurses_path, subj)
            zips = [f for f in os.listdir(subj_path) if f.endswith('.zip')]
            total_zips += len(zips)
        
        dataset_info['samples'] = total_zips
        print(f"✅ Found {len(subjects)} subjects, {total_zips} samples")
        
        self.report['datasets']['stress_nurses'] = dataset_info
        print(f"✅ Stress Nurses exploration complete")
    
    def explore_e4selflearning(self):
        """Explore E4 Self-Learning Dataset"""
        print("\n" + "=" * 80)
        print("📁 E4 SELF-LEARNING DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'E4_SelfLearning',
            'folder': DATASET_FOLDERS['e4selflearning'],
            'participants': 0
        }
        
        e4_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['e4selflearning'])
        if not os.path.exists(e4_path):
            print(f"⚠️  E4 Self-Learning folder not found at: {e4_path}")
            self.report['datasets']['e4selflearning'] = dataset_info
            return
        
        wearable_path = os.path.join(e4_path, 'class_wearable_data')
        if os.path.exists(wearable_path):
            classes = [d for d in os.listdir(wearable_path) if os.path.isdir(os.path.join(wearable_path, d))]
            dataset_info['participants'] = len(classes)
            print(f"✅ Found {len(classes)} class sessions")
        
        self.report['datasets']['e4selflearning'] = dataset_info
        print(f"✅ E4 Self-Learning exploration complete")
    
    def explore_big_ideas_lab(self):
        """Explore Big Ideas Lab Dataset"""
        print("\n" + "=" * 80)
        print("📁 BIG IDEAS LAB DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'Big_Ideas_Lab',
            'folder': DATASET_FOLDERS['big_ideas_lab'],
            'subjects': 0
        }
        
        big_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['big_ideas_lab'])
        if not os.path.exists(big_path):
            print(f"⚠️  Big Ideas Lab folder not found at: {big_path}")
            self.report['datasets']['big_ideas_lab'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(big_path) 
                   if d.isdigit() and os.path.isdir(os.path.join(big_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects")
        
        self.report['datasets']['big_ideas_lab'] = dataset_info
        print(f"✅ Big Ideas Lab exploration complete")
    
    def explore_adarp(self):
        """Explore ADARP Dataset"""
        print("\n" + "=" * 80)
        print("📁 ADARP DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'ADARP',
            'folder': DATASET_FOLDERS['adarp'],
            'parts': 0
        }
        
        adarp_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['adarp'])
        if not os.path.exists(adarp_path):
            print(f"⚠️  ADARP folder not found at: {adarp_path}")
            self.report['datasets']['adarp'] = dataset_info
            return
        
        parts = [d for d in os.listdir(adarp_path) 
                if d.startswith('Part') and os.path.isdir(os.path.join(adarp_path, d))]
        dataset_info['parts'] = len(parts)
        print(f"✅ Found {len(parts)} parts")
        
        self.report['datasets']['adarp'] = dataset_info
        print(f"✅ ADARP exploration complete")
    
    def explore_weee(self):
        """Explore WEEE Dataset"""
        print("\n" + "=" * 80)
        print("📁 WEEE DATASET")
        print("=" * 80)
        
        dataset_info = {
            'name': 'WEEE',
            'folder': DATASET_FOLDERS['weee'],
            'subjects': 0
        }
        
        weee_path = os.path.join(self.raw_data_dir, DATASET_FOLDERS['weee'])
        if not os.path.exists(weee_path):
            print(f"⚠️  WEEE folder not found at: {weee_path}")
            self.report['datasets']['weee'] = dataset_info
            return
        
        subjects = [d for d in os.listdir(weee_path) 
                   if d.startswith('P') and os.path.isdir(os.path.join(weee_path, d))]
        dataset_info['subjects'] = len(subjects)
        print(f"✅ Found {len(subjects)} subjects")
        
        self.report['datasets']['weee'] = dataset_info
        print(f"✅ WEEE exploration complete")
    
    def generate_summary(self):
        """Generate summary statistics"""
        print("\n" + "=" * 80)
        print("📊 GENERATING SUMMARY")
        print("=" * 80)
        
        total_datasets = len(self.report['datasets'])
        total_subjects = sum([d.get('subjects', d.get('participants', d.get('records', d.get('sessions', d.get('samples', 0))))) 
                             for d in self.report['datasets'].values()])
        
        self.report['summary'] = {
            'total_datasets': total_datasets,
            'total_subjects_approx': total_subjects,
            'datasets_with_data': len([d for d in self.report['datasets'].values() if d.get('subjects', 0) > 0]),
            'recommended_for_classifier': ['BIDMC', 'WESAD', 'SPD', 'WESD'],
            'recommended_for_gan': ['WESAD', 'SPD', 'WESD', 'PPG_DaLiA']
        }
        
        print(f"📈 Total datasets: {total_datasets}")
        print(f"👥 Total subjects (approx): {total_subjects}")
        print(f"✅ Datasets with data: {self.report['summary']['datasets_with_data']}")
        
        # Generate recommendations
        self.report['recommendations'] = [
            "✅ BIDMC: Best for HR, SpO2 (53 subjects, hospital quality)",
            "✅ WESAD: Best for stress detection with labels (15 subjects)",
            "✅ SPD: Good for real-world stress (35 subjects)",
            "✅ WESD: Exam stress with sessions (10 subjects)",
            "💡 For classifier: Use BIDMC + WESAD + SPD for 7 features",
            "💡 For GAN: Use WESAD + WESD for training plan templates"
        ]
        
        for rec in self.report['recommendations']:
            print(rec)
    
    def save_report(self):
        """Save exploration report to JSON"""
        os.makedirs(OUTPUT_DIR, exist_ok=True)
        
        with open(REPORT_FILE, 'w', encoding='utf-8') as f:
            json.dump(self.report, f, indent=2, ensure_ascii=False, default=str)
        
        print(f"\n💾 Report saved to: {REPORT_FILE}")
        
        # Also save as CSV summary
        summary_data = []
        for name, info in self.report['datasets'].items():
            summary_data.append({
                'dataset': name,
                'subjects': info.get('subjects', info.get('participants', info.get('records', info.get('sessions', info.get('samples', 0))))),
                'available': '✅' if info.get('subjects', 0) > 0 else '❌'
            })
        
        df_summary = pd.DataFrame(summary_data)
        csv_path = f'{OUTPUT_DIR}/dataset_summary.csv'
        df_summary.to_csv(csv_path, index=False, encoding='utf-8')
        print(f"📊 Summary CSV saved to: {csv_path}")


def main():
    """Main entry point"""
    explorer = DatasetExplorer(RAW_DATA_DIR)
    report = explorer.explore_all()
    
    print("\n" + "=" * 80)
    print("✅ EXPLORATION COMPLETE!")
    print("=" * 80)
    print(f"\nNext steps:")
    print("1. Review: {REPORT_FILE}")
    print("2. Run: cmd/ml-classifier/preprocess_real_data.py")
    print("3. Train: cmd/ml-classifier/train.py")
    
    return report


if __name__ == '__main__':
    main()