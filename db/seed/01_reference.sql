-- Dữ liệu danh mục nền. Chạy lại nhiều lần không sinh trùng.

INSERT INTO departments (department_name) VALUES
    ('Phòng Phát triển phần mềm'),
    ('Phòng Kiểm thử'),
    ('Phòng Hạ tầng'),
    ('Phòng Sản phẩm'),
    ('Phòng Hành chính - Nhân sự')
ON CONFLICT (department_name) DO NOTHING;

INSERT INTO positions (position_name) VALUES
    ('Quản trị hệ thống'),
    ('Project Manager'),
    ('Technical Lead'),
    ('Business Analyst'),
    ('Backend Developer'),
    ('Frontend Developer'),
    ('Fullstack Developer'),
    ('Mobile Developer'),
    ('QA Engineer'),
    ('DevOps Engineer')
ON CONFLICT (position_name) DO NOTHING;

INSERT INTO skill_categories (category_name) VALUES
    ('Backend'),
    ('Frontend'),
    ('Cơ sở dữ liệu'),
    ('DevOps'),
    ('Kiểm thử'),
    ('Di động'),
    ('Phân tích nghiệp vụ'),
    ('Kỹ năng mềm')
ON CONFLICT (category_name) DO NOTHING;

INSERT INTO skills (skill_name, category_id)
SELECT v.skill_name, c.category_id
FROM (VALUES
    ('Java',                    'Backend'),
    ('Spring Boot',             'Backend'),
    ('Go',                      'Backend'),
    ('Node.js',                 'Backend'),
    ('NestJS',                  'Backend'),
    ('Python',                  'Backend'),
    ('Django',                  'Backend'),
    ('.NET Core',               'Backend'),
    ('PHP',                     'Backend'),
    ('Laravel',                 'Backend'),
    ('Thiết kế REST API',       'Backend'),
    ('GraphQL',                 'Backend'),
    ('Kiến trúc Microservices', 'Backend'),

    ('HTML/CSS',                'Frontend'),
    ('JavaScript',              'Frontend'),
    ('TypeScript',              'Frontend'),
    ('React',                   'Frontend'),
    ('Next.js',                 'Frontend'),
    ('Vue.js',                  'Frontend'),
    ('Angular',                 'Frontend'),
    ('Tailwind CSS',            'Frontend'),

    ('SQL',                     'Cơ sở dữ liệu'),
    ('PostgreSQL',              'Cơ sở dữ liệu'),
    ('MySQL',                   'Cơ sở dữ liệu'),
    ('SQL Server',              'Cơ sở dữ liệu'),
    ('MongoDB',                 'Cơ sở dữ liệu'),
    ('Redis',                   'Cơ sở dữ liệu'),
    ('Thiết kế cơ sở dữ liệu',  'Cơ sở dữ liệu'),

    ('Docker',                  'DevOps'),
    ('Kubernetes',              'DevOps'),
    ('CI/CD',                   'DevOps'),
    ('AWS',                     'DevOps'),
    ('Nginx',                   'DevOps'),
    ('Linux',                   'DevOps'),

    ('Kiểm thử thủ công',       'Kiểm thử'),
    ('Kiểm thử API',            'Kiểm thử'),
    ('Kiểm thử tự động',        'Kiểm thử'),
    ('Selenium',                'Kiểm thử'),
    ('Kiểm thử hiệu năng',      'Kiểm thử'),

    ('Flutter',                 'Di động'),
    ('React Native',            'Di động'),
    ('Android',                 'Di động'),
    ('iOS',                     'Di động'),

    ('Phân tích yêu cầu',       'Phân tích nghiệp vụ'),
    ('UML',                     'Phân tích nghiệp vụ'),
    ('Viết tài liệu đặc tả',    'Phân tích nghiệp vụ'),
    ('Agile/Scrum',             'Phân tích nghiệp vụ'),

    ('Giao tiếp',               'Kỹ năng mềm'),
    ('Làm việc nhóm',           'Kỹ năng mềm'),
    ('Quản lý thời gian',       'Kỹ năng mềm'),
    ('Giải quyết vấn đề',       'Kỹ năng mềm')
) AS v(skill_name, category_name)
JOIN skill_categories c ON c.category_name = v.category_name
ON CONFLICT (skill_name) DO NOTHING;
