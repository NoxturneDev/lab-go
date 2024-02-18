package main

import (
	"github.com/gen2brain/beeep"
	//"github.com/gen2brain/beeep"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

// var notify *notificator.Notificator
// TODO : Add database to add streak, add cron job for every 4 hours

type Streak struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey"`
	CurrentStreak uint      `json:"current-streak"`
	HighestStreak uint      `json:"highest-streak"`
	LastStreak    time.Time `json:"last-streak"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

var database *gorm.DB

func CreateStreak(c echo.Context) error {
	db := DB()

	s := new(Streak)
	if err := c.Bind(s); err != nil {
		return err
	}

	if err := db.Create(&s).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.JSON(http.StatusCreated, s)
}

func UpdateStreak(c echo.Context) error {
	db := DB()
	id := c.Param("id")
	println(id)

	s := new(Streak)

	if err:= c.Bind(s); err != nil {
		log.Fatalf("failed to bind")
	}

	existingStreak := new(Streak)

	if err := db.First(&existingStreak, id).Error; err != nil {
		log.Fatalf("failed get existing streak")
	}

	existingStreak.CurrentStreak++
	existingStreak.LastStreak = s.LastStreak

	if existingStreak.CurrentStreak > existingStreak.HighestStreak {
		existingStreak.HighestStreak = existingStreak.CurrentStreak
	}

	if err := db.Save(&existingStreak).Error; err != nil {
		log.Fatalf("Failed to update the data")
	}

	return c.JSON(http.StatusCreated, existingStreak)
}

func getStreak(c echo.Context) {
	db := DB()

	s := new(Streak)
	if err := c.Bind(s); err != nil {
		log.Fatalf("cant bind")
	}

	data := db.Find(&s)

	if data.Error != nil {
		log.Fatalf("can't get any data")
	}

}

func dbInit() {
	dsn := "root@tcp(localhost:3306)/streak_go_db?parseTime=true&timeout=300ms&charset=utf8mb4&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect Database: %v", err)
		//panic(err)
	}

	database = db

	migrateErr := db.AutoMigrate(&Streak{})
	if migrateErr != nil {
		//panic(err)
		log.Fatalf("Failed to migrate Database: %v", migrateErr)
	}
}

func DB() *gorm.DB {
	return database
}

func main() {
	e := echo.New()

	dbInit()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/streak", CreateStreak)
	e.PUT("/streak/:id", UpdateStreak)
	e.GET("/streak", getStreak)

	err := beeep.Notify("Friendly Reminder", "DO YOUR FUCKING LEARNING CODING DUMBASS!", "")
	//
	if err != nil {
		panic(err)
	}

	e.Logger.Fatal(e.Start(":8080"))
	//notify = notificator.New(notificator.Options{
	//	DefaultIcon: "icon/default.png",
	//	AppName:     "My test App",
	//})
	//
	//notify.Push("title", "text", "", notificator.UR_CRITICAL)
}
