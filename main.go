package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Amount    int    `json:"Amount"`
	OrderNo   string `json:"order_no"`
	ProductID int    `json:"product_id"`
	QrType    string `json:"qr_type"`
	Cmd       string `json:"cmd"`
	VmcNo     int    `json:"vmc_no"`
}

type Response struct {
	Cmd     string `json:"cmd"`
	VmcNo   int    `json:"vmc_no"`
	QrType  string `json:"qr_type"`
	Qrcode  string `json:"qrcode"`
	OrderNo string `json:"order_no"`
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
	if err := json.Unmarshal([]byte(jsonString), &request); err != nil {
		fmt.Printf("Ошибка при разборе JSON: %v\n", err)
		return
	}

	response := Response{
		Cmd:     request.Cmd + "_r",
		VmcNo:   request.VmcNo,
		QrType:  request.QrType,
		Qrcode:  "XXXXXXXXXXXXXX", // Здесь должен быть ваш QR код
		OrderNo: request.OrderNo,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Ошибка при формировании ответа для %s: %v\n", clientAddr, err)
		return
	}

	_, err = conn.Write(responseData)
	if err != nil {
		fmt.Printf("Ошибка при отправке ответа клиенту %s: %v\n", clientAddr, err)
	}
}
