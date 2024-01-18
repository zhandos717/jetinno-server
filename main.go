package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type PaymentRequest struct {
	Amount    int    `json:"Amount"`
	OrderNo   string `json:"order_no"`
	ProductID int    `json:"product_id"`
	QrType    string `json:"qr_type"`
	Cmd       string `json:"cmd"`
	VmcNo     int    `json:"vmc_no"`
}

type Request struct {
	Cmd string `json:"cmd"`
}

type Response struct {
	Cmd     string `json:"cmd"`
	VmcNo   int    `json:"vmc_no"`
	QrType  string `json:"qr_type"`
	Qrcode  string `json:"qrcode"`
	OrderNo string `json:"order_no"`
}

type LoginRequest struct {
	CompID     int    `json:"comp_id"`
	LoginCount int    `json:"login_count"`
	Sign       string `json:"sign"`
	Timestamp  string `json:"timestamp"`
	Version    string `json:"version"`
	Cmd        string `json:"cmd"`
	VmcNo      int    `json:"vmc_no"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginResponse struct {
	Cmd         string `json:"cmd"`
	VmcNo       int    `json:"vmc_no"`
	CarrierCode string `json:"carrier_code"`
	DateTime    string `json:"date_time"`
	ServerList  string `json:"server_list"`
	Ret         int    `json:"ret"`
}

func main() {
	const PORT = ":4040"

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("Ошибка при создании сервера: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Сервер запущен и слушает порт %s\n", PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Ошибка при подключении: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Клиент подключен: %s\n", clientAddr)

	buffer := make([]byte, 1024)
	length, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Ошибка при чтении данных от %s: %v\n", clientAddr, err)
		return
	}

	fmt.Printf("Сырые данные запроса от %s: %s\n", clientAddr, buffer[:length])

	rawData := buffer[:length]
	startIndex := strings.Index(string(rawData), "{")
	if startIndex == -1 {
		fmt.Println("Ошибка: данные не содержат JSON объекта")
		return
	}
	jsonString := string(rawData[startIndex:])

	request := Request{}

	err = json.Unmarshal([]byte(jsonString), &request)
	if err != nil {
		sendErrorResponse(conn, "Ошибка при разборе запроса")
		return
	}

	fmt.Printf("JSon от запроса от %s \n", jsonString)

	fmt.Printf("Тип запроса %s \n", request.Cmd)

	switch request.Cmd {
	case "login":
		handleLogin(conn, []byte(jsonString))
	case "qrcode":
		handlePayment(conn, []byte(jsonString))
	default:
		fmt.Println("Неизвестная команда")
	}
}

func handleLogin(conn net.Conn, message []byte) {
	var loginReq LoginRequest
	err := json.Unmarshal(message, &loginReq)
	if err != nil {
		fmt.Println("Ошибка при разборе запроса логина:", err)
		sendErrorResponse(conn, "Ошибка при разборе запроса")
		return
	}

	response := LoginResponse{
		Cmd:         loginReq.Cmd + "_r",
		VmcNo:       loginReq.VmcNo, // Или другое значение, если требуется
		CarrierCode: "TW-00418",     // Пример значения
		DateTime:    time.Now().Format("2006-01-02 15:04:05"),
		ServerList:  "185.100.67.252",
		Ret:         0,
	}

	respJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Ошибка при формировании ответа JSON:", err)
		return
	}

	fmt.Printf("Отправлен ответ %s \n", respJSON)

	_, err = conn.Write(respJSON)
	if err != nil {
		fmt.Println("Ошибка при отправке ответа:", err)
		return
	}
}

func handlePayment(conn net.Conn, message []byte) {
	// Распарсивание запроса
	var request PaymentRequest
	err := json.Unmarshal(message, &request)
	if err != nil {
		fmt.Printf("Ошибка при разборе запроса на оплату: %v\n", err)
		sendErrorResponse(conn, "Ошибка при разборе запроса")
		return
	}

	response := Response{
		Cmd:     request.Cmd + "_r",
		VmcNo:   request.VmcNo,
		QrType:  request.QrType,
		Qrcode:  "XXXXXXXXXXXXXX", // Здесь должен быть сгенерированный QR код
		OrderNo: request.OrderNo,
	}

	respJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Ошибка при формировании ответа JSON:", err)
		return
	}

	fmt.Printf("Отправлен ответ %s \n", respJSON)

	_, err = conn.Write(respJSON)
	if err != nil {
		fmt.Println("Ошибка при отправке ответа:", err)
		return
	}
}

func sendErrorResponse(conn net.Conn, errMsg string) {
	// Создаем объект ErrorResponse с сообщением об ошибке
	errorResponse := ErrorResponse{Error: errMsg}

	// Преобразуем его в JSON
	jsonResponse, err := json.Marshal(errorResponse)
	if err != nil {
		fmt.Printf("Ошибка при формировании JSON ответа об ошибке: %v\n", err)
		return
	}

	// Отправляем JSON клиенту
	_, err = conn.Write(jsonResponse)
	if err != nil {
		fmt.Printf("Ошибка при отправке ответа об ошибке: %v\n", err)
	}
}
