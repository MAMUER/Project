// Doctor API Client — работа с doctor-service через gateway
class DoctorAPI {
    constructor(baseURL = '') {
        this.baseURL = baseURL;
    }

    // ===== Doctors =====
    async listDoctors(page = 1, specialty = '') {
        const params = new URLSearchParams({ page, page_size: 20 });
        if (specialty) params.set('specialty', specialty);
        return this._get(`/api/v1/doctors?${params}`);
    }

    async getDoctor(doctorId) {
        return this._get(`/api/v1/doctors/${doctorId}`);
    }

    // ===== Subscriptions =====
    async subscribe(userId, doctorId, planType = 'monthly') {
        return this._post('/api/v1/doctor/subscribe', {
            user_id: userId,
            doctor_id: doctorId,
            plan_type: planType
        });
    }

    async getSubscription(userId, doctorId) {
        return this._get(`/api/v1/doctor/subscription?user_id=${userId}&doctor_id=${doctorId}`);
    }

    async cancelSubscription(userId, doctorId) {
        return this._post('/api/v1/doctor/subscription/cancel', {
            user_id: userId,
            doctor_id: doctorId
        });
    }

    // ===== Messages =====
    async sendMessage(userId, doctorId, senderId, senderType, message, messageType = 'text') {
        return this._post('/api/v1/doctor/message', {
            user_id: userId,
            doctor_id: doctorId,
            sender_id: senderId,
            sender_type: senderType,
            message,
            message_type: messageType
        });
    }

    async getChatHistory(userId, doctorId, page = 1, pageSize = 50) {
        return this._get(`/api/v1/doctor/messages?user_id=${userId}&doctor_id=${doctorId}&page=${page}&page_size=${pageSize}`);
    }

    async markMessagesRead(userId, doctorId) {
        return this._post('/api/v1/doctor/messages/read', {
            user_id: userId,
            doctor_id: doctorId
        });
    }

    async getUnreadCount(userId) {
        return this._get(`/api/v1/doctor/messages/unread?user_id=${userId}`);
    }

    // ===== Prescriptions =====
    async createPrescription(userId, doctorId, type, title, description, priority = 'normal') {
        return this._post('/api/v1/doctor/prescription', {
            user_id: userId,
            doctor_id: doctorId,
            prescription_type: type,
            title,
            description,
            priority
        });
    }

    async getPrescriptions(userId, statusFilter = 'active') {
        return this._get(`/api/v1/doctor/prescriptions?user_id=${userId}&status_filter=${statusFilter}`);
    }

    async updatePrescriptionStatus(prescriptionId, newStatus) {
        return this._post(`/api/v1/doctor/prescription/${prescriptionId}/status`, {
            new_status: newStatus
        });
    }

    // ===== Training Modifications =====
    async modifyTrainingPlan(userId, doctorId, trainingPlanId, modificationType, oldValue, newValue, reason) {
        return this._post('/api/v1/doctor/training/modify', {
            user_id: userId,
            doctor_id: doctorId,
            training_plan_id: trainingPlanId,
            modification_type: modificationType,
            old_value: oldValue,
            new_value: newValue,
            reason
        });
    }

    async getTrainingModifications(userId, trainingPlanId = '') {
        const params = new URLSearchParams({ user_id: userId });
        if (trainingPlanId) params.set('training_plan_id', trainingPlanId);
        return this._get(`/api/v1/doctor/training/modifications?${params}`);
    }

    // ===== Consultations =====
    async scheduleConsultation(userId, doctorId, scheduledAt) {
        return this._post('/api/v1/doctor/consultation/schedule', {
            user_id: userId,
            doctor_id: doctorId,
            scheduled_at: scheduledAt
        });
    }

    async getConsultations(userId, statusFilter = 'all') {
        return this._get(`/api/v1/doctor/consultations?user_id=${userId}&status_filter=${statusFilter}`);
    }

    async completeConsultation(consultationId, notes = '') {
        return this._post(`/api/v1/doctor/consultation/${consultationId}/complete`, {
            notes
        });
    }

    // ===== Helper =====
    async _get(url) {
        const token = localStorage.getItem('access_token');
        const resp = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        if (!resp.ok) throw new Error(`API error: ${resp.status}`);
        return resp.json();
    }

    async _post(url, data) {
        const token = localStorage.getItem('access_token');
        const resp = await fetch(url, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        if (!resp.ok) throw new Error(`API error: ${resp.status}`);
        return resp.json();
    }
}

// Export singleton
window.doctorAPI = new DoctorAPI();
