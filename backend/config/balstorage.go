package config

import "scrapingmanga/backend/utils"

type BalStorageConfig struct {
	BaseURL        string
	Email          string
	Password       string
	RootFolderName string
}

func LoadBalStorageConfig() BalStorageConfig {
	return BalStorageConfig{
		BaseURL:        utils.GetEnv("BALSTORAGE_BASE_URL", "http://localhost:8000/api/v1"),
		Email:          utils.GetEnv("BALSTORAGE_EMAIL", ""),
		Password:       utils.GetEnv("BALSTORAGE_PASSWORD", ""),
		RootFolderName: utils.GetEnv("BALSTORAGE_ROOT_FOLDER_NAME", "Manga"),
	}
}
