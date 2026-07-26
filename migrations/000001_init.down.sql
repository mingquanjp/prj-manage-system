-- Rollback schema v2. Thứ tự ngược với bản .up để không vướng khóa ngoại.

DROP TABLE IF EXISTS cv_extracted_skills;
DROP TABLE IF EXISTS cv_extractions;
DROP TABLE IF EXISTS employee_cvs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS task_status_history;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS work_logs;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS recommendation_items;
DROP TABLE IF EXISTS recommendation_runs;
DROP TABLE IF EXISTS task_requirements;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS task_groups;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS employee_skills;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS skill_categories;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS skill_source;
DROP TYPE IF EXISTS change_source;
DROP TYPE IF EXISTS decision_status;
DROP TYPE IF EXISTS extraction_status;
DROP TYPE IF EXISTS cv_status;
DROP TYPE IF EXISTS assignment_type;
DROP TYPE IF EXISTS join_source;
DROP TYPE IF EXISTS proficiency_level;
DROP TYPE IF EXISTS skill_status;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS project_status;
DROP TYPE IF EXISTS priority_level;
DROP TYPE IF EXISTS employment_status;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_system_role;
