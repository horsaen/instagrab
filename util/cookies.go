package util

import (
	"bufio"
	"os"
)

func LoadCookies() [2]string {
	home, _ := os.UserHomeDir()

	configBase := home + "/.instagrab"

	cookieDir := configBase + "/cookies"

	file, _ := os.Open(cookieDir)

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	cookies := [2]string{" ", " "}

	index := 0
	for scanner.Scan() {
		cookies[index] = scanner.Text()
		index++
	}

	file.Close()

	return cookies
}
