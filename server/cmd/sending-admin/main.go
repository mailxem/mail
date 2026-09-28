// sending-admin is an operator-only CLI. It never enables an AWS account, sends
// mail, or exposes approval controls to workspace administrators.
package main

import (
	"flag"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/sending"
	"log"
	"os"
)

func main() {
	team := flag.String("team", "", "Workspace UUID")
	action := flag.String("action", "", "approve or suspend")
	daily := flag.Int64("daily", 100, "Daily recipient limit")
	monthly := flag.Int64("monthly", 3000, "Monthly recipient limit")
	budget := flag.Int64("budget-micros", 3000000, "Monthly reserved delivery allowance in micro-USD")
	flag.Parse()
	if uuid.Validate(*team) != nil || (*action != "approve" && *action != "suspend") || *daily <= 0 || *monthly <= 0 || *budget <= 0 || *daily > 1000000000 || *monthly > 1000000000 || *budget > 1000000000 {
		log.Fatal("Use -team UUID -action approve|suspend and positive bounded limits")
	}
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Connect(cfg); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	conn := db.GetDB()
	if err = sending.Migrate(conn); err != nil {
		log.Fatal(err)
	}
	err = conn.Transaction(func(tx *gorm.DB) error {
		var a sending.Account
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", *team).Error; e != nil {
			return e
		}
		fields := map[string]any{"suspended": true}
		if *action == "approve" {
			fields = map[string]any{"approved": true, "suspended": false, "daily_limit": *daily, "monthly_limit": *monthly, "monthly_budget_micros": *budget}
		}
		if e := tx.Model(&a).Updates(fields).Error; e != nil {
			return e
		}
		return tx.Create(&sending.Audit{ID: uuid.NewString(), TeamID: *team, Actor: "operator:" + os.Getenv("USER"), Action: *action}).Error
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Managed sending %s applied to workspace %s; current usage preserved", *action, *team)
}
