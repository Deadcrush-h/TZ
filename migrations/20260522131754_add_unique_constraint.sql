-- +goose Up
CREATE UNIQUE INDEX idx_unique_dept_name_per_parent 
ON departments(name, COALESCE(parent_id, 0));

-- +goose Down
DROP INDEX IF EXISTS idx_unique_dept_name_per_parent;