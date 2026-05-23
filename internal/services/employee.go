package services

import (
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tz-app/internal/logger"
	"tz-app/internal/models"
)

type EmployeeService struct {
	db *gorm.DB
}

func NewEmployeeService(db *gorm.DB) *EmployeeService {
	return &EmployeeService{db: db}
}

func (s *EmployeeService) Create(departmentID uint, fullName, position string, hiredAt *time.Time) (*models.Employee, error) {
	fullName = strings.TrimSpace(fullName)
	position = strings.TrimSpace(position)

	if len(fullName) == 0 || len(fullName) > 200 {
		logger.Warn("Invalid full name", zap.String("full_name", fullName))
		return nil, errors.New("full_name must be between 1 and 200 characters")
	}

	if len(position) == 0 || len(position) > 200 {
		logger.Warn("Invalid position", zap.String("position", position))
		return nil, errors.New("position must be between 1 and 200 characters")
	}

	var dept models.Department
	if err := s.db.First(&dept, departmentID).Error; err != nil {
		logger.Warn("Department not found for employee creation", zap.Uint("department_id", departmentID))
		return nil, errors.New("department not found")
	}

	emp := &models.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := s.db.Create(emp).Error; err != nil {
		logger.Error("Failed to create employee", zap.Error(err))
		return nil, err
	}

	logger.Info("Employee created",
		zap.Uint("id", emp.ID),
		zap.Uint("department_id", departmentID),
		zap.String("full_name", fullName),
		zap.String("position", position),
	)

	return emp, nil
}
