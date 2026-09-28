package services

import (
	"context"
	"fmt"
	"gorm.io/gorm/clause"
	"kori/internal/events"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BaseService interface defines common CRUD operations
type BaseService[T any] interface {
	Create(ctx context.Context, entity *T, includes ...string) error
	Get(ctx context.Context, id string, includes ...string) (*T, error)
	List(ctx context.Context, page, limit int, filters map[string]interface{}, excludeFields map[string]bool, sortFields []string, order string, includes ...string) ([]T, int64, error)
	Update(ctx context.Context, id string, entity *T, includes ...string) error
	Delete(ctx context.Context, id string) error
}

// BaseServiceImpl implements BaseService
type BaseServiceImpl[T any] struct {
	db        *gorm.DB
	modelType T
}

func GormTableName(db *gorm.DB, v any) string {
	struct_name := reflect.TypeOf(v).Name()
	return db.NamingStrategy.TableName(struct_name)
}

// NewBaseService creates a new base service
func NewBaseService[T any](db *gorm.DB, modelType T) BaseService[T] {
	return &BaseServiceImpl[T]{
		db:        db,
		modelType: modelType,
	}
}

// applyIncludes adds preload statements to the query for each include
func (s *BaseServiceImpl[T]) applyIncludes(query *gorm.DB, includes ...string) *gorm.DB {
	for _, include := range includes {
		query = query.Preload(include)
		// Handle nested includes with field selection
		//parts := strings.Split(include, ".")
		//if len(parts) > 1 {
		//	log.Info(
		//		"parts[0]: %s, parts[1:]: %s", parts[0], parts[1:])
		//	// For nested preloads like "HtmlFile.name", use closure to specify fields
		//	query = query.Preload(parts[0], func(db *gorm.DB) *gorm.DB {
		//		return db.Select(parts[1:])
		//	})
		//} else {
		//	// Regular preload for single relationships
		//	query = query.Preload(include)
		//}
	}
	return query
}

func (s *BaseServiceImpl[T]) applyExcludes(query *gorm.DB, excludes map[string]bool) *gorm.DB {
	for field := range excludes {
		query = query.Omit(field)
	}
	return query
}

func (s *BaseServiceImpl[T]) Create(ctx context.Context, entity *T, includes ...string) error {
	if err := s.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}

	// Reload the entity with includes if any are specified
	if len(includes) > 0 {
		if err := s.applyIncludes(s.db.WithContext(ctx), includes...).First(entity, "id = ?", reflect.ValueOf(*entity).FieldByName("ID").String()).Error; err != nil {
			return err
		}
	}

	// Get the table name of the gorm model
	events.Emit(fmt.Sprintf("%s.created", GormTableName(s.db, s.modelType)), entity)

	return nil
}

func (s *BaseServiceImpl[T]) Get(ctx context.Context, id string, includes ...string) (*T, error) {
	var entity T
	query := s.db.WithContext(ctx)
	query = s.applyIncludes(query, includes...)

	// filter deleted entities
	query = query.Where("is_deleted = ?", false)

	if err := query.First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (s *BaseServiceImpl[T]) List(ctx context.Context, page, limit int, filters map[string]interface{}, excludes map[string]bool, sortFields []string, order string, includes ...string) ([]T, int64, error) {
	var entities []T
	var total int64

	query := s.db.WithContext(ctx).Model(s.modelType)

	// Apply filters
	for key, value := range filters {
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(key) {
			return nil, 0, fmt.Errorf("invalid filter column")
		}
		query = query.Where(clause.Eq{Column: clause.Column{Name: key}, Value: value})
	}

	// Apply includes
	query = s.applyIncludes(query, includes...)

	// Apply excludes
	query = s.applyExcludes(query, excludes)

	// Apply sort
	if len(sortFields) > 0 {
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(sortFields[0]) {
			return nil, 0, fmt.Errorf("invalid sort column")
		}
		if order != "" && !strings.EqualFold(order, "ASC") && !strings.EqualFold(order, "DESC") {
			return nil, 0, fmt.Errorf("invalid sort order")
		}
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortFields[0]}, Desc: strings.EqualFold(order, "DESC")})
	}

	// filter deleted entities
	query = query.Where("is_deleted = ?", false)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (s *BaseServiceImpl[T]) Update(ctx context.Context, id string, entity *T, includes ...string) error {
	if err := s.db.WithContext(ctx).Model(entity).Where("id = ? AND is_deleted = ?", id, false).Omit("id").Omit("teamId").Updates(entity).Error; err != nil {
		return err
	}

	// Reload the entity with includes if any are specified
	if len(includes) > 0 {
		if err := s.applyIncludes(s.db.WithContext(ctx), includes...).First(entity, "id = ?", id).Error; err != nil {
			return err
		}
	}

	events.Emit(fmt.Sprintf("%s.updated", GormTableName(s.db, s.modelType)), entity)

	return nil
}

func (s *BaseServiceImpl[T]) Delete(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Model(s.modelType).Where("id = ? AND is_deleted = ?", id, false).Update("deleted_at", time.Now()).Update("is_deleted", true).Error; err != nil {
		return err
	}

	events.Emit(fmt.Sprintf("%s.deleted", GormTableName(s.db, s.modelType)), id)

	return nil
}
