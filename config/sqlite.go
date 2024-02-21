package config

import (
	"os"

	"github.com/KaueChristian/Goportunitties/schemas"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitializeSQlite() (*gorm.DB, error) {
    logger := GetLogger("Sqlite")
    dbPath := "./db/main.db"

    // Chech if DB file exists
    _, err := os.Stat(dbPath)
    if os.IsNotExist(err){
        logger.Info("DataBase file not found. Creating...")
        //Create DataBase file and directory
        err = os.MkdirAll("./db", os.ModePerm)
        if err != nil {
            return nil, err
        }
        file, err := os.Create(dbPath)
        if err != nil{
            return nil,err
        }
        file.Close()
    }

    // Create DB and Connect
    db, err :=gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        logger.Errorf("Sqlite Opening Erro: %v", err)
        return nil, err
    }

    err = db.AutoMigrate(&schemas.Opening{})
    if err != nil{
        logger.Errorf("sqlite automigration erro:%v", err)
        return nil, err
    }

    return db, nil
}