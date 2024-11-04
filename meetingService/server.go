package meetingservice

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// server - структура, которая реализует интерфейс MeetingServiceServer и хранит подключение к базе данных
type server struct {
	UnimplementedMeetingServiceServer
	db *sql.DB
}

// NewServer - функция для создания нового экземпляра сервера с подключением к базе данных
func NewServer(db *sql.DB) *server {
	return &server{db: db}
}

// CreateMeeting - метод для создания встречи и добавления в базу данных
func (s *server) CreateMeeting(ctx context.Context, req *NewMeetingRequest) (*NewMeetingResponse, error) {
	if req.NewMeeting == nil {
		log.Println("NewMeeting is nil")
		return nil, fmt.Errorf("NewMeeting is nil")
	}
	log.Printf("Создание встречи: %v", req)

	// Проверка доступности комнаты с помощью метода CheckAvailability
	availabilityReq := &AvailabilityRequest{
		RoomId:    req.NewMeeting.Id,
		StartTime: req.NewMeeting.TimeStart,
		EndTime:   req.NewMeeting.TimeEnd,
	}
	availabilityResp, err := s.CheckAvailability(ctx, availabilityReq)
	if err != nil {
		log.Printf("Ошибка при проверке доступности: %v", err)
		return &NewMeetingResponse{
			MeetingId:    0,
			ErrorMessage: "Ошибка при проверке доступности",
		}, err
	}

	if !availabilityResp.Available {
		// Если комната занята в указанный интервал, возвращаем сообщение об ошибке
		return &NewMeetingResponse{
			MeetingId:    0,
			ErrorMessage: "Комната занята в указанный временной интервал",
		}, nil
	}
	_, err = s.db.Exec(
		"INSERT INTO meetings (room_id, topic, start_time, end_time) VALUES ($1, $2, $3, $4)",
		req.NewMeeting.Id, req.NewMeeting.Topic, time.Unix(req.NewMeeting.TimeStart, 0), time.Unix(req.NewMeeting.TimeEnd, 0),
	)
	if err != nil {
		log.Printf("Ошибка при создании встречи: %v", err)
		return &NewMeetingResponse{
			MeetingId:    0, // Или другое значение по умолчанию
			ErrorMessage: "Ошибка при создании встречи",
		}, err
	}

	// Получаем ID созданной встречи, если это требуется.
	var meetingID int64
	err = s.db.QueryRow("SELECT lastval()").Scan(&meetingID)
	if err != nil {
		log.Printf("Ошибка при получении ID встречи: %v", err)
		return &NewMeetingResponse{
			MeetingId:    0,
			ErrorMessage: "Ошибка при получении ID встречи",
		}, err
	}

	return &NewMeetingResponse{
		MeetingId:    meetingID,
		ErrorMessage: "",
	}, nil
}

// CheckAvailability - метод для проверки доступности комнаты
func (s *server) CheckAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error) {
	log.Printf("Проверка доступности комнаты %d на интервал: %v - %v", req.RoomId, req.StartTime, req.EndTime)

	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM meetings
		WHERE room_id = $1 AND (
			(start_time, end_time) OVERLAPS ($2::timestamp, $3::timestamp)
		)
	`, req.RoomId, time.Unix(req.StartTime, 0), time.Unix(req.EndTime, 0)).Scan(&count)
	if err != nil {
		log.Printf("Ошибка при проверке доступности: %v", err)
		return nil, err
	}

	available := count == 0
	return &AvailabilityResponse{Available: available}, nil
}

// GetMeetingsInRoom - метод для получения списка встреч в конкретной комнате
func (s *server) GetMeetingsInRoom(ctx context.Context, req *MeetingsInTheRoomRequest) (*MeetingsInTheRoomResponse, error) {
	log.Printf("Получение встреч для комнаты %d", req.RoomId)

	rows, err := s.db.Query("SELECT id, topic, start_time, end_time FROM meetings WHERE room_id = $1", req.RoomId)
	if err != nil {
		log.Printf("Ошибка при получении встреч: %v", err)
		return nil, err
	}
	defer rows.Close()

	var meetings []*Meeting
	for rows.Next() {
		var meeting Meeting
		var startTime, endTime time.Time
		if err := rows.Scan(&meeting.Id, &meeting.Topic, &startTime, &endTime); err != nil {
			log.Printf("Ошибка при сканировании встречи: %v", err)
			return nil, err
		}
		meeting.TimeStart = startTime.Unix()
		meeting.TimeEnd = endTime.Unix()
		meetings = append(meetings, &meeting)
	}

	return &MeetingsInTheRoomResponse{Meetings: meetings}, nil
}
