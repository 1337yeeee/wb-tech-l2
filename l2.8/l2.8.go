package main

import (
	"fmt"
	"github.com/beevik/ntp"
	"os"
	"time"
)

// TimeProvider определяет интерфейс для получения времени от указанного сервера.
type TimeProvider interface {
	Now(server string) (time.Time, error)
}

// NTPProvider реализует TimeProvider и получает время через NTP.
type NTPProvider struct{}

// Now возвращает текущее время с указанного NTP-сервера.
// В случае ошибки возвращает нулевое значение time.Time и ошибку.
func (NTPProvider) Now(server string) (time.Time, error) {
	return ntp.Time(server)
}

func main() {
	const ntpServer = "1.ru.pool.ntp.org"

	var ntpTimeProvider TimeProvider = NTPProvider{}

	ntpTime, err := ntpTimeProvider.Now(ntpServer)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "NTP Error:", err)
		os.Exit(1)
	}

	fmt.Println(ntpTime.Format(time.RFC1123Z))
}

// Вывод программы:
//Fri, 15 Aug 2025 12:20:02 +0000

// Примеры ошибок ntp:

// Сервер не ответил в пределах таймаута
// NTP Error:  read udp ip:port->ntp-ip:123: i/o timeout
// exit status 1

// Не удалось разрешить DNS-имя NTP-сервера
// NTP Error:  lookup 10.ru.pool.ntp.org: no such host
// exit status 1
