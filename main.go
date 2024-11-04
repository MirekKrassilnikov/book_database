package main

import (
	"log"
	"net"

	"github.com/MirekKrassilnikov/book_the_room/database"
	meetingservice "github.com/MirekKrassilnikov/book_the_room/meetingService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Создаем и подключаемся к базе данных
	db := database.CreateDatabase()

	// Создание нового gRPC-сервера и регистрация сервиса
	grpcServer := grpc.NewServer()
	s := meetingservice.NewServer(db)
	meetingservice.RegisterMeetingServiceServer(grpcServer, s)

	// Регистрация рефлексии
	reflection.Register(grpcServer)

	// Настройка прослушивания на порту
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка при прослушивании: %v", err)
	}

	log.Println("Сервер запущен на порту :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
