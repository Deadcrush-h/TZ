package services

import (
	"errors"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tz-app/internal/logger"
	"tz-app/internal/models"
)

type DepartmentService struct {
	db *gorm.DB
}

func NewDepartmentService(db *gorm.DB) *DepartmentService {
	return &DepartmentService{db: db}
}

func (s *DepartmentService) Create(name string, parentID *uint) (*models.Department, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 || len(name) > 200 {
		logger.Warn("Invalid department name",
			zap.String("name", name),
			zap.Int("length", len(name)),
		)
		return nil, errors.New("name must be between 1 and 200 characters")
	}

	if parentID != nil && *parentID != 0 {
		var parent models.Department
		if err := s.db.First(&parent, *parentID).Error; err != nil {
			logger.Warn("Parent department not found", zap.Uint("parent_id", *parentID))
			return nil, errors.New("parent department not found")
		}
	}

	var count int64
	query := s.db.Model(&models.Department{}).Where("name = ?", name)
	if parentID != nil && *parentID != 0 {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	query.Count(&count)
	if count > 0 {
		logger.Warn("Duplicate department name",
			zap.String("name", name),
			zap.Any("parent_id", parentID),
		)
		return nil, errors.New("department with this name already exists at this level")
	}

	dept := &models.Department{
		Name:     name,
		ParentID: parentID,
	}

	if err := s.db.Create(dept).Error; err != nil {
		logger.Error("Failed to create department", zap.Error(err))
		return nil, err
	}

	logger.Info("Department created",
		zap.Uint("id", dept.ID),
		zap.String("name", dept.Name),
		zap.Any("parent_id", dept.ParentID),
	)

	return dept, nil
}

func (s *DepartmentService) GetByID(id uint, depth int, includeEmployees bool) (map[string]interface{}, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	var dept models.Department
	if err := s.db.First(&dept, id).Error; err != nil {
		logger.Warn("Department not found", zap.Uint("id", id))
		return nil, errors.New("department not found")
	}

	result := map[string]interface{}{
		"id":         dept.ID,
		"name":       dept.Name,
		"parent_id":  dept.ParentID,
		"created_at": dept.CreatedAt,
	}

	if includeEmployees {
		var employees []models.Employee
		s.db.Where("department_id = ?", id).Order("created_at ASC").Find(&employees)
		result["employees"] = employees
		logger.Debug("Employees loaded for department",
			zap.Uint("department_id", id),
			zap.Int("count", len(employees)),
		)
	}

	if depth > 1 {
		var children []models.Department
		s.db.Where("parent_id = ?", id).Find(&children)

		childResults := []map[string]interface{}{}
		for _, child := range children {
			childResult, _ := s.getChildTree(child.ID, depth-1, includeEmployees)
			childResults = append(childResults, childResult)
		}
		result["children"] = childResults
	}

	return result, nil
}

func (s *DepartmentService) getChildTree(id uint, depth int, includeEmployees bool) (map[string]interface{}, error) {
	var dept models.Department
	if err := s.db.First(&dept, id).Error; err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":         dept.ID,
		"name":       dept.Name,
		"parent_id":  dept.ParentID,
		"created_at": dept.CreatedAt,
	}

	if includeEmployees {
		var employees []models.Employee
		s.db.Where("department_id = ?", id).Order("created_at ASC").Find(&employees)
		result["employees"] = employees
	}

	if depth > 1 {
		var children []models.Department
		s.db.Where("parent_id = ?", id).Find(&children)

		childResults := []map[string]interface{}{}
		for _, child := range children {
			childResult, _ := s.getChildTree(child.ID, depth-1, includeEmployees)
			childResults = append(childResults, childResult)
		}
		result["children"] = childResults
	}

	return result, nil
}

func (s *DepartmentService) Update(id uint, name *string, parentID *uint) (*models.Department, error) {
	var dept models.Department
	if err := s.db.First(&dept, id).Error; err != nil {
		logger.Warn("Department not found for update", zap.Uint("id", id))
		return nil, errors.New("department not found")
	}

	oldName := dept.Name
	oldParentID := dept.ParentID

	if name != nil {
		newName := strings.TrimSpace(*name)
		if len(newName) == 0 || len(newName) > 200 {
			return nil, errors.New("name must be between 1 and 200 characters")
		}

		var count int64
		query := s.db.Model(&models.Department{}).Where("name = ? AND id != ?", newName, id)
		if dept.ParentID != nil {
			query = query.Where("parent_id = ?", dept.ParentID)
		} else {
			query = query.Where("parent_id IS NULL")
		}

		query.Count(&count)
		if count > 0 {
			return nil, errors.New("department with this name already exists at this level")
		}

		dept.Name = newName
	}

	if parentID != nil {
		if *parentID == id {
			return nil, errors.New("department cannot be parent of itself")
		}

		if *parentID != 0 {
			var parent models.Department
			if err := s.db.First(&parent, *parentID).Error; err != nil {
				return nil, errors.New("parent department not found")
			}

			if s.wouldCreateCycle(id, *parentID) {
				logger.Warn("Cycle detected in hierarchy",
					zap.Uint("id", id),
					zap.Uint("new_parent_id", *parentID),
				)
				return nil, errors.New("this move would create a cycle")
			}
		}

		if *parentID == 0 {
			dept.ParentID = nil
		} else {
			dept.ParentID = parentID
		}
	}

	if err := s.db.Save(&dept).Error; err != nil {
		logger.Error("Failed to update department", zap.Error(err))
		return nil, err
	}

	logger.Info("Department updated",
		zap.Uint("id", id),
		zap.String("old_name", oldName),
		zap.String("new_name", dept.Name),
		zap.Any("old_parent_id", oldParentID),
		zap.Any("new_parent_id", dept.ParentID),
	)

	return &dept, nil
}

func (s *DepartmentService) wouldCreateCycle(deptID, newParentID uint) bool {
	current := newParentID
	for current != 0 {
		if current == deptID {
			return true
		}
		var parent models.Department
		if err := s.db.Select("parent_id").First(&parent, current).Error; err != nil {
			return false
		}
		if parent.ParentID == nil {
			break
		}
		current = *parent.ParentID
	}
	return false
}

func (s *DepartmentService) Delete(id uint, mode string, reassignToID *uint) error {
	logger.Info("Deleting department",
		zap.Uint("id", id),
		zap.String("mode", mode),
	)

	return s.db.Transaction(func(tx *gorm.DB) error {
		if mode == "cascade" {
			if err := tx.Delete(&models.Department{}, id).Error; err != nil {
				logger.Error("Failed to cascade delete department", zap.Error(err))
				return err
			}
			logger.Info("Department cascade deleted", zap.Uint("id", id))
			return nil
		} else if mode == "reassign" {
			if reassignToID == nil {
				return errors.New("reassign_to_department_id is required")
			}

			var targetDept models.Department
			if err := tx.First(&targetDept, *reassignToID).Error; err != nil {
				logger.Warn("Target department not found", zap.Uint("target_id", *reassignToID))
				return errors.New("target department not found")
			}

			// Переназначаем сотрудников
			if err := tx.Model(&models.Employee{}).Where("department_id = ?", id).Update("department_id", *reassignToID).Error; err != nil {
				return err
			}

			// Переназначаем дочерние подразделения
			if err := tx.Model(&models.Department{}).Where("parent_id = ?", id).Update("parent_id", *reassignToID).Error; err != nil {
				return err
			}

			if err := tx.Delete(&models.Department{}, id).Error; err != nil {
				return err
			}

			logger.Info("Department deleted with reassign",
				zap.Uint("id", id),
				zap.Uint("reassign_to", *reassignToID),
			)
			return nil
		}
		return errors.New("invalid mode. Use 'cascade' or 'reassign'")
	})
}
