import { useState } from 'react';
import api from '../services/api';

export const useBiometric = () => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [history, setHistory] = useState<any[]>([]);

    const submitData = async (data: any) => {
        setLoading(true);
        try {
            await api.post('/biometric', data);
            return true;
        } catch (err: any) {
            setError(err.response?.data?.error || 'Ошибка отправки');
            return false;
        } finally {
            setLoading(false);
        }
    };

    const getHistory = async (userId: string, from?: string, to?: string) => {
        setLoading(true);
        try {
            const response = await api.get(`/biometric/${userId}`, { params: { from, to } });
            setHistory(response.data);
            return response.data;
        } catch (err: any) {
            setError(err.response?.data?.error || 'Ошибка получения истории');
            return [];
        } finally {
            setLoading(false);
        }
    };

    return { loading, error, history, submitData, getHistory };
};