// Package grpcadapter implements the gRPC server adapter for the doctor service.
// It translates gRPC requests into domain service calls and vice versa.
package grpcadapter

import (
	"context"

	doctorpb "github.com/MAMUER/Project/api/gen/doctor"
	"github.com/MAMUER/Project/internal/doctor/model"
	"github.com/MAMUER/Project/internal/doctor/port"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the gRPC server for the doctor service.
type Handler struct {
	doctorpb.UnimplementedDoctorServiceServer
	svc port.DoctorService
}

// New creates a new gRPC handler with the given service implementation.
func New(svc port.DoctorService) *Handler {
	return &Handler{svc: svc}
}

// ===== Doctors =====

func (h *Handler) ListDoctors(ctx context.Context, req *doctorpb.ListDoctorsRequest) (*doctorpb.ListDoctorsResponse, error) {
	doctors, total, err := h.svc.ListDoctors(ctx, req.GetSpecialty(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list doctors")
	}

	pbDoctors := make([]*doctorpb.Doctor, len(doctors))
	for i, d := range doctors {
		pbDoctors[i] = toProtoDoctor(&d)
	}

	return &doctorpb.ListDoctorsResponse{
		Doctors: pbDoctors,
		Total:   safeInt64ToInt32(total),
	}, nil
}

func (h *Handler) GetDoctor(ctx context.Context, req *doctorpb.GetDoctorRequest) (*doctorpb.Doctor, error) {
	d, err := h.svc.GetDoctor(ctx, req.GetDoctorId())
	if err != nil {
		if err.Error() == "doctor not found" {
			return nil, status.Error(codes.NotFound, "doctor not found")
		}
		return nil, status.Error(codes.Internal, "failed to get doctor")
	}
	return toProtoDoctor(d), nil
}

// ===== Subscriptions =====

func (h *Handler) SubscribeToDoctor(ctx context.Context, req *doctorpb.SubscribeRequest) (*doctorpb.SubscribeResponse, error) {
	sub, err := h.svc.Subscribe(ctx, req.GetUserId(), req.GetDoctorId(), req.GetPlanType())
	if err != nil {
		if err.Error() == "invalid plan type" {
			return nil, status.Error(codes.InvalidArgument, "invalid plan type")
		}
		return nil, status.Error(codes.Internal, "failed to create subscription")
	}

	return &doctorpb.SubscribeResponse{
		SubscriptionId: sub.ID,
		ExpiresAt:      timestamppb.New(sub.ExpiresAt),
		Price:          sub.Price,
	}, nil
}

func (h *Handler) GetSubscription(ctx context.Context, req *doctorpb.GetSubscriptionRequest) (*doctorpb.Subscription, error) {
	sub, err := h.svc.GetSubscription(ctx, req.GetUserId(), req.GetDoctorId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "subscription not found")
	}
	return toProtoSubscription(sub), nil
}

func (h *Handler) CancelSubscription(ctx context.Context, req *doctorpb.CancelSubscriptionRequest) (*doctorpb.CancelSubscriptionResponse, error) {
	err := h.svc.CancelSubscription(ctx, req.GetUserId(), req.GetDoctorId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "active subscription not found")
	}

	return &doctorpb.CancelSubscriptionResponse{
		Success: true,
		Message: "Подписка отменена",
	}, nil
}

// ===== Messages =====

func (h *Handler) SendMessage(ctx context.Context, req *doctorpb.SendMessageRequest) (*doctorpb.Message, error) {
	msg := &model.Message{
		UserID:      req.GetUserId(),
		DoctorID:    req.GetDoctorId(),
		Message:     req.GetMessage(),
		MessageType: req.GetMessageType(),
	}

	if req.GetSenderId() != "" {
		senderUser := req.GetSenderId()
		msg.SenderUserID = &senderUser
	}

	result, err := h.svc.SendMessage(ctx, msg)
	if err != nil {
		if err.Error() == "active subscription required" {
			return nil, status.Error(codes.PermissionDenied, "active subscription required")
		}
		return nil, status.Error(codes.Internal, "failed to send message")
	}

	return toProtoMessage(result), nil
}

func (h *Handler) GetChatHistory(ctx context.Context, req *doctorpb.GetChatHistoryRequest) (*doctorpb.GetChatHistoryResponse, error) {
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 50
	}

	messages, err := h.svc.GetChatHistory(ctx, req.GetUserId(), req.GetDoctorId(), req.GetPage(), pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get chat history")
	}

	pbMessages := make([]*doctorpb.Message, len(messages))
	for i, m := range messages {
		pbMessages[i] = toProtoMessage(&m)
	}

	return &doctorpb.GetChatHistoryResponse{
		Messages: pbMessages,
		Total:    safeIntToInt32(len(messages)),
	}, nil
}

func (h *Handler) MarkMessagesRead(ctx context.Context, req *doctorpb.MarkMessagesReadRequest) (*doctorpb.MarkMessagesReadResponse, error) {
	count, err := h.svc.MarkMessagesRead(ctx, req.GetUserId(), req.GetDoctorId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark messages read")
	}

	return &doctorpb.MarkMessagesReadResponse{
		MarkedCount: safeInt64ToInt32(count),
	}, nil
}

func (h *Handler) GetUnreadCount(ctx context.Context, req *doctorpb.GetUnreadCountRequest) (*doctorpb.GetUnreadCountResponse, error) {
	byDoctor, total, err := h.svc.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get unread count")
	}

	return &doctorpb.GetUnreadCountResponse{
		UnreadByDoctor: mapInt64ToInt32(byDoctor),
		TotalUnread:    safeInt64ToInt32(total),
	}, nil
}

// ===== Prescriptions =====

func (h *Handler) CreatePrescription(ctx context.Context, req *doctorpb.CreatePrescriptionRequest) (*doctorpb.Prescription, error) {
	p := &model.Prescription{
		UserID:           req.GetUserId(),
		DoctorID:         req.GetDoctorId(),
		PrescriptionType: req.GetPrescriptionType(),
		Title:            req.GetTitle(),
		Description:      req.GetDescription(),
		Priority:         req.GetPriority(),
	}

	result, err := h.svc.CreatePrescription(ctx, p)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create prescription")
	}

	return toProtoPrescription(result), nil
}

func (h *Handler) GetPrescriptions(ctx context.Context, req *doctorpb.GetPrescriptionsRequest) (*doctorpb.GetPrescriptionsResponse, error) {
	prescriptions, err := h.svc.GetPrescriptions(ctx, req.GetUserId(), req.GetStatusFilter())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get prescriptions")
	}

	pbPrescriptions := make([]*doctorpb.Prescription, len(prescriptions))
	for i, p := range prescriptions {
		pbPrescriptions[i] = toProtoPrescription(&p)
	}

	return &doctorpb.GetPrescriptionsResponse{
		Prescriptions: pbPrescriptions,
		Total:         safeIntToInt32(len(prescriptions)),
	}, nil
}

func (h *Handler) UpdatePrescriptionStatus(ctx context.Context, req *doctorpb.UpdatePrescriptionStatusRequest) (*doctorpb.UpdatePrescriptionStatusResponse, error) {
	err := h.svc.UpdatePrescriptionStatus(ctx, req.GetPrescriptionId(), req.GetNewStatus())
	if err != nil {
		return nil, status.Error(codes.NotFound, "prescription not found")
	}

	return &doctorpb.UpdatePrescriptionStatusResponse{
		Success: true,
		Message: "Статус обновлен",
	}, nil
}

// ===== Training Modifications =====

func (h *Handler) ModifyTrainingPlan(ctx context.Context, req *doctorpb.ModifyTrainingPlanRequest) (*doctorpb.ModifyTrainingPlanResponse, error) {
	tm := &model.TrainingModification{
		DoctorID:         req.GetDoctorId(),
		TrainingPlanID:   req.GetTrainingPlanId(),
		ModificationType: req.GetModificationType(),
		OldValue:         structpbToString(req.GetOldValue()),
		NewValue:         structpbToString(req.GetNewValue()),
		Reason:           req.GetReason(),
	}

	err := h.svc.CreateTrainingModification(ctx, tm)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create training modification")
	}

	return &doctorpb.ModifyTrainingPlanResponse{
		ModificationId: tm.ID,
		Success:        true,
		Message:        "Тренировочный план обновлен",
	}, nil
}

func (h *Handler) GetTrainingModifications(ctx context.Context, req *doctorpb.GetTrainingModificationsRequest) (*doctorpb.GetTrainingModificationsResponse, error) {
	modifications, err := h.svc.GetTrainingModifications(ctx, req.GetUserId(), req.GetTrainingPlanId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get modifications")
	}

	pbModifications := make([]*doctorpb.TrainingModification, len(modifications))
	for i, m := range modifications {
		pbModifications[i] = toProtoTrainingModification(&m)
	}

	return &doctorpb.GetTrainingModificationsResponse{
		Modifications: pbModifications,
		Total:         safeIntToInt32(len(modifications)),
	}, nil
}

// ===== Consultations =====

func (h *Handler) ScheduleConsultation(ctx context.Context, req *doctorpb.ScheduleConsultationRequest) (*doctorpb.Consultation, error) {
	c := &model.Consultation{
		UserID:      req.GetUserId(),
		DoctorID:    req.GetDoctorId(),
		ScheduledAt: req.GetScheduledAt().AsTime(),
	}

	result, err := h.svc.ScheduleConsultation(ctx, c)
	if err != nil {
		if err.Error() == "active subscription required" {
			return nil, status.Error(codes.PermissionDenied, "active subscription required")
		}
		return nil, status.Error(codes.Internal, "failed to schedule consultation")
	}

	return toProtoConsultation(result), nil
}

func (h *Handler) GetConsultations(ctx context.Context, req *doctorpb.GetConsultationsRequest) (*doctorpb.GetConsultationsResponse, error) {
	consultations, err := h.svc.GetConsultations(ctx, req.GetUserId(), req.GetStatusFilter())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get consultations")
	}

	pbConsultations := make([]*doctorpb.Consultation, len(consultations))
	for i, c := range consultations {
		pbConsultations[i] = toProtoConsultation(&c)
	}

	return &doctorpb.GetConsultationsResponse{
		Consultations: pbConsultations,
		Total:         safeIntToInt32(len(consultations)),
	}, nil
}

func (h *Handler) CompleteConsultation(ctx context.Context, req *doctorpb.CompleteConsultationRequest) (*doctorpb.CompleteConsultationResponse, error) {
	err := h.svc.CompleteConsultation(ctx, req.GetConsultationId(), req.GetNotes())
	if err != nil {
		return nil, status.Error(codes.NotFound, "consultation not found or already completed")
	}

	return &doctorpb.CompleteConsultationResponse{
		Success: true,
		Message: "Консультация завершена",
	}, nil
}

// ===== Proto Converters =====

func toProtoDoctor(d *model.Doctor) *doctorpb.Doctor {
	return &doctorpb.Doctor{
		Id:            d.ID,
		Specialty:     d.Specialty,
		LicenseNumber: d.LicenseNumber,
		Phone:         d.Phone,
		Bio:           d.Bio,
		IsActive:      d.IsActive,
	}
}

func toProtoSubscription(s *model.Subscription) *doctorpb.Subscription {
	return &doctorpb.Subscription{
		Id:        s.ID,
		UserId:    s.UserID,
		DoctorId:  s.DoctorID,
		PlanType:  s.PlanType,
		StartsAt:  timestamppb.New(s.StartsAt),
		ExpiresAt: timestamppb.New(s.ExpiresAt),
		IsActive:  s.IsActive,
		Price:     s.Price,
	}
}

func toProtoMessage(m *model.Message) *doctorpb.Message {
	pb := &doctorpb.Message{
		Id:          m.ID,
		UserId:      m.UserID,
		DoctorId:    m.DoctorID,
		Message:     m.Message,
		MessageType: m.MessageType,
		IsRead:      m.IsRead,
		CreatedAt:   timestamppb.New(m.CreatedAt),
	}
	if m.SenderUserID != nil {
		pb.SenderId = *m.SenderUserID
		pb.SenderType = "user"
	}
	if m.SenderDoctorID != nil {
		pb.SenderId = *m.SenderDoctorID
		pb.SenderType = "doctor"
	}
	return pb
}

func toProtoPrescription(p *model.Prescription) *doctorpb.Prescription {
	return &doctorpb.Prescription{
		Id:               p.ID,
		UserId:           p.UserID,
		DoctorId:         p.DoctorID,
		PrescriptionType: p.PrescriptionType,
		Title:            p.Title,
		Description:      p.Description,
		Priority:         p.Priority,
		Status:           p.Status,
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
	}
}

func toProtoTrainingModification(tm *model.TrainingModification) *doctorpb.TrainingModification {
	return &doctorpb.TrainingModification{
		Id:               tm.ID,
		DoctorId:         tm.DoctorID,
		TrainingPlanId:   tm.TrainingPlanID,
		ModificationType: tm.ModificationType,
		OldValue:         stringToStructpb(tm.OldValue),
		NewValue:         stringToStructpb(tm.NewValue),
		Reason:           tm.Reason,
		CreatedAt:        timestamppb.New(tm.CreatedAt),
	}
}

func toProtoConsultation(c *model.Consultation) *doctorpb.Consultation {
	pb := &doctorpb.Consultation{
		Id:          c.ID,
		UserId:      c.UserID,
		DoctorId:    c.DoctorID,
		Status:      c.Status,
		ScheduledAt: timestamppb.New(c.ScheduledAt),
		CreatedAt:   timestamppb.New(c.CreatedAt),
		Notes:       c.Notes,
	}
	if c.StartedAt != nil {
		pb.StartedAt = timestamppb.New(*c.StartedAt)
	}
	if c.EndedAt != nil {
		pb.EndedAt = timestamppb.New(*c.EndedAt)
	}
	return pb
}

// ===== Helpers =====

func safeInt64ToInt32(v int64) int32 {
	if v > 2147483647 {
		return 2147483647
	}
	if v < -2147483648 {
		return -2147483648
	}
	return int32(v)
}

func safeIntToInt32(v int) int32 {
	if v > 2147483647 {
		return 2147483647
	}
	if v < -2147483648 {
		return -2147483648
	}
	return int32(v)
}

func mapInt64ToInt32(m map[string]int64) map[string]int32 {
	result := make(map[string]int32, len(m))
	for k, v := range m {
		result[k] = safeInt64ToInt32(v)
	}
	return result
}

func structpbToString(s *structpb.Struct) string {
	if s == nil {
		return ""
	}
	b, _ := s.MarshalJSON()
	return string(b)
}

func stringToStructpb(s string) *structpb.Struct {
	if s == "" {
		return nil
	}
	result, _ := structpb.NewStruct(map[string]interface{}{"value": s})
	return result
}
