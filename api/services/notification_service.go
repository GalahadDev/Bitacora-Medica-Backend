package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/google/uuid"
)

type NotificationService struct {
	cfg *config.Config
}

func NewNotificationService(cfg *config.Config) *NotificationService {
	return &NotificationService{cfg: cfg}
}

func (s *NotificationService) getNameFromProfile(jsonData interface{}) string {

	type Profile struct {
		FullName  string `json:"full_name"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Name      string `json:"name"`
	}

	var p Profile
	bytes, _ := json.Marshal(jsonData)
	_ = json.Unmarshal(bytes, &p)

	if p.FullName != "" {
		return p.FullName
	}
	if p.FirstName != "" {
		return p.FirstName + " " + p.LastName
	}
	if p.Name != "" {
		return p.Name
	}
	return "Usuario"
}

func (s *NotificationService) getHTMLTemplate(title, bodyContent, actionButton, accentColor string) string {
	if accentColor == "" {
		accentColor = "#2563eb"
	}

	header := fmt.Sprintf(`
		<div style="background-color: %s; padding: 24px; text-align: center;">
			<h1 style="color: white; margin: 0; font-family: 'Segoe UI', sans-serif; font-size: 24px;">MedLog Digital</h1>
		</div>`, accentColor)

	btn := ""
	if actionButton != "" {
		btn = fmt.Sprintf(`<div style="text-align: center; margin: 30px 0;">%s</div>`, actionButton)
	}

	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="margin:0; padding:0; background-color:#f3f4f6; font-family:'Segoe UI', sans-serif; color:#374151;">
			<div style="max-width:600px; margin:20px auto; background-color:white; border-radius:8px; overflow:hidden; box-shadow:0 4px 6px rgba(0,0,0,0.05);">
				%s
				<div style="padding:40px; line-height:1.6;">
					<h2 style="color:#111827; margin-top:0; border-bottom:1px solid #eee; padding-bottom:10px;">%s</h2>
					%s
					%s
				</div>
				<div style="background-color:#f9fafb; padding:20px; text-align:center; font-size:12px; color:#9ca3af;">
					<p>© 2026 MedLog. Notificación automática.</p>
				</div>
			</div>
		</body>
		</html>
	`, header, title, bodyContent, btn)
}

func (s *NotificationService) createAndNotify(userID uuid.UUID, notifType string, subject string, textSummary string, htmlBody string, relatedID *uuid.UUID) {
	db := database.GetDB()

	notif := domains.Notification{
		UserID:    userID,
		Type:      notifType,
		Message:   textSummary, // Texto plano para la UI
		RelatedID: relatedID,
		IsRead:    false,
	}

	go func() {
		if err := db.Create(&notif).Error; err != nil {
			slog.Error("Failed to save notification DB", "error", err)
		}

		var user domains.User
		if err := db.Select("email").First(&user, "id = ?", userID).Error; err != nil {
			slog.Error("User email not found", "userID", userID)
			return
		}

		s.sendRealEmail(user.Email, subject, htmlBody)
	}()
}

func (s *NotificationService) sendRealEmail(to string, subject string, htmlBody string) {
	auth := smtp.PlainAuth("", s.cfg.SMTPEmail, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("To: " + to + "\r\n" +
		"From: MedLog <" + s.cfg.SMTPEmail + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		headers + "\r\n" +
		htmlBody)

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	if err := smtp.SendMail(addr, auth, s.cfg.SMTPEmail, []string{to}, msg); err != nil {
		slog.Error("❌ Email failed", "to", to, "error", err)
	} else {
		slog.Info("✅ Email sent", "to", to)
	}
}

func (s *NotificationService) NotifyNewUser(newUserID uuid.UUID, userEmail string) {
	var admins []domains.User
	database.GetDB().Where("role = ?", domains.RoleAdmin).Find(&admins)

	subject := "Nuevo Registro en Plataforma"
	summary := fmt.Sprintf("Usuario %s registrado.", userEmail)

	body := fmt.Sprintf(`
		<p>Un nuevo profesional se ha registrado y requiere verificación.</p>
		<p><strong>Email:</strong> %s</p>
		<p>Por favor, ingrese al panel administrativo para aprobar o rechazar la solicitud.</p>
	`, userEmail)

	btn := `<a href="#" style="background-color:#2563eb; color:white; padding:10px 20px; text-decoration:none; border-radius:5px;">Ir al Panel Admin</a>`
	html := s.getHTMLTemplate("Solicitud de Acceso", body, btn, "#2563eb")

	for _, admin := range admins {
		s.createAndNotify(admin.ID, "NEW_USER", subject, summary, html, &newUserID)
	}
}

func (s *NotificationService) NotifyAccountStatus(userID uuid.UUID, status domains.UserStatus, reason string) {
	var subject, title, color, message string

	if status == domains.StatusActive {
		subject = "¡Cuenta Aprobada!"
		title = "Bienvenido a MedLog"
		color = "#16a34a"
		message = "Tu cuenta ha sido verificada exitosamente. Ya puedes acceder a todas las funcionalidades."
	} else {
		subject = "Estado de tu cuenta"
		title = "Solicitud Rechazada"
		color = "#dc2626"
		message = fmt.Sprintf("Lamentamos informarte que tu solicitud no ha sido aprobada.<br/><br/><strong>Motivo:</strong> %s", reason)
	}

	html := s.getHTMLTemplate(title, fmt.Sprintf("<p>%s</p>", message), "", color)
	s.createAndNotify(userID, "ACCOUNT_STATUS", subject, "Tu cuenta ha sido "+string(status), html, nil)
}

func (s *NotificationService) NotifyIncident(patientID uuid.UUID, incidentDetails string) {
	db := database.GetDB()
	var patient domains.Patient

	db.Preload("PersonalInfo").First(&patient, "id = ?", patientID)

	patientName := "Paciente ID " + patientID.String()
	pName := s.getNameFromProfile(patient.PersonalInfo)
	if pName != "Usuario" {
		patientName = pName
	}

	var collaborators []domains.User
	db.Table("users").
		Joins("JOIN collaborations ON collaborations.professional_id = users.id").
		Where("collaborations.patient_id = ? AND collaborations.status = ?", patientID, domains.CollabAccepted).
		Find(&collaborators)

	var creator domains.User
	db.First(&creator, "id = ?", patient.CreatorID)

	recipients := append(collaborators, creator)
	uniqueUsers := make(map[string]domains.User)
	for _, u := range recipients {
		uniqueUsers[u.ID.String()] = u
	}

	subject := "⚠️ ALERTA: Incidente con " + patientName
	summary := "Incidente reportado para " + patientName

	body := fmt.Sprintf(`
		<p style="color:#b91c1c;"><strong>Se ha reportado un evento adverso.</strong></p>
		<p><strong>Paciente:</strong> %s</p>
		<div style="background-color:#fee2e2; border-left:4px solid #dc2626; padding:15px; margin:20px 0; color:#7f1d1d;">
			<strong>Detalle:</strong><br/>%s
		</div>
		<p>Por favor, revise la bitácora antes de la próxima intervención.</p>
	`, patientName, incidentDetails)

	html := s.getHTMLTemplate("Reporte de Incidente", body, "", "#dc2626")

	for _, professional := range uniqueUsers {
		s.createAndNotify(professional.ID, "INCIDENT_ALERT", subject, summary, html, &patientID)
	}
}

func (s *NotificationService) NotifyCollabInvite(invitedUserID uuid.UUID, patientID uuid.UUID, inviterName string) {
	subject := "Invitación a Colaborar"
	summary := fmt.Sprintf("%s te ha invitado a un equipo médico.", inviterName)

	body := fmt.Sprintf(`
		<p>El profesional <strong>%s</strong> te ha invitado a colaborar en un expediente clínico.</p>
		<p>Ingresa a la plataforma para aceptar o rechazar la solicitud.</p>
	`, inviterName)

	btn := `<a href="#" style="background-color:#2563eb; color:white; padding:10px 20px; text-decoration:none; border-radius:5px;">Ver Invitaciones</a>`
	html := s.getHTMLTemplate("Nueva Colaboración", body, btn, "#2563eb")

	s.createAndNotify(invitedUserID, "COLLAB_INVITE", subject, summary, html, &patientID)
}

func (s *NotificationService) NotifyInviteResponse(creatorID uuid.UUID, responderEmail string, status domains.CollabStatus) {
	subject := "Respuesta a Invitación"
	summary := fmt.Sprintf("%s ha %s tu invitación.", responderEmail, status)

	color := "#2563eb"
	if status == domains.CollabAccepted {
		color = "#16a34a"
	}
	if status == domains.CollabRejected {
		color = "#dc2626"
	}

	body := fmt.Sprintf(`<p>El profesional <strong>%s</strong> ha respondido a tu invitación con el estado: <strong>%s</strong>.</p>`, responderEmail, status)
	html := s.getHTMLTemplate("Actualización de Equipo", body, "", color)

	s.createAndNotify(creatorID, "INVITE_RESPONSE", subject, summary, html, nil)
}

func (s *NotificationService) NotifyTicketReply(userID uuid.UUID, ticketSubject string, reply string) {
	subject := "Respuesta a tu Ticket de Soporte"
	summary := "Admin ha respondido a: " + ticketSubject

	body := fmt.Sprintf(`
		<p>Hola,</p>
		<p>Hemos respondido a tu solicitud de soporte: <strong>"%s"</strong></p>
		<div style="background-color:#f0fdf4; border:1px solid #bbf7d0; padding:15px; border-radius:6px; margin-top:10px;">
			<strong>Respuesta:</strong><br/>
			%s
		</div>
	`, ticketSubject, reply)

	html := s.getHTMLTemplate("Soporte MedLog", body, "", "#16a34a")

	s.createAndNotify(userID, "TICKET_REPLY", subject, summary, html, nil)
}

func (s *NotificationService) NotifyTicketCreated(userEmail string, ticketSubject string) {
	var admins []domains.User
	database.GetDB().Where("role = ?", domains.RoleAdmin).Find(&admins)

	subject := "Nuevo Ticket de Soporte"
	summary := "Ticket de " + userEmail

	body := fmt.Sprintf(`
		<p>El usuario <strong>%s</strong> ha abierto un nuevo ticket.</p>
		<p><strong>Asunto:</strong> %s</p>
	`, userEmail, ticketSubject)

	html := s.getHTMLTemplate("Mesa de Ayuda", body, "", "#7c3aed")

	for _, admin := range admins {
		s.createAndNotify(admin.ID, "TICKET_CREATED", subject, summary, html, nil)
	}
}
