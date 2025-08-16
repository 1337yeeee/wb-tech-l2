package main

import (
	"fmt"
	"github.com/beevik/ntp"
	"os"
	"time"
)

func main() {
	const NtpServer = "1.ru.pool.ntp.org"

	ntpTime, err := ntp.Query(NtpServer)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "NTP Error: ", err)
		os.Exit(1)
	}

	fmt.Println(ntpTime.Time.Format(time.RFC1123Z))
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
