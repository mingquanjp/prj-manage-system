-- Hồ sơ năng lực chính thức của nhân sự mẫu.
--
-- Tất cả bản ghi đều đánh dấu đã được Quản trị viên (EMP001) xác nhận, vì engine
-- đề xuất chỉ được phép dùng dữ liệu đã duyệt (UC003.1 dữ liệu đầu vào #3).

INSERT INTO employee_skills (
    employee_id, skill_id, proficiency_level, years_experience,
    source, approved_by_employee_id, approved_at
)
SELECT
    e.employee_id,
    s.skill_id,
    v.level::proficiency_level,
    v.years,
    'MANUAL'::skill_source,
    admin.employee_id,
    now()
FROM (VALUES
    -- Backend Developer — Senior
    ('EMP004', 'Java',                    'EXPERT',       6.0),
    ('EMP004', 'Spring Boot',             'EXPERT',       5.5),
    ('EMP004', 'SQL',                     'ADVANCED',     6.0),
    ('EMP004', 'PostgreSQL',              'ADVANCED',     4.0),
    ('EMP004', 'Thiết kế REST API',       'EXPERT',       5.0),
    ('EMP004', 'Kiến trúc Microservices', 'ADVANCED',     3.0),
    ('EMP004', 'Docker',                  'INTERMEDIATE', 2.5),

    -- Backend Developer — Middle
    ('EMP005', 'Go',                      'ADVANCED',     3.5),
    ('EMP005', 'PostgreSQL',              'ADVANCED',     3.0),
    ('EMP005', 'SQL',                     'ADVANCED',     3.5),
    ('EMP005', 'Thiết kế REST API',       'ADVANCED',     3.0),
    ('EMP005', 'Redis',                   'INTERMEDIATE', 1.5),
    ('EMP005', 'Docker',                  'INTERMEDIATE', 2.0),

    -- Backend Developer — Junior
    ('EMP006', 'Node.js',                 'INTERMEDIATE', 1.5),
    ('EMP006', 'NestJS',                  'BASIC',        1.0),
    ('EMP006', 'SQL',                     'INTERMEDIATE', 1.5),
    ('EMP006', 'MySQL',                   'INTERMEDIATE', 1.0),
    ('EMP006', 'Làm việc nhóm',           'ADVANCED',     1.5),

    -- Frontend Developer — Senior
    ('EMP007', 'JavaScript',              'EXPERT',       6.0),
    ('EMP007', 'TypeScript',              'EXPERT',       4.5),
    ('EMP007', 'React',                   'EXPERT',       5.0),
    ('EMP007', 'Next.js',                 'ADVANCED',     3.0),
    ('EMP007', 'HTML/CSS',                'EXPERT',       6.0),
    ('EMP007', 'Tailwind CSS',            'ADVANCED',     2.5),

    -- Frontend Developer — Middle
    ('EMP008', 'JavaScript',              'ADVANCED',     3.0),
    ('EMP008', 'TypeScript',              'INTERMEDIATE', 2.0),
    ('EMP008', 'React',                   'ADVANCED',     2.5),
    ('EMP008', 'HTML/CSS',                'ADVANCED',     3.0),
    ('EMP008', 'Tailwind CSS',            'INTERMEDIATE', 1.5),

    -- Frontend Developer — Junior
    ('EMP009', 'JavaScript',              'INTERMEDIATE', 1.0),
    ('EMP009', 'HTML/CSS',                'ADVANCED',     1.5),
    ('EMP009', 'Vue.js',                  'INTERMEDIATE', 1.0),
    ('EMP009', 'React',                   'BASIC',        0.5),

    -- Fullstack — Senior
    ('EMP010', 'TypeScript',              'EXPERT',       5.0),
    ('EMP010', 'React',                   'ADVANCED',     4.0),
    ('EMP010', 'Node.js',                 'EXPERT',       5.0),
    ('EMP010', 'NestJS',                  'ADVANCED',     3.0),
    ('EMP010', 'PostgreSQL',              'ADVANCED',     4.0),
    ('EMP010', 'Thiết kế REST API',       'EXPERT',       4.5),
    ('EMP010', 'Docker',                  'ADVANCED',     3.0),

    -- Fullstack — Middle
    ('EMP011', 'JavaScript',              'ADVANCED',     3.0),
    ('EMP011', 'React',                   'INTERMEDIATE', 2.0),
    ('EMP011', 'Python',                  'ADVANCED',     3.0),
    ('EMP011', 'Django',                  'INTERMEDIATE', 2.0),
    ('EMP011', 'SQL',                     'INTERMEDIATE', 2.5),

    -- Mobile — Middle
    ('EMP012', 'Flutter',                 'ADVANCED',     3.0),
    ('EMP012', 'Android',                 'INTERMEDIATE', 2.0),
    ('EMP012', 'Thiết kế REST API',       'INTERMEDIATE', 2.0),
    ('EMP012', 'Giải quyết vấn đề',       'ADVANCED',     3.0),

    -- Mobile — Junior
    ('EMP013', 'React Native',            'INTERMEDIATE', 1.0),
    ('EMP013', 'JavaScript',              'INTERMEDIATE', 1.5),
    ('EMP013', 'iOS',                     'BASIC',        0.5),

    -- QA — Senior
    ('EMP014', 'Kiểm thử thủ công',       'EXPERT',       6.0),
    ('EMP014', 'Kiểm thử tự động',        'ADVANCED',     4.0),
    ('EMP014', 'Selenium',                'ADVANCED',     4.0),
    ('EMP014', 'Kiểm thử API',            'EXPERT',       4.5),
    ('EMP014', 'Kiểm thử hiệu năng',      'INTERMEDIATE', 2.0),
    ('EMP014', 'SQL',                     'INTERMEDIATE', 3.0),

    -- QA — Middle
    ('EMP015', 'Kiểm thử thủ công',       'ADVANCED',     2.5),
    ('EMP015', 'Kiểm thử API',            'INTERMEDIATE', 2.0),
    ('EMP015', 'Selenium',                'INTERMEDIATE', 1.5),
    ('EMP015', 'Làm việc nhóm',           'ADVANCED',     2.5),

    -- DevOps — Senior
    ('EMP016', 'Docker',                  'EXPERT',       5.0),
    ('EMP016', 'Kubernetes',              'ADVANCED',     3.5),
    ('EMP016', 'CI/CD',                   'EXPERT',       5.0),
    ('EMP016', 'AWS',                     'ADVANCED',     4.0),
    ('EMP016', 'Linux',                   'EXPERT',       6.0),
    ('EMP016', 'Nginx',                   'ADVANCED',     4.0),

    -- DevOps — Junior
    ('EMP017', 'Docker',                  'INTERMEDIATE', 1.5),
    ('EMP017', 'Linux',                   'INTERMEDIATE', 2.0),
    ('EMP017', 'CI/CD',                   'BASIC',        1.0),

    -- Business Analyst — Senior
    ('EMP018', 'Phân tích yêu cầu',       'EXPERT',       6.0),
    ('EMP018', 'UML',                     'ADVANCED',     5.0),
    ('EMP018', 'Viết tài liệu đặc tả',    'EXPERT',       6.0),
    ('EMP018', 'Agile/Scrum',             'ADVANCED',     4.0),
    ('EMP018', 'Giao tiếp',               'EXPERT',       6.0),
    ('EMP018', 'SQL',                     'INTERMEDIATE', 2.0),

    -- Business Analyst — Junior
    ('EMP019', 'Phân tích yêu cầu',       'INTERMEDIATE', 1.0),
    ('EMP019', 'UML',                     'INTERMEDIATE', 1.0),
    ('EMP019', 'Viết tài liệu đặc tả',    'INTERMEDIATE', 1.5),
    ('EMP019', 'Giao tiếp',               'ADVANCED',     1.5),

    -- Technical Lead — Senior
    ('EMP020', 'Java',                    'EXPERT',       8.0),
    ('EMP020', 'Kiến trúc Microservices', 'EXPERT',       5.0),
    ('EMP020', 'Thiết kế cơ sở dữ liệu',  'EXPERT',       7.0),
    ('EMP020', 'PostgreSQL',              'EXPERT',       6.0),
    ('EMP020', 'Docker',                  'ADVANCED',     4.0),
    ('EMP020', 'Agile/Scrum',             'ADVANCED',     5.0),
    ('EMP020', 'Giải quyết vấn đề',       'EXPERT',       8.0),

    -- Hai Quản lý dự án cũng có hồ sơ năng lực để tra cứu
    ('EMP002', 'Agile/Scrum',             'EXPERT',       6.0),
    ('EMP002', 'Phân tích yêu cầu',       'ADVANCED',     5.0),
    ('EMP002', 'Giao tiếp',               'EXPERT',       7.0),
    ('EMP002', 'Quản lý thời gian',       'EXPERT',       6.0),

    ('EMP003', 'Agile/Scrum',             'ADVANCED',     4.0),
    ('EMP003', 'Java',                    'ADVANCED',     5.0),
    ('EMP003', 'Quản lý thời gian',       'ADVANCED',     4.0),
    ('EMP003', 'Giao tiếp',               'ADVANCED',     5.0)
) AS v(employee_code, skill_name, level, years)
JOIN employees e   ON e.employee_code = v.employee_code
JOIN skills    s   ON s.skill_name    = v.skill_name
CROSS JOIN LATERAL (SELECT employee_id FROM employees WHERE employee_code = 'EMP001') AS admin
ON CONFLICT ON CONSTRAINT uq_employee_skills DO NOTHING;
