// ML Classification and Plan Generation helpers

async function loadBiometricParams() {
    try {
        const records = await getBiometricRecords('', null, null, 50);

        let hr = 0, hrCount = 0;
        let hrv = 50;
        let spo2 = 0, spo2Count = 0;
        let temp = 0, tempCount = 0;
        let bp = 0, bpCount = 0;

        records.records?.forEach(rec => {
            switch (rec.metric_type) {
                case 'heart_rate':
                    hr += rec.value;
                    hrCount++;
                    break;
                case 'hrv':
                    hrv = rec.value;
                    break;
                case 'spo2':
                    spo2 += rec.value;
                    spo2Count++;
                    break;
                case 'temperature':
                    temp += rec.value;
                    tempCount++;
                    break;
                case 'blood_pressure':
                    bp += rec.value;
                    bpCount++;
                    break;
            }
        });

        const hrEl = document.getElementById('hrValue');
        const hrvEl = document.getElementById('hrvValue');
        const spo2El = document.getElementById('spo2Value');
        const tempEl = document.getElementById('tempValue');
        const bpEl = document.getElementById('bpValue');

        if (hrEl) hrEl.textContent = hrCount > 0 ? Math.round(hr / hrCount) : '--';
        if (hrvEl) hrvEl.textContent = Math.round(hrv);
        if (spo2El) spo2El.textContent = spo2Count > 0 ? Math.round(spo2 / spo2Count) : '--';
        if (tempEl) tempEl.textContent = tempCount > 0 ? (temp / tempCount).toFixed(1) : '--';
        if (bpEl) bpEl.textContent = bpCount > 0 ? Math.round(bp / bpCount) + '/80' : '--';

    } catch (error) {
        console.error('Failed to load biometric params:', error);
    }
}
