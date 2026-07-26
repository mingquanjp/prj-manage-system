-- =============================================================================
-- Schema v2 — Hệ thống quản lý dự án và phân công nhân sự hỗ trợ AI
--
-- Bám theo mục 4.2.3 (Hình 4.6 - Mô hình quan hệ dữ liệu) của tài liệu thiết kế.
-- Quy ước đặt tên: tài liệu dùng PascalCase (EmployeeSkills), CSDL dùng snake_case
-- (employee_skills) theo chuẩn PostgreSQL. Ánh xạ 1-1, không đổi ngữ nghĩa.
--
-- Kiểu `datetime` trong tài liệu được hiện thực bằng `timestamptz` để lưu đúng
-- múi giờ — cần cho việc tính thời gian thực hiện theo lịch làm việc ở UC004.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. ENUM (15)
-- -----------------------------------------------------------------------------

-- Vai trò khớp đúng 3 tác nhân của Chương 2: Quản trị viên / Quản lý dự án / Nhân viên
CREATE TYPE user_system_role  AS ENUM ('ADMIN', 'PM', 'EMPLOYEE');
CREATE TYPE user_status       AS ENUM ('ACTIVE', 'LOCKED', 'DISABLED');
CREATE TYPE employment_status AS ENUM ('ACTIVE', 'INACTIVE', 'RESIGNED');
CREATE TYPE priority_level    AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'URGENT');
CREATE TYPE project_status    AS ENUM ('PLANNING', 'IN_PROGRESS', 'COMPLETED', 'CLOSED', 'CANCELLED');

-- PENDING_REVIEW bổ sung so với tài liệu: UC004 bước 5 yêu cầu trạng thái "Chờ review"
CREATE TYPE task_status       AS ENUM ('PENDING_ASSIGNMENT', 'ASSIGNED', 'IN_PROGRESS', 'PENDING_REVIEW', 'DONE', 'CLOSED');

CREATE TYPE skill_status      AS ENUM ('ACTIVE', 'INACTIVE');
CREATE TYPE proficiency_level AS ENUM ('BASIC', 'INTERMEDIATE', 'ADVANCED', 'EXPERT');
CREATE TYPE join_source       AS ENUM ('MANUAL', 'FROM_ASSIGNMENT');
CREATE TYPE assignment_type   AS ENUM ('AI', 'MANUAL');
CREATE TYPE cv_status         AS ENUM ('ACTIVE', 'ARCHIVED', 'REPLACED');
CREATE TYPE extraction_status AS ENUM ('PENDING_REVIEW', 'APPROVED', 'REJECTED');
CREATE TYPE decision_status   AS ENUM ('PENDING', 'APPROVED', 'REJECTED');
CREATE TYPE change_source     AS ENUM ('MANUAL', 'SYSTEM');

-- Bổ sung: truy vết nguồn gốc kỹ năng chính thức (mục 2.4.2)
CREATE TYPE skill_source      AS ENUM ('MANUAL', 'CV');

-- -----------------------------------------------------------------------------
-- 2. Tài khoản và nhân sự
-- -----------------------------------------------------------------------------

CREATE TABLE users (
    user_id       int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username      varchar(50)      NOT NULL UNIQUE,
    password_hash varchar(255)     NOT NULL,
    system_role   user_system_role NOT NULL,
    status        user_status      NOT NULL DEFAULT 'ACTIVE',
    created_at    timestamptz      NOT NULL DEFAULT now(),
    updated_at    timestamptz      NOT NULL DEFAULT now()
);

CREATE TABLE departments (
    department_id   int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    department_name varchar(100) NOT NULL UNIQUE,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE positions (
    position_id   int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    position_name varchar(100) NOT NULL UNIQUE,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE employees (
    employee_id       int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id           int               NOT NULL UNIQUE REFERENCES users (user_id),
    employee_code     varchar(30)       NOT NULL UNIQUE,
    fullname          varchar(100)      NOT NULL,
    email             varchar(100)      NOT NULL UNIQUE,
    phone             varchar(20),
    department_id     int               REFERENCES departments (department_id),
    position_id       int               NOT NULL REFERENCES positions (position_id),
    seniority_level   varchar(30),
    employment_status employment_status NOT NULL DEFAULT 'ACTIVE',
    created_at        timestamptz       NOT NULL DEFAULT now(),
    updated_at        timestamptz       NOT NULL DEFAULT now()
);

CREATE INDEX idx_employees_department ON employees (department_id);
CREATE INDEX idx_employees_position   ON employees (position_id);
CREATE INDEX idx_employees_status     ON employees (employment_status);

-- -----------------------------------------------------------------------------
-- 3. Hồ sơ năng lực
-- -----------------------------------------------------------------------------

CREATE TABLE skill_categories (
    category_id   int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_name varchar(100) NOT NULL UNIQUE,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE skills (
    skill_id    int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    skill_name  varchar(100) NOT NULL UNIQUE,
    category_id int          NOT NULL REFERENCES skill_categories (category_id),
    status      skill_status NOT NULL DEFAULT 'ACTIVE',
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_skills_category ON skills (category_id);
CREATE INDEX idx_skills_status   ON skills (status);

-- Nguồn dữ liệu chính thức cho tra cứu và engine đề xuất.
-- source / approved_by / approved_at là phần bổ sung so với tài liệu, để chứng minh
-- luận điểm "chỉ dùng hồ sơ đã được xác nhận" (mục 2.4.2, UC003.1 đầu vào #3).
CREATE TABLE employee_skills (
    employee_skill_id     int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employee_id           int               NOT NULL REFERENCES employees (employee_id) ON DELETE CASCADE,
    skill_id              int               NOT NULL REFERENCES skills (skill_id),
    proficiency_level     proficiency_level NOT NULL,
    years_experience      decimal(4, 1) CHECK (years_experience IS NULL OR years_experience >= 0),
    note                  text,
    source                skill_source      NOT NULL DEFAULT 'MANUAL',
    approved_by_employee_id int             REFERENCES employees (employee_id),
    approved_at           timestamptz,
    created_at            timestamptz       NOT NULL DEFAULT now(),
    updated_at            timestamptz       NOT NULL DEFAULT now(),
    CONSTRAINT uq_employee_skills UNIQUE (employee_id, skill_id)
);

CREATE INDEX idx_employee_skills_skill ON employee_skills (skill_id);

-- -----------------------------------------------------------------------------
-- 4. Dự án và công việc
-- -----------------------------------------------------------------------------

CREATE TABLE projects (
    project_id       int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_code     varchar(30)    NOT NULL UNIQUE,
    project_name     varchar(150)   NOT NULL,
    description      text,
    start_date       date           NOT NULL,
    planned_end_date date           NOT NULL,
    priority         priority_level NOT NULL DEFAULT 'MEDIUM',
    status           project_status NOT NULL DEFAULT 'PLANNING',
    pm_employee_id   int            NOT NULL REFERENCES employees (employee_id),
    created_at       timestamptz    NOT NULL DEFAULT now(),
    updated_at       timestamptz    NOT NULL DEFAULT now(),
    -- UC001 dữ liệu đầu vào #5: ngày kết thúc không nhỏ hơn ngày bắt đầu
    CONSTRAINT ck_projects_dates CHECK (planned_end_date >= start_date)
);

CREATE INDEX idx_projects_pm     ON projects (pm_employee_id);
CREATE INDEX idx_projects_status ON projects (status);

CREATE TABLE project_members (
    project_member_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id        int         NOT NULL REFERENCES projects (project_id) ON DELETE CASCADE,
    employee_id       int         NOT NULL REFERENCES employees (employee_id),
    join_source       join_source NOT NULL,
    joined_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_project_members UNIQUE (project_id, employee_id)
);

CREATE INDEX idx_project_members_employee ON project_members (employee_id);

-- Bổ sung so với Hình 4.6 nhưng ĐÃ CÓ trong Hình 4.5 (ERD).
-- Nhóm công việc chỉ để gom nhóm, không bao giờ là đơn vị phân công.
CREATE TABLE task_groups (
    task_group_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id    int          NOT NULL REFERENCES projects (project_id) ON DELETE CASCADE,
    group_name    varchar(150) NOT NULL,
    description   text,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now(),
    -- UC002 dữ liệu đầu vào #2: không trùng tên nhóm trong cùng dự án
    CONSTRAINT uq_task_groups_name UNIQUE (project_id, group_name)
);

CREATE TABLE tasks (
    task_id             int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id          int            NOT NULL REFERENCES projects (project_id) ON DELETE CASCADE,
    task_group_id       int            REFERENCES task_groups (task_group_id) ON DELETE SET NULL,
    parent_task_id      int            REFERENCES tasks (task_id) ON DELETE SET NULL,
    current_assignee_id int            REFERENCES employees (employee_id),
    task_code           varchar(30)    NOT NULL,
    task_name           varchar(150)   NOT NULL,
    description         text,
    task_type           varchar(30)    NOT NULL DEFAULT 'CHILD',
    estimate_hours      decimal(10, 2) NOT NULL,
    actual_hours        decimal(10, 2) NOT NULL DEFAULT 0,
    planned_start_date  date,
    planned_end_date    date,
    priority            priority_level NOT NULL DEFAULT 'MEDIUM',
    status              task_status    NOT NULL DEFAULT 'PENDING_ASSIGNMENT',
    created_at          timestamptz    NOT NULL DEFAULT now(),
    updated_at          timestamptz    NOT NULL DEFAULT now(),
    -- Sửa so với tài liệu: mã công việc chỉ cần duy nhất trong phạm vi một dự án
    CONSTRAINT uq_tasks_code       UNIQUE (project_id, task_code),
    -- UC002 alt 13e: thời gian ước tính phải lớn hơn 0
    CONSTRAINT ck_tasks_estimate   CHECK (estimate_hours > 0),
    CONSTRAINT ck_tasks_actual     CHECK (actual_hours >= 0),
    -- UC002 alt 13c
    CONSTRAINT ck_tasks_dates      CHECK (
        planned_start_date IS NULL OR planned_end_date IS NULL
        OR planned_end_date >= planned_start_date
    ),
    CONSTRAINT ck_tasks_not_self_parent CHECK (parent_task_id IS DISTINCT FROM task_id)
);

CREATE INDEX idx_tasks_project  ON tasks (project_id);
CREATE INDEX idx_tasks_group    ON tasks (task_group_id);
CREATE INDEX idx_tasks_parent   ON tasks (parent_task_id);
CREATE INDEX idx_tasks_assignee ON tasks (current_assignee_id);
CREATE INDEX idx_tasks_status   ON tasks (status);

CREATE TABLE task_requirements (
    task_req_id    int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id        int               NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    skill_id       int               NOT NULL REFERENCES skills (skill_id),
    required_level proficiency_level NOT NULL,
    weight         decimal(5, 2)     NOT NULL DEFAULT 1,
    is_mandatory   boolean           NOT NULL DEFAULT false,
    created_at     timestamptz       NOT NULL DEFAULT now(),
    updated_at     timestamptz       NOT NULL DEFAULT now(),
    CONSTRAINT uq_task_requirements UNIQUE (task_id, skill_id),
    CONSTRAINT ck_task_req_weight   CHECK (weight > 0)
);

CREATE INDEX idx_task_requirements_skill ON task_requirements (skill_id);

-- -----------------------------------------------------------------------------
-- 5. Đề xuất nhân sự (bổ sung — UC003.1 dữ liệu đầu ra, mục 2.4.3)
-- -----------------------------------------------------------------------------

-- Mỗi lần PM bấm "Phân công bằng AI đề xuất" sinh một run cho mỗi task.
-- params lưu bộ trọng số để tái lập được kết quả về sau.
CREATE TABLE recommendation_runs (
    run_id                  int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id                 int         NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    engine_version          varchar(30) NOT NULL,
    params                  jsonb       NOT NULL DEFAULT '{}'::jsonb,
    generated_by_employee_id int        NOT NULL REFERENCES employees (employee_id),
    generated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_recommendation_runs_task ON recommendation_runs (task_id, generated_at DESC);

CREATE TABLE recommendation_items (
    item_id          int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    run_id           int           NOT NULL REFERENCES recommendation_runs (run_id) ON DELETE CASCADE,
    employee_id      int           NOT NULL REFERENCES employees (employee_id),
    rank_no          int           NOT NULL,
    matching_score   decimal(5, 2) NOT NULL,
    skill_score      decimal(5, 2) NOT NULL,
    workload_score   decimal(5, 2) NOT NULL,
    perf_score       decimal(5, 2),
    matched_skills   jsonb         NOT NULL DEFAULT '[]'::jsonb,
    missing_skills   jsonb         NOT NULL DEFAULT '[]'::jsonb,
    current_workload decimal(10, 2) NOT NULL DEFAULT 0,
    explanation      text          NOT NULL,
    is_selected      boolean       NOT NULL DEFAULT false,
    CONSTRAINT uq_recommendation_items UNIQUE (run_id, employee_id),
    CONSTRAINT ck_recommendation_rank  CHECK (rank_no > 0)
);

CREATE INDEX idx_recommendation_items_run ON recommendation_items (run_id, rank_no);

-- -----------------------------------------------------------------------------
-- 6. Phân công và theo dõi thực hiện
-- -----------------------------------------------------------------------------

CREATE TABLE assignments (
    assignment_id          int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id                int             NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    employee_id            int             NOT NULL REFERENCES employees (employee_id),
    assigned_by_employee_id int            NOT NULL REFERENCES employees (employee_id),
    assignment_type        assignment_type NOT NULL,
    matching_score         decimal(5, 2),
    note                   text,
    recommendation_item_id int             REFERENCES recommendation_items (item_id) ON DELETE SET NULL,
    assigned_at            timestamptz     NOT NULL DEFAULT now()
);

CREATE INDEX idx_assignments_task     ON assignments (task_id, assigned_at DESC);
CREATE INDEX idx_assignments_employee ON assignments (employee_id);

CREATE TABLE work_logs (
    timelog_id  int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id     int            NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    employee_id int            NOT NULL REFERENCES employees (employee_id),
    work_date   date           NOT NULL,
    hours_spent decimal(10, 2) NOT NULL,
    work_note   text,
    -- Nguồn ghi nhận: UC004 quy định hệ thống tự tính, không cho nhân viên nhập tùy ý
    source      change_source  NOT NULL DEFAULT 'SYSTEM',
    created_at  timestamptz    NOT NULL DEFAULT now(),
    updated_at  timestamptz    NOT NULL DEFAULT now(),
    CONSTRAINT ck_work_logs_hours CHECK (hours_spent >= 0)
);

CREATE INDEX idx_work_logs_task     ON work_logs (task_id);
CREATE INDEX idx_work_logs_employee ON work_logs (employee_id, work_date);

CREATE TABLE task_comments (
    comment_id  int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id     int         NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    employee_id int         NOT NULL REFERENCES employees (employee_id),
    content     text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_comments_task ON task_comments (task_id, created_at);

-- Audit trail. old_status cho phép NULL ở bản ghi đầu tiên (task vừa được tạo).
CREATE TABLE task_status_history (
    history_id              int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id                 int           NOT NULL REFERENCES tasks (task_id) ON DELETE CASCADE,
    changed_by_employee_id  int           REFERENCES employees (employee_id),
    old_status              task_status,
    new_status              task_status   NOT NULL,
    change_source           change_source NOT NULL,
    note                    text,
    changed_at              timestamptz   NOT NULL DEFAULT now()
);

-- Truy vấn nóng của UC004: lấy mốc thời điểm task chuyển sang IN_PROGRESS
CREATE INDEX idx_task_status_history_task ON task_status_history (task_id, changed_at);

-- -----------------------------------------------------------------------------
-- 7. Thông báo (bổ sung — UC003 bước 15)
-- -----------------------------------------------------------------------------

CREATE TABLE notifications (
    notification_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employee_id     int          NOT NULL REFERENCES employees (employee_id) ON DELETE CASCADE,
    type            varchar(50)  NOT NULL,
    title           varchar(200) NOT NULL,
    content         text,
    ref_task_id     int          REFERENCES tasks (task_id) ON DELETE CASCADE,
    is_read         boolean      NOT NULL DEFAULT false,
    created_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_employee ON notifications (employee_id, is_read, created_at DESC);

-- -----------------------------------------------------------------------------
-- 8. CV và trích xuất kỹ năng
-- -----------------------------------------------------------------------------

CREATE TABLE employee_cvs (
    cv_id              int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employee_id        int          NOT NULL REFERENCES employees (employee_id) ON DELETE CASCADE,
    storage_provider   varchar(30)  NOT NULL,
    storage_key        varchar(255) NOT NULL,
    file_url           text         NOT NULL,
    original_file_name varchar(255),
    mime_type          varchar(100),
    uploaded_at        timestamptz  NOT NULL DEFAULT now(),
    status             cv_status    NOT NULL DEFAULT 'ACTIVE'
);

CREATE INDEX idx_employee_cvs_employee ON employee_cvs (employee_id, status);

CREATE TABLE cv_extractions (
    extraction_id           int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cv_id                   int               NOT NULL REFERENCES employee_cvs (cv_id) ON DELETE CASCADE,
    extracted_text          text,
    extracted_markdown      text,
    extraction_status       extraction_status NOT NULL DEFAULT 'PENDING_REVIEW',
    extracted_at            timestamptz       NOT NULL DEFAULT now(),
    reviewed_at             timestamptz,
    reviewed_by_employee_id int               REFERENCES employees (employee_id),
    review_note             text
);

CREATE INDEX idx_cv_extractions_cv     ON cv_extractions (cv_id);
CREATE INDEX idx_cv_extractions_status ON cv_extractions (extraction_status);

-- skill_id NULL khi kỹ năng đọc được từ CV chưa khớp với danh mục Skills.
-- Quản trị viên phải map thủ công trước khi duyệt.
CREATE TABLE cv_extracted_skills (
    extracted_skill_id       int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    extraction_id            int             NOT NULL REFERENCES cv_extractions (extraction_id) ON DELETE CASCADE,
    skill_id                 int             REFERENCES skills (skill_id),
    raw_skill_name           varchar(100)    NOT NULL,
    suggested_level          proficiency_level,
    suggested_years_experience decimal(4, 1),
    confidence_score         decimal(5, 2),
    decision_status          decision_status NOT NULL DEFAULT 'PENDING',
    evidence_text            text
);

CREATE INDEX idx_cv_extracted_skills_extraction ON cv_extracted_skills (extraction_id, decision_status);

-- -----------------------------------------------------------------------------
-- 9. Trigger tự cập nhật updated_at
-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'users', 'departments', 'positions', 'employees',
        'skill_categories', 'skills', 'employee_skills',
        'projects', 'task_groups', 'tasks', 'task_requirements',
        'work_logs', 'task_comments'
    ] LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%1$s_updated_at BEFORE UPDATE ON %1$s
             FOR EACH ROW EXECUTE FUNCTION set_updated_at()', t
        );
    END LOOP;
END;
$$;
