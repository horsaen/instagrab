package util

import "os"

func Exists(dir string) {
	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

func InitConfDir() {
	home, _ := os.UserHomeDir()

	configBase := home + "/.instagrab"

	_, err := os.Stat(configBase)

	if os.IsNotExist(err) {
		os.MkdirAll(configBase, os.ModePerm)
		os.Create(configBase + "/cookies")
	}
}
