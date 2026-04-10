// Doctor API Client — работа с doctor-service через gateway
// Uses the shared apiRequest() from api.js for consistency

class DoctorAPI {
    async listDoctors(page = 1, specialty = '') {
        const params = new URLSearchParams({ page, page_size: 20 });
        if (specialty) params.set('specialty', specialty);
        return window.apiRequest(`/doctors?${params}`);
    }

    async getDoctor(doctorId) {
        return window.apiRequest(`/doctors/${doctorId}`);
    }

    async subscribe(userId, doctorId, planType = 'monthly') {
        return window.apiRequest('/doctor/subscribe', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId, doctor_id: doctorId, plan_type: planType })
        });
    }

    async getSubscription(userId, doctorId) {
        return window.apiRequest(`/doctor/subscription?user_id=${userId}&doctor_id=${doctorId}`);
    }

    async cancelSubscription(userId, doctorId) {
        return window.apiRequest('/doctor/subscription/cancel', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId, doctor_id: doctorId })
        });
    }

    async sendMessage(userId, doctorId, senderId, senderType, message, messageType = 'text') {
        return window.apiRequest('/doctor/message', {
            method: 'POST',
            body: JSON.stringify({
                user_id: userId, doctor_id: doctorId,
                sender_id: senderId, sender_type: senderType,
                message, message_type: messageType
            })
        });
    }

    async getChatHistory(userId, doctorId, page = 1, pageSize = 50) {
        return window.apiRequest(`/doctor/messages?user_id=${userId}&doctor_id=${doctorId}&page=${page}&page_size=${pageSize}`);
    }

    async markMessagesRead(userId, doctorId) {
        return window.apiRequest('/doctor/messages/read', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId, doctor_id: doctorId })
        });
    }

    async getUnreadCount(userId) {
        return window.apiRequest(`/doctor/messages/unread?user_id=${userId}`);
    }

    async createPrescription(userId, doctorId, type, title, description, priority = 'normal') {
        return window.apiRequest('/doctor/prescription', {
            method: 'POST',
            body: JSON.stringify({
                user_id: userId, doctor_id: doctorId,
                prescription_type: type, title, description, priority
            })
        });
    }

    async getPrescriptions(userId, statusFilter = 'active') {
        return window.apiRequest(`/doctor/prescriptions?user_id=${userId}&status_filter=${statusFilter}`);
    }

    async updatePrescriptionStatus(prescriptionId, newStatus) {
        return window.apiRequest(`/doctor/prescription/${prescriptionId}/status`, {
            method: 'POST',
            body: JSON.stringify({ new_status: newStatus })
        });
    }

    async modifyTrainingPlan(userId, doctorId, trainingPlanId, modificationType, oldValue, newValue, reason) {
        return window.apiRequest('/doctor/training/modify', {
            method: 'POST',
            body: JSON.stringify({
                user_id: userId, doctor_id: doctorId,
                training_plan_id: trainingPlanId,
                modification_type: modificationType,
                old_value: oldValue, new_value: newValue, reason
            })
        });
    }

    async getTrainingModifications(userId, trainingPlanId = '') {
        const params = new URLSearchParams({ user_id: userId });
        if (trainingPlanId) params.set('training_plan_id', trainingPlanId);
        return window.apiRequest(`/doctor/training/modifications?${params}`);
    }

    async scheduleConsultation(userId, doctorId, scheduledAt) {
        return window.apiRequest('/doctor/consultation/schedule', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId, doctor_id: doctorId, scheduled_at: scheduledAt })
        });
    }

    async getConsultations(userId, statusFilter = 'all') {
        return window.apiRequest(`/doctor/consultations?user_id=${userId}&status_filter=${statusFilter}`);
    }

    async completeConsultation(consultationId, notes = '') {
        return window.apiRequest(`/doctor/consultation/${consultationId}/complete`, {
            method: 'POST',
            body: JSON.stringify({ notes })
        });
    }
}

// Export singleton
window.doctorAPI = new DoctorAPI();
