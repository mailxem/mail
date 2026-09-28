package controllers

import (
	"github.com/labstack/echo/v4"
	"kori/internal/models"
	"net/http"
	"reflect"
)

// authorizeEntity runs before relationship preloads, updates, or deletes.
func (c *BaseController[T]) authorizeEntity(ctx echo.Context, id string) error {
	entity, err := c.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "entity not found")
	}
	v := reflect.ValueOf(entity).Elem()
	if f := v.FieldByName("TeamID"); f.IsValid() && f.Kind() == reflect.String {
		if f.String() != ctx.Get("teamID") {
			return echo.NewHTTPError(http.StatusNotFound, "entity not found")
		}
	}
	if f := v.FieldByName("UserID"); f.IsValid() && f.Kind() == reflect.String {
		if f.String() != ctx.Get("userID") {
			return echo.NewHTTPError(http.StatusNotFound, "entity not found")
		}
	}
	if v.Type() == reflect.TypeOf(models.Team{}) && v.FieldByName("ID").String() != ctx.Get("teamID") {
		return echo.NewHTTPError(http.StatusNotFound, "entity not found")
	}
	if sender, ok := any(entity).(*models.SMTPConfig); ok && sender.Provider == "MANAGED" && ctx.Request().Method != "GET" {
		return echo.NewHTTPError(403, "Manage this sender from Sending settings")
	}
	return nil
}
func enforceOwner[T any](ctx echo.Context, entity *T, create bool) {
	v := reflect.ValueOf(entity).Elem()
	if v.Type() == reflect.TypeOf(models.Domain{}) {
		v.FieldByName("IsVerified").SetBool(false)
	}
	for field, key := range map[string]string{"TeamID": "teamID", "UserID": "userID"} {
		if f := v.FieldByName(field); f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			if value, ok := ctx.Get(key).(string); ok {
				f.SetString(value)
			}
		}
	}
	if create {
		if f := v.FieldByName("ID"); f.IsValid() && f.CanSet() {
			f.SetString("")
		}
	}
	if f := v.FieldByName("IsDeleted"); f.IsValid() && f.CanSet() {
		f.SetBool(false)
	}
	// Relationships are read through includes, but never mass-assigned from a body.
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Struct && f.Type().Elem().PkgPath() == "kori/internal/models" && f.CanSet() {
			f.SetZero()
		}
	}
}
